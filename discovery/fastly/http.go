package fastly

import (
	"github.com/r27153733/natlisten/natnet/nathttp"
	"net/http"
)

func (c *IPPortCli) HttpReuseListenAndServeIPV4DDNSPort(s *http.Server) error {
	return nathttp.ReuseListenAndServeIPV4PubNat(s, c.UpdateIPPortCache)
}

func (c *IPPortCli) HttpReuseListenAndServeTLSIPV4DDNSPort(s *http.Server, certFile, keyFile string) error {
	return nathttp.ReuseListenAndServeTLSIPV4PubNat(s, certFile, keyFile, c.UpdateIPPortCache)
}
