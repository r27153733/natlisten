package cloudflare

import (
	"github.com/r27153733/natlisten/natnet/nathttp"
	"net/http"
)

func (c *DDNSPortCli) HttpReuseListenAndServeIPV4DDNSPort(s *http.Server) error {
	return nathttp.ReuseListenAndServeIPV4PubNat(s, c.UpdateDNSPortCache)
}

func (c *DDNSPortCli) HttpReuseListenAndServeTLSIPV4DDNSPort(s *http.Server, certFile, keyFile string) error {
	return nathttp.ReuseListenAndServeTLSIPV4PubNat(s, certFile, keyFile, c.UpdateDNSPortCache)
}
