package serve

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// NewProxy returns an http.Handler that reverse-proxies to pbURL,
// injecting the given Bearer token into every proxied request.
// Any Authorization header supplied by the browser client is overwritten.
func NewProxy(pbURL, token string) http.Handler {
	target, _ := url.Parse(pbURL)

	return &httputil.ReverseProxy{ //nolint:exhaustruct // only Rewrite needed; all other fields have safe zero values
		Rewrite: func(req *httputil.ProxyRequest) {
			req.SetURL(target)
			req.Out.Host = target.Host
			req.Out.Header.Set("Authorization", "Bearer "+token)
		},
	}
}
