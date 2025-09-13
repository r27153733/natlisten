package fastly

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/r27153733/natlisten/natnet/nathttp"
)

func (c *IPPortCli) HttpReuseListenAndServeIPV4DDNSPort(ctx context.Context, s *http.Server) error {
	return nathttp.ReuseListenAndServeIPV4PubNat(ctx, s, c.UpdateIPPortCache)
}

func (c *IPPortCli) HttpReuseListenAndServeTLSIPV4DDNSPort(ctx context.Context, s *http.Server, certFile, keyFile string) error {
	return nathttp.ReuseListenAndServeTLSIPV4PubNat(ctx, s, certFile, keyFile, c.UpdateIPPortCache)
}

func (c *IPPortCli) HttpReuseListenAndServeTLSConfigIPV4DDNSPort(ctx context.Context, s *http.Server, config *tls.Config) error {
	return nathttp.ReuseListenAndServeTLSConfigIPV4PubNat(ctx, s, config, c.UpdateIPPortCache)
}
