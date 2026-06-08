package usecase

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"openbridge/backend/internal/config"
	"openbridge/backend/internal/pkg/logger"

	"go.uber.org/zap"
)

const (
	fileTreeCacheVersion  = 1
	fileTreeCacheMinBytes = int64(4 * 1024)
)

type FILETREE struct {
	Version         int           `json:"version"`
	OpenListBaseURL string        `json:"openlist_base_url"`
	SavedAt         time.Time     `json:"saved_at"`
	ApproxBytes     int64         `json:"approx_bytes"`
	Root            *FileTreeNode `json:"root"`
}

type FileTreeNode struct {
	Name     string                   `json:"name"`
	Path     string                   `json:"path"`
	Provider string                   `json:"provider,omitempty"`
	Total    int                      `json:"total"`
	Readme   string                   `json:"readme,omitempty"`
	Header   string                   `json:"header,omitempty"`
	Write    bool                     `json:"write"`
	Loaded   bool                     `json:"loaded"`
	CachedAt time.Time                `json:"cached_at"`
	Content  []FileItem               `json:"content,omitempty"`
	Children map[string]*FileTreeNode `json:"children,omitempty"`
}

type fileTreePruneCandidate struct {
	Parent   *FileTreeNode
	Name     string
	Depth    int
	CachedAt time.Time
}

type FileTreeCache struct {
	mu            sync.RWMutex
	filePath      string
	sourceBaseURL string
	maxBytes      int64
	maxDepth      int
	tree          FILETREE
	dirty         bool
}

func NewFileTreeCache(cfg *config.Config) *FileTreeCache {
	cachePath := filepath.Join("data", "filetree_cache.json")
	if cfg != nil && strings.TrimSpace(cfg.DB.Path) != "" {
		cachePath = filepath.Join(filepath.Dir(cfg.DB.Path), "filetree_cache.json")
	}

	maxBytes := int64(1024 * 1024)
	maxDepth := 2
	sourceBaseURL := ""
	if cfg != nil {
		if cfg.FileTree.MaxBytes >= fileTreeCacheMinBytes {
			maxBytes = cfg.FileTree.MaxBytes
		}
		if cfg.FileTree.MaxDepth >= 1 && cfg.FileTree.MaxDepth <= 5 {
			maxDepth = cfg.FileTree.MaxDepth
		}
		sourceBaseURL = config.NormalizeBaseURLScope(cfg.OpenList.BaseURL)
	}

	return &FileTreeCache{
		filePath:      cachePath,
		sourceBaseURL: sourceBaseURL,
		maxBytes:      maxBytes,
		maxDepth:      maxDepth,
		tree: FILETREE{
			Version:         fileTreeCacheVersion,
			OpenListBaseURL: sourceBaseURL,
			Root:            newFileTreeRootNode(),
		},
	}
}

func (c *FileTreeCache) Enabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.maxBytes >= fileTreeCacheMinBytes && c.maxDepth >= 1
}

func (c *FileTreeCache) MaxDepth() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.maxDepth
}

func (c *FileTreeCache) Load() error {
	payload, err := os.ReadFile(c.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var persisted FILETREE
	if err := json.Unmarshal(payload, &persisted); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if config.NormalizeBaseURLScope(persisted.OpenListBaseURL) != c.sourceBaseURL {
		c.resetLocked()
		return nil
	}

	if persisted.Root == nil {
		persisted.Root = newFileTreeRootNode()
	}

	c.tree = persisted
	c.tree.Version = fileTreeCacheVersion
	c.pruneForLimitsLocked()
	c.dirty = false
	return nil
}

func (c *FileTreeCache) SwitchSource(sourceBaseURL string) error {
	c.mu.Lock()
	c.sourceBaseURL = config.NormalizeBaseURLScope(sourceBaseURL)
	c.resetLocked()
	c.mu.Unlock()

	return c.Load()
}

func (c *FileTreeCache) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.pruneForLimitsLocked()
	c.tree.Version = fileTreeCacheVersion
	c.tree.OpenListBaseURL = c.sourceBaseURL
	c.tree.SavedAt = time.Now()
	c.tree.ApproxBytes = c.estimateLocked()

	payload, err := json.MarshalIndent(c.tree, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(c.filePath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(c.filePath, payload, 0644); err != nil {
		return err
	}

	c.dirty = false
	return nil
}

func (c *FileTreeCache) UpdateConfig(maxBytes int64, maxDepth int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.maxBytes = maxBytes
	c.maxDepth = maxDepth
	c.pruneForLimitsLocked()
	c.dirty = true
	return nil
}

func (c *FileTreeCache) ShouldCache(storagePath string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.shouldCacheLocked(normalizeStoragePath(storagePath))
}

func (c *FileTreeCache) Get(storagePath string, page uint, pageSize uint) (*FileListData, bool) {
	normalized := normalizeStoragePath(storagePath)

	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.shouldCacheLocked(normalized) {
		return nil, false
	}

	node := c.lookupLocked(normalized)
	if node == nil || !c.isUsableNodeLocked(node) {
		return nil, false
	}

	return paginateFileListData(&FileListData{
		Content:  cloneFileItems(node.Content),
		Total:    node.Total,
		Readme:   node.Readme,
		Header:   node.Header,
		Write:    node.Write,
		Provider: node.Provider,
	}, page, pageSize), true
}

func (c *FileTreeCache) Put(storagePath string, data *FileListData) error {
	if data == nil {
		return nil
	}

	normalized := normalizeStoragePath(storagePath)

	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.shouldCacheLocked(normalized) {
		return nil
	}

	node := c.ensureNodeLocked(normalized)
	node.Provider = data.Provider
	node.Total = data.Total
	if node.Total <= 0 {
		node.Total = len(data.Content)
	}
	node.Readme = data.Readme
	node.Header = data.Header
	node.Write = data.Write
	node.Loaded = true
	node.CachedAt = time.Now()
	node.Content = cloneFileItems(data.Content)

	existingChildren := node.Children
	nextChildren := make(map[string]*FileTreeNode)
	for _, item := range data.Content {
		if !item.IsDir && item.Type != 1 {
			continue
		}

		childPath := joinStoragePath(normalized, item.Name)
		child := existingChildren[item.Name]
		if child == nil {
			child = &FileTreeNode{
				Name: item.Name,
				Path: childPath,
			}
		}
		child.Name = item.Name
		child.Path = childPath
		nextChildren[item.Name] = child
	}
	node.Children = nextChildren

	c.pruneForLimitsLocked()
	c.dirty = true
	return nil
}

func (c *FileTreeCache) resetLocked() {
	c.tree = FILETREE{
		Version:         fileTreeCacheVersion,
		OpenListBaseURL: c.sourceBaseURL,
		Root:            newFileTreeRootNode(),
	}
	c.dirty = false
}

func (c *FileTreeCache) shouldCacheLocked(storagePath string) bool {
	return c.maxDepth >= 1 && c.maxBytes >= fileTreeCacheMinBytes && storagePathDepth(storagePath) <= c.maxDepth
}

func (c *FileTreeCache) isUsableNodeLocked(node *FileTreeNode) bool {
	if node == nil || !node.Loaded {
		return false
	}
	if !node.CachedAt.IsZero() && (node.Provider != "" || node.Total > 0 || len(node.Content) > 0 || node.Readme != "" || node.Header != "" || node.Write) {
		return true
	}
	return false
}

func (c *FileTreeCache) lookupLocked(storagePath string) *FileTreeNode {
	if c.tree.Root == nil {
		return nil
	}
	if storagePath == "/" {
		return c.tree.Root
	}

	current := c.tree.Root
	for _, segment := range splitStoragePath(storagePath) {
		if current.Children == nil {
			return nil
		}
		next := current.Children[segment]
		if next == nil {
			return nil
		}
		current = next
	}

	return current
}

func (c *FileTreeCache) ensureNodeLocked(storagePath string) *FileTreeNode {
	if c.tree.Root == nil {
		c.tree.Root = newFileTreeRootNode()
	}
	if storagePath == "/" {
		return c.tree.Root
	}

	current := c.tree.Root
	currentPath := ""
	for _, segment := range splitStoragePath(storagePath) {
		currentPath = joinStoragePath(currentPath, segment)
		if current.Children == nil {
			current.Children = make(map[string]*FileTreeNode)
		}
		next := current.Children[segment]
		if next == nil {
			next = &FileTreeNode{
				Name: segment,
				Path: currentPath,
			}
			current.Children[segment] = next
		}
		current = next
	}

	return current
}

func (c *FileTreeCache) pruneForLimitsLocked() {
	if c.tree.Root == nil {
		c.tree.Root = newFileTreeRootNode()
	}

	c.pruneByDepthLocked(c.tree.Root, 0)

	for c.estimateLocked() > c.maxBytes {
		candidates := make([]fileTreePruneCandidate, 0)
		c.collectPruneCandidatesLocked(c.tree.Root, 0, &candidates)
		if len(candidates) == 0 {
			c.tree.Root.Content = nil
			c.tree.Root.Children = nil
			c.tree.Root.Loaded = false
			c.tree.Root.Total = 0
			c.tree.Root.Provider = ""
			c.tree.Root.Readme = ""
			c.tree.Root.Header = ""
			c.tree.Root.Write = false
			break
		}

		sort.Slice(candidates, func(i, j int) bool {
			if candidates[i].Depth != candidates[j].Depth {
				return candidates[i].Depth > candidates[j].Depth
			}
			return candidates[i].CachedAt.Before(candidates[j].CachedAt)
		})

		delete(candidates[0].Parent.Children, candidates[0].Name)
	}

	c.tree.ApproxBytes = c.estimateLocked()
}

func (c *FileTreeCache) pruneByDepthLocked(node *FileTreeNode, depth int) {
	if node == nil {
		return
	}
	if depth >= c.maxDepth {
		node.Children = nil
		return
	}
	for _, child := range node.Children {
		c.pruneByDepthLocked(child, depth+1)
	}
}

func (c *FileTreeCache) collectPruneCandidatesLocked(node *FileTreeNode, depth int, candidates *[]fileTreePruneCandidate) {
	if node == nil || node.Children == nil {
		return
	}

	for name, child := range node.Children {
		*candidates = append(*candidates, fileTreePruneCandidate{
			Parent:   node,
			Name:     name,
			Depth:    depth + 1,
			CachedAt: child.CachedAt,
		})
		c.collectPruneCandidatesLocked(child, depth+1, candidates)
	}
}

func (c *FileTreeCache) estimateLocked() int64 {
	c.tree.OpenListBaseURL = c.sourceBaseURL
	payload, err := json.Marshal(c.tree)
	if err != nil {
		logger.L().Warn("estimate filetree cache size failed", zap.Error(err))
		return 0
	}
	return int64(len(payload))
}

func newFileTreeRootNode() *FileTreeNode {
	return &FileTreeNode{
		Name: "/",
		Path: "/",
	}
}

func cloneFileItems(items []FileItem) []FileItem {
	if len(items) == 0 {
		return nil
	}
	cloned := make([]FileItem, len(items))
	copy(cloned, items)
	return cloned
}

func normalizeStoragePath(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "/" {
		return "/"
	}
	cleaned := path.Clean("/" + strings.TrimLeft(trimmed, "/"))
	if cleaned == "." {
		return "/"
	}
	return cleaned
}

func splitStoragePath(storagePath string) []string {
	normalized := normalizeStoragePath(storagePath)
	if normalized == "/" {
		return nil
	}
	return strings.Split(strings.Trim(normalized, "/"), "/")
}

func storagePathDepth(storagePath string) int {
	return len(splitStoragePath(storagePath))
}

func paginateFileListData(data *FileListData, page uint, pageSize uint) *FileListData {
	if data == nil {
		return nil
	}

	cloned := &FileListData{
		Content:  cloneFileItems(data.Content),
		Total:    data.Total,
		Readme:   data.Readme,
		Header:   data.Header,
		Write:    data.Write,
		Provider: data.Provider,
	}

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = uint(len(cloned.Content))
	}
	if pageSize == 0 {
		cloned.Content = []FileItem{}
		if cloned.Total == 0 {
			cloned.Total = 0
		}
		return cloned
	}

	start := int((page - 1) * pageSize)
	if start >= len(cloned.Content) {
		cloned.Content = []FileItem{}
		if cloned.Total == 0 {
			cloned.Total = len(data.Content)
		}
		return cloned
	}

	end := start + int(pageSize)
	if end > len(cloned.Content) {
		end = len(cloned.Content)
	}
	cloned.Content = cloneFileItems(cloned.Content[start:end])
	if cloned.Total == 0 {
		cloned.Total = len(data.Content)
	}
	return cloned
}
