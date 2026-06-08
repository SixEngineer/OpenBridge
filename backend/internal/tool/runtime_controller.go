package tool

import (
	"context"
	"net/http"
	"sync"
	"time"

	"openbridge/backend/internal/pkg/logger"

	"go.uber.org/zap"
)

type RuntimeAction string

const (
	RuntimeActionNone    RuntimeAction = ""
	RuntimeActionExit    RuntimeAction = "exit"
	RuntimeActionRestart RuntimeAction = "restart"
)

type RuntimeController struct {
	mu         sync.Mutex
	server     *http.Server
	action     RuntimeAction
	triggered  bool
	beforeStop []func() error
}

func NewRuntimeController() *RuntimeController {
	return &RuntimeController{}
}

func (c *RuntimeController) BindServer(server *http.Server) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.server = server
}

func (c *RuntimeController) RegisterBeforeStop(callback func() error) {
	if callback == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.beforeStop = append(c.beforeStop, callback)
}

func (c *RuntimeController) RequestExit() {
	c.request(RuntimeActionExit)
}

func (c *RuntimeController) RequestRestart() {
	c.request(RuntimeActionRestart)
}

func (c *RuntimeController) Action() RuntimeAction {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.action
}

func (c *RuntimeController) request(action RuntimeAction) {
	c.mu.Lock()
	if c.triggered {
		c.mu.Unlock()
		return
	}

	c.triggered = true
	c.action = action
	server := c.server
	callbacks := append([]func() error(nil), c.beforeStop...)
	c.mu.Unlock()

	go func() {
		for _, callback := range callbacks {
			if err := callback(); err != nil {
				logger.L().Warn("runtime stop hook failed", zap.Error(err), zap.String("action", string(action)))
			}
		}

		if server == nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
			logger.L().Warn("runtime shutdown failed", zap.Error(err), zap.String("action", string(action)))
		}
	}()
}
