package local_metrics_proxy

import (
	"fmt"
	"testing"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/google/go-cmp/cmp"
)

const tf string = "Testfile"

func TestParseCaddyfile(t *testing.T) {
	testcases := []struct {
		name      string
		d         *caddyfile.Dispenser
		want      *LocalMetricsProxy
		shouldErr bool
		err       error
	}{
		{
			name: "test parse proper config",
			d: caddyfile.NewTestDispenser(`
            local_metrics_proxy {
		    uds {
			    path "/run/foo"
		    }
            }`),
			want: &LocalMetricsProxy{
				UdsBackend: &UnixBackend{
					Path: "/run/foo",
				},
			},
		},
		{
			name: "test parse config with unsupported argument",
			d: caddyfile.NewTestDispenser(`
            local_metrics_proxy foo_arg {
            }`),
			shouldErr: true,
			err:       fmt.Errorf("%s:%d - Error during parsing: extra unrecognized arguments: %s", tf, 2, "[foo_arg]"),
		},
		{
			name: "test parse config with unknown backend",
			d: caddyfile.NewTestDispenser(`
            local_metrics_proxy {
              foo_kind { }
            }`),
			shouldErr: true,
			err:       fmt.Errorf("%s:%d - Error during parsing: unrecognized backend kind: %s", tf, 3, "foo_kind"),
		},
		{
			name: "test parse config with missing path",
			d: caddyfile.NewTestDispenser(`
            local_metrics_proxy {
              uds { }
            }`),
			shouldErr: true,
			err:       fmt.Errorf("%s:%d - Error during parsing: missing path for uds kind", tf, 3),
		},
		{
			name: "test parse config with multiple path",
			d: caddyfile.NewTestDispenser(`
            local_metrics_proxy {
              uds {
		      path "/foo"
		      path "/bar"
	      }
            }`),
			shouldErr: true,
			err:       fmt.Errorf("%s:%d - Error during parsing: multiple path arguments", tf, 5),
		},
		{
			name: "test parse config with multiple uds blocks",
			d: caddyfile.NewTestDispenser(`
            local_metrics_proxy {
              uds {
		      path "/foo"
	      }
              uds {
		      path "/bar"
	      }
            }`),
			shouldErr: true,
			err:       fmt.Errorf("%s:%d - Error during parsing: multiple uds blocks", tf, 6),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

			var lmp LocalMetricsProxy
			err := lmp.UnmarshalCaddyfile(tc.d)
			if err != nil {
				if !tc.shouldErr {
					t.Fatalf("expected success, got: %v", err)
				}
				if diff := cmp.Diff(err.Error(), tc.err.Error()); diff != "" {
					t.Fatalf("unexpected error: %v, want: %v", err, tc.err)
				}
				return
			}
			if tc.shouldErr {
				t.Fatalf("unexpected success, want: %v", tc.err)
			}
			if diff := cmp.Diff(tc.want, &lmp, cmp.AllowUnexported(LocalMetricsProxy{})); diff != "" {
				t.Errorf("parseCaddyfile() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
