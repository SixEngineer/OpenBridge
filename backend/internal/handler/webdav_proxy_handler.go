package handler

import (
	"errors"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"openbridge/backend/internal/config"
	"openbridge/backend/internal/usecase"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	webDAVHrefPattern           = regexp.MustCompile(`(?is)<(?:[A-Za-z0-9_-]+:)?href>.*?</(?:[A-Za-z0-9_-]+:)?href>`)
	webDAVQuotaAvailablePattern = regexp.MustCompile(`(?is)<(?:[A-Za-z0-9_-]+:)?quota-available-bytes>.*?</(?:[A-Za-z0-9_-]+:)?quota-available-bytes>`)
	webDAVQuotaUsedPattern      = regexp.MustCompile(`(?is)<(?:[A-Za-z0-9_-]+:)?quota-used-bytes>.*?</(?:[A-Za-z0-9_-]+:)?quota-used-bytes>`)
	webDAVPropClosePattern      = regexp.MustCompile(`(?i)</(?:[A-Za-z0-9_-]+:)?prop>`)
)

type WebDAVProxyHandler struct {
	mountUseCase *usecase.MountUseCase
	config       *config.Config
}

func NewWebDAVProxyHandler(mountUseCase *usecase.MountUseCase, cfg *config.Config) *WebDAVProxyHandler {
	return &WebDAVProxyHandler{
		mountUseCase: mountUseCase,
		config:       cfg,
	}
}

func (h *WebDAVProxyHandler) ProxyMount(c *gin.Context) {
	mountID, err := parseMountID(c)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	mount, err := h.mountUseCase.GetMountForWebDAV(c.Request.Context(), mountID)
	if err != nil {
		c.String(statusForMountProxyError(err), err.Error())
		return
	}

	var quota usecase.MountQuotaResult
	if strings.EqualFold(c.Request.Method, "PROPFIND") {
		quota, err = h.mountUseCase.GetMountQuotaReadonlyForMount(c.Request.Context(), mount)
		if err != nil {
			c.String(statusForMountProxyError(err), err.Error())
			return
		}
	}

	target, err := url.Parse(strings.TrimSpace(h.config.OpenList.BaseURL))
	if err != nil || target.Scheme == "" || target.Host == "" {
		c.String(http.StatusBadGateway, "openlist base url invalid")
		return
	}

	proxyPrefix := webDAVProxyPrefix(mountID)
	requestSuffix := webDAVRequestSuffix(c.Request.URL.Path, proxyPrefix)
	upstreamRoot := joinWebDAVPath(target.Path, "dav", mount.MountPath)
	upstreamPath := joinWebDAVPath(upstreamRoot, requestSuffix)
	upstreamOrigin := target.Scheme + "://" + target.Host
	proxyAbsoluteRoot := requestOrigin(c) + proxyPrefix
	upstreamAbsoluteRoot := upstreamOrigin + upstreamRoot

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = upstreamPath
			req.URL.RawPath = ""
			req.URL.RawQuery = c.Request.URL.RawQuery
			req.Host = target.Host
			req.RequestURI = ""

			if destination := req.Header.Get("Destination"); destination != "" {
				req.Header.Set("Destination", rewriteWebDAVReference(destination, proxyPrefix, upstreamRoot, proxyAbsoluteRoot, upstreamAbsoluteRoot))
			}
		},
		ModifyResponse: func(resp *http.Response) error {
			rewriteWebDAVResponseHeaders(resp.Header, upstreamRoot, proxyPrefix, upstreamAbsoluteRoot, proxyAbsoluteRoot)

			if !strings.EqualFold(c.Request.Method, "PROPFIND") {
				return nil
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			_ = resp.Body.Close()

			rewritten := rewriteWebDAVPropfindBody(body, upstreamRoot, proxyPrefix, upstreamAbsoluteRoot, proxyAbsoluteRoot, quota)
			resp.Body = io.NopCloser(strings.NewReader(rewritten))
			resp.ContentLength = int64(len(rewritten))
			resp.Header.Set("Content-Length", strconv.Itoa(len(rewritten)))
			return nil
		},
		ErrorHandler: func(rw http.ResponseWriter, req *http.Request, proxyErr error) {
			http.Error(rw, proxyErr.Error(), http.StatusBadGateway)
		},
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func statusForMountProxyError(err error) int {
	switch {
	case errors.Is(err, usecase.ErrMountDisabled):
		return http.StatusForbidden
	case errors.Is(err, gorm.ErrRecordNotFound):
		return http.StatusNotFound
	default:
		return http.StatusNotFound
	}
}

func rewriteWebDAVPropfindBody(body []byte, upstreamRoot string, proxyPrefix string, upstreamAbsoluteRoot string, proxyAbsoluteRoot string, quota usecase.MountQuotaResult) string {
	rewritten := webDAVHrefPattern.ReplaceAllStringFunc(string(body), func(tag string) string {
		start := strings.Index(tag, ">")
		end := strings.LastIndex(tag, "</")
		if start < 0 || end <= start {
			return tag
		}

		href := tag[start+1 : end]
		return tag[:start+1] + rewriteWebDAVReference(href, upstreamRoot, proxyPrefix, upstreamAbsoluteRoot, proxyAbsoluteRoot) + tag[end:]
	})

	var hasAvailable bool
	rewritten, hasAvailable = replaceWebDAVQuotaTag(rewritten, webDAVQuotaAvailablePattern, "quota-available-bytes", quotaMBToBytes(quota.Quota.Available))

	var hasUsed bool
	rewritten, hasUsed = replaceWebDAVQuotaTag(rewritten, webDAVQuotaUsedPattern, "quota-used-bytes", quotaMBToBytes(quota.Quota.Used))

	if !hasAvailable || !hasUsed {
		rewritten = injectMissingWebDAVQuotaTags(
			rewritten,
			!hasAvailable,
			!hasUsed,
			quotaMBToBytes(quota.Quota.Available),
			quotaMBToBytes(quota.Quota.Used),
		)
	}

	return rewritten
}

func replaceWebDAVQuotaTag(body string, pattern *regexp.Regexp, localName string, value int64) (string, bool) {
	if !pattern.MatchString(body) {
		return body, false
	}

	replacement := strconv.FormatInt(value, 10)
	rewritten := pattern.ReplaceAllStringFunc(body, func(tag string) string {
		prefix := xmlTagPrefix(tag, localName)
		return "<" + prefix + localName + ">" + replacement + "</" + prefix + localName + ">"
	})
	return rewritten, true
}

func injectMissingWebDAVQuotaTags(body string, injectAvailable bool, injectUsed bool, available int64, used int64) string {
	if !injectAvailable && !injectUsed {
		return body
	}

	availableValue := strconv.FormatInt(available, 10)
	usedValue := strconv.FormatInt(used, 10)

	return webDAVPropClosePattern.ReplaceAllStringFunc(body, func(closeTag string) string {
		prefix := xmlPropPrefix(closeTag)
		var builder strings.Builder
		if injectAvailable {
			builder.WriteString("<")
			builder.WriteString(prefix)
			builder.WriteString("quota-available-bytes>")
			builder.WriteString(availableValue)
			builder.WriteString("</")
			builder.WriteString(prefix)
			builder.WriteString("quota-available-bytes>")
		}
		if injectUsed {
			builder.WriteString("<")
			builder.WriteString(prefix)
			builder.WriteString("quota-used-bytes>")
			builder.WriteString(usedValue)
			builder.WriteString("</")
			builder.WriteString(prefix)
			builder.WriteString("quota-used-bytes>")
		}
		builder.WriteString(closeTag)
		return builder.String()
	})
}

func rewriteWebDAVResponseHeaders(header http.Header, upstreamRoot string, proxyPrefix string, upstreamAbsoluteRoot string, proxyAbsoluteRoot string) {
	for _, key := range []string{"Location", "Content-Location"} {
		value := header.Get(key)
		if value == "" {
			continue
		}
		header.Set(key, rewriteWebDAVReference(value, upstreamRoot, proxyPrefix, upstreamAbsoluteRoot, proxyAbsoluteRoot))
	}
}

func rewriteWebDAVReference(value string, fromPath string, toPath string, fromAbsolute string, toAbsolute string) string {
	switch {
	case fromAbsolute != "" && strings.HasPrefix(value, fromAbsolute):
		return toAbsolute + strings.TrimPrefix(value, fromAbsolute)
	case strings.HasPrefix(value, fromPath):
		return toPath + strings.TrimPrefix(value, fromPath)
	default:
		return value
	}
}

func webDAVProxyPrefix(mountID uint) string {
	return "/api/v1/webdav/mounts/" + strconv.FormatUint(uint64(mountID), 10)
}

func webDAVRequestSuffix(requestPath string, proxyPrefix string) string {
	suffix := strings.TrimPrefix(requestPath, proxyPrefix)
	if suffix == "" || suffix == "/" {
		return ""
	}
	if !strings.HasPrefix(suffix, "/") {
		return "/" + suffix
	}
	return suffix
}

func requestOrigin(c *gin.Context) string {
	if forwarded := strings.TrimSpace(c.Request.Header.Get("X-Forwarded-Proto")); forwarded != "" {
		proto := strings.TrimSpace(strings.Split(forwarded, ",")[0])
		return proto + "://" + c.Request.Host
	}
	if c.Request.TLS != nil {
		return "https://" + c.Request.Host
	}
	return "http://" + c.Request.Host
}

func joinWebDAVPath(basePath string, parts ...string) string {
	joined := basePath
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}
		trailingSlash := strings.HasSuffix(part, "/")
		if joined == "" {
			joined = "/"
		}
		joined = strings.TrimRight(joined, "/") + "/" + strings.TrimLeft(part, "/")
		joined = path.Clean(joined)
		if !strings.HasPrefix(joined, "/") {
			joined = "/" + joined
		}
		if trailingSlash && joined != "/" {
			joined += "/"
		}
	}

	if strings.TrimSpace(joined) == "" {
		return "/"
	}
	if !strings.HasPrefix(joined, "/") {
		joined = "/" + joined
	}
	return joined
}

func xmlTagPrefix(tag string, localName string) string {
	idx := strings.Index(strings.ToLower(tag), localName)
	if idx <= 1 {
		return ""
	}
	return tag[1:idx]
}

func xmlPropPrefix(closeTag string) string {
	trimmed := strings.TrimPrefix(strings.TrimSpace(closeTag), "</")
	lower := strings.ToLower(trimmed)
	index := strings.Index(lower, "prop>")
	if index <= 0 {
		return ""
	}
	return trimmed[:index]
}

func quotaMBToBytes(value int64) int64 {
	if value <= 0 {
		return 0
	}
	return value * 1024 * 1024
}
