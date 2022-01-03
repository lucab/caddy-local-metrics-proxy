package local_metrics_proxy

import (
	"path/filepath"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

const (
	// UnixBackendKind is the configuration kind for the unix-socket backend
	UnixBackendKind = "uds"
)

func init() {
	httpcaddyfile.RegisterHandlerDirective(ModuleName, parseCaddyfile)
}

// parseCaddyfile parses the local_metrics_proxy directive.
// This module proxies requests to a local metrics endpoint, and can be
// configured with this syntax:
//
//   local_metrics_proxy [<matcher>] {
//     uds {
//       path "/path/to/unix/socket"
//     }
//   }
//
func (lmp *LocalMetricsProxy) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	// skip block beginning: "local_metrics_proxy"
	for d.Next() {
		args := d.RemainingArgs()
		if len(args) > 0 {
			return d.Errf("extra unrecognized arguments: %v", args)
		}

		for nesting := d.Nesting(); d.NextBlock(nesting); {
			kind := d.Val()
			switch kind {
			case UnixBackendKind:
				if lmp.UdsBackend != nil {
					return d.Err("multiple uds blocks")
				}
				out, err := parseUdsBlock(lmp.logger, d)
				if err != nil {
					return err
				}
				lmp.UdsBackend = out
			default:
				return d.Errf("unrecognized backend kind: %v", kind)
			}
		}
	}

	if lmp.UdsBackend == nil {
		return d.Err("missing kind")
	}

	return nil
}

func parseUdsBlock(logger *zap.Logger, d *caddyfile.Dispenser) (*UnixBackend, error) {
	out := UnixBackend{
		Path: "",
	}

	args := d.RemainingArgs()
	if len(args) > 0 {
		return nil, d.Errf("extra unrecognized uds arguments: %v", args)
	}

	for nesting := d.Nesting(); d.NextBlock(nesting); {
		arg := d.Val()
		switch arg {
		case "path":
			if out.Path != "" {
				return nil, d.Err("multiple path arguments")
			}
			d.Next()
			out.Path = d.Val()
			if !filepath.IsAbs(out.Path) {
				return nil, d.Err("path value must be an absolute filepath")
			}
		default:
			return nil, d.Errf("unrecognized uds argument: %v", arg)
		}
	}
	if out.Path == "" {
		return nil, d.Err("missing path for uds kind")
	}

	return &out, nil
}

// parseCaddyfileHandler unmarshals tokens from h into a new middleware handler value.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var out LocalMetricsProxy
	err := out.UnmarshalCaddyfile(h.Dispenser)
	return out, err
}
