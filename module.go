package local_metrics_proxy

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

const (
	// ModuleName is the name of the module.
	ModuleName = "local_metrics_proxy"
	// ModuleName is the namespace of the module.
	ModuleNamespace = "http.handlers"
)

// Interface guards
var (
	_ caddy.Provisioner           = (*LocalMetricsProxy)(nil)
	_ caddyhttp.MiddlewareHandler = (*LocalMetricsProxy)(nil)
	_ caddyfile.Unmarshaler       = (*LocalMetricsProxy)(nil)
)

func init() {
	caddy.RegisterModule(LocalMetricsProxy{})
}

// UnixBackend is the unix-socket backend
type UnixBackend struct {
	Path string `json:"path"`
}

// LocalMetricsProxy holds a configured instance of the proxy.
type LocalMetricsProxy struct {
	UdsBackend *UnixBackend `json:"uds,omitempty"`
	logger     *zap.Logger
}

// CaddyModule returns the Caddy module information.
func (LocalMetricsProxy) CaddyModule() caddy.ModuleInfo {
	id := fmt.Sprintf("%s.%s", ModuleNamespace, ModuleName)
	return caddy.ModuleInfo{
		ID:  caddy.ModuleID(id),
		New: func() caddy.Module { return new(LocalMetricsProxy) },
	}
}

// Provision sets up the module.
func (lmp *LocalMetricsProxy) Provision(ctx caddy.Context) error {
	if lmp.UdsBackend == nil || lmp.UdsBackend.Path == "" {
		return errors.New("proxy backend not configured")
	}
	lmp.logger = ctx.Logger(lmp)
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (lmp LocalMetricsProxy) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	if lmp.UdsBackend == nil || lmp.UdsBackend.Path == "" {
		return errors.New("proxy backend not configured")
	}

	dialer := net.Dialer{}
	conn, err := dialer.Dial("unix", lmp.UdsBackend.Path)
	if err != nil {
		return err
	}

	if _, err := io.Copy(w, conn); err != nil {
		return err
	}

	if err := conn.Close(); err != nil {
		return err
	}

	return next.ServeHTTP(w, r)
}
