package fastly

import (
	"crypto/tls"
	"net/http"

	"github.com/r27153733/natlisten/natnet/nathttp"
)

func (c *IPPortCli) HttpReuseListenAndServeIPV4DDNSPort(s *http.Server) error {
	return nathttp.ReuseListenAndServeIPV4PubNat(s, c.UpdateIPPortCache)
}

func (c *IPPortCli) HttpReuseListenAndServeTLSIPV4DDNSPort(s *http.Server, certFile, keyFile string) error {
	return nathttp.ReuseListenAndServeTLSIPV4PubNat(s, certFile, keyFile, c.UpdateIPPortCache)
}

func (c *IPPortCli) HttpReuseListenAndServeTLSConfigIPV4DDNSPort(s *http.Server, config *tls.Config) error {
	return nathttp.ReuseListenAndServeTLSConfigIPV4PubNat(s, config, c.UpdateIPPortCache)
}
