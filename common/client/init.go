package client

import (
	"fmt"
	"github.com/jhopan/Panway/common/config"
	"github.com/jhopan/Panway/common/logger"
	"net/http"
	"net/url"
	"time"
)

var HTTPClient *http.Client
var ImpatientHTTPClient *http.Client
var UserContentRequestHTTPClient *http.Client

// userContentTransport membungkus transport dengan User-Agent browser.
//
// CDN dan host gambar (Wikimedia dan sejenisnya) menolak request tanpa
// User-Agent dengan HTTP 400/403. Tanpa header ini, EVERY request gambar
// dari user gagal — bukan hanya di test.
func userContentTransport(base http.RoundTripper) http.RoundTripper {
	return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		clone := req.Clone(req.Context())
		if clone.Header.Get("User-Agent") == "" {
			clone.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Panway/1.0; +https://github.com/jhopan/Panway)")
		}
		if clone.Header.Get("Accept") == "" {
			clone.Header.Set("Accept", "image/avif,image/webp,image/apng,image/*,*/*;q=0.8")
		}
		return base.RoundTrip(clone)
	})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func Init() {
	if config.UserContentRequestProxy != "" {
		logger.SysLog(fmt.Sprintf("using %s as proxy to fetch user content", config.UserContentRequestProxy))
		proxyURL, err := url.Parse(config.UserContentRequestProxy)
		if err != nil {
			logger.FatalLog(fmt.Sprintf("USER_CONTENT_REQUEST_PROXY set but invalid: %s", config.UserContentRequestProxy))
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		UserContentRequestHTTPClient = &http.Client{
			Transport: userContentTransport(transport),
			Timeout:   time.Second * time.Duration(config.UserContentRequestTimeout),
		}
	} else {
		UserContentRequestHTTPClient = &http.Client{
			Transport: userContentTransport(http.DefaultTransport),
		}
	}
	var transport http.RoundTripper
	if config.RelayProxy != "" {
		logger.SysLog(fmt.Sprintf("using %s as api relay proxy", config.RelayProxy))
		proxyURL, err := url.Parse(config.RelayProxy)
		if err != nil {
			logger.FatalLog(fmt.Sprintf("USER_CONTENT_REQUEST_PROXY set but invalid: %s", config.UserContentRequestProxy))
		}
		transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	if config.RelayTimeout == 0 {
		HTTPClient = &http.Client{
			Transport: transport,
		}
	} else {
		HTTPClient = &http.Client{
			Timeout:   time.Duration(config.RelayTimeout) * time.Second,
			Transport: transport,
		}
	}

	ImpatientHTTPClient = &http.Client{
		Timeout:   5 * time.Second,
		Transport: transport,
	}
}
