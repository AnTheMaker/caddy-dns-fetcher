package DNSFetcher

import (
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(DNSFetcher{})
	httpcaddyfile.RegisterHandlerDirective("dnsfetcher", parseCaddyfile)
	httpcaddyfile.RegisterDirectiveOrder("dnsfetcher", "before", "basic_auth")
}

type DNSFetcher struct {
	Type   string `json:"type,omitempty"`
	Name   string `json:"name,omitempty"`
	logger *zap.Logger
}

// CaddyModule returns the Caddy module information.
func (DNSFetcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.matchers.dnsfetcher",
		New: func() caddy.Module { return new(DNSFetcher) },
	}
}

func (s *DNSFetcher) Provision(ctx caddy.Context) error {
	s.logger = ctx.Logger()
	return nil
}

func (s *DNSFetcher) Validate() error {
	if s.Type == "" {
		return fmt.Errorf("type is required")
	}

	switch strings.ToUpper(s.Type) {
	case "TXT", "IP", "A", "AAAA", "CNAME":
		// ok
	default:
		return fmt.Errorf("type set to unsupported dns record type")
	}
	if s.Name == "" {
		return fmt.Errorf("name is required")
	}
	if !regexp.MustCompile(`^([\p{L}\w\-]+\.)+[A-Za-z]{2,}$`).MatchString(s.Name) {
		return fmt.Errorf("name is not a valid hostname")
	}
	return nil
}

func (s DNSFetcher) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	response := ""

	switch strings.ToUpper(s.Type) {
	case "TXT":
		result, err := net.LookupTXT(s.Name)
		if err != nil || len(result) == 0 {
			return next.ServeHTTP(w, r)
		}
		response = result[0]
	case "IP", "A", "AAAA":
		result, err := net.LookupAddr(s.Name)
		if err != nil || len(result) == 0 {
			return next.ServeHTTP(w, r)
		}
		response = result[0]
	case "CNAME":
		result, err := net.LookupCNAME(s.Name)
		if err != nil || len(result) == 0 {
			return next.ServeHTTP(w, r)
		}
		response = result
	}

	repl := r.Context().Value(caddy.ReplacerCtxKey).(*caddy.Replacer)
	repl.Set("dnsfetcher.response", response)

	return next.ServeHTTP(w, r)
}

func (s *DNSFetcher) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		args := d.RemainingArgs()

		switch len(args) {
		case 2:
			s.Type = args[0]
			s.Name = args[1]
		default:
			return d.Err("unexpected number of arguments")
		}
	}

	return nil
}

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var s DNSFetcher
	err := s.UnmarshalCaddyfile(h.Dispenser)
	return s, err
}

// Interface guards
var (
	_ caddy.Provisioner           = (*DNSFetcher)(nil)
	_ caddy.Validator             = (*DNSFetcher)(nil)
	_ caddy.Module                = (*DNSFetcher)(nil)
	_ caddyhttp.MiddlewareHandler = (*DNSFetcher)(nil)
	_ caddyfile.Unmarshaler       = (*DNSFetcher)(nil)
)
