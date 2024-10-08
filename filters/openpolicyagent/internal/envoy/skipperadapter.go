package envoy

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	ext_authz_v3_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	ext_authz_v3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
)

func AdaptToExtAuthRequest(req *http.Request, metadata *ext_authz_v3_core.Metadata, contextExtensions map[string]string, rawBody []byte) (*ext_authz_v3.CheckRequest, error) {
	if err := validateURLForInvalidUTF8(req.URL); err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	headers := make(map[string]string, len(req.Header))
	for h, vv := range req.Header {
		// This makes headers in the input compatible with what Envoy does, i.e. allows to use policy fragments designed for envoy
		// See: https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/header_casing#http-1-1-header-casing
		headers[strings.ToLower(h)] = strings.Join(vv, ", ")
	}

	ereq := &ext_authz_v3.CheckRequest{
		Attributes: &ext_authz_v3.AttributeContext{
			Request: &ext_authz_v3.AttributeContext_Request{
				Http: &ext_authz_v3.AttributeContext_HttpRequest{
					Host:    req.Host,
					Method:  req.Method,
					Path:    req.URL.RequestURI(),
					Headers: headers,
					RawBody: rawBody,
				},
			},
			ContextExtensions: contextExtensions,
			MetadataContext:   metadata,
		},
	}

	return ereq, nil
}

func validateURLForInvalidUTF8(u *url.URL) error {
	if !utf8.ValidString(u.Path) {
		return fmt.Errorf("invalid utf8 in path: %q", u.Path)
	}

	decodedQuery, err := url.QueryUnescape(u.RawQuery)
	if err != nil {
		return fmt.Errorf("error unescaping query string %q: %w", u.RawQuery, err)
	}

	if !utf8.ValidString(decodedQuery) {
		return fmt.Errorf("invalid utf8 in query: %q", u.RawQuery)
	}

	return nil
}
