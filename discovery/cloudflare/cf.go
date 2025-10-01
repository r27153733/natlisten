package cloudflare

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	cloudflaresdk "github.com/cloudflare/cloudflare-go/v5"
	"github.com/cloudflare/cloudflare-go/v5/dns"
	"github.com/cloudflare/cloudflare-go/v5/option"
	"github.com/cloudflare/cloudflare-go/v5/rulesets"
)

type Config struct {
	Retry     int           `json:"retry"`
	Timeout   time.Duration `json:"timeout"`
	Zone      string        `json:"zone"`       // Cloudflare 区域 ID
	Record    string        `json:"record"`     // DNS 记录 ID
	RulesetID string        `json:"ruleset_id"` // 规则集 ID
	RuleID    string        `json:"rule_id"`    // 规则 ID
	APIKey    string        `json:"apiKey"`     // API 令牌
	Domain    string        `json:"domain"`     // 目标域名
}

type IPPortCli struct {
	dCli        dns.RecordService
	rCli        rulesets.RuleService
	cfg         Config
	expression  string
	cacheIPPort atomic.Pointer[ipPort]
}

func GetCli(cfg Config) IPPortCli {
	opt := []option.RequestOption{
		option.WithAPIToken(cfg.APIKey),
		option.WithEnvironmentProduction(),
		option.WithRequestTimeout(cfg.Timeout),
	}
	if cfg.Retry == 0 {
		cfg.Retry = 10
	}
	return IPPortCli{
		dCli:       dns.RecordService{Options: opt},
		rCli:       rulesets.RuleService{Options: opt},
		cfg:        cfg,
		expression: fmt.Sprintf(`(http.host wildcard "%s")`, cfg.Domain),
	}
}

func (c *IPPortCli) UpdateIPPort(ip net.IP, port int) error {
	ctx := context.Background()
	err := c.updateRule(ctx, port)
	if err != nil {
		return err
	}
	err = c.updateDNS(ctx, ip)
	if err != nil {
		return err
	}
	return err
}

func (c *IPPortCli) UpdateIPPortCache(ip net.IP, port int) error {
	load := c.cacheIPPort.Load()
	if load != nil && load.port == port && bytes.Equal(load.ip, ip) {
		return nil
	}
	err := c.UpdateIPPort(ip, port)
	if err != nil {
		return err
	}
	c.cacheIPPort.Store(&ipPort{
		ip:   ip,
		port: port,
	})
	return nil
}

func (c *IPPortCli) updateDNS(ctx context.Context, ip net.IP) (err error) {
	var recordUpdateParams dns.RecordUpdateParams
	if ip.To4() != nil {
		recordUpdateParams = dns.RecordUpdateParams{
			ZoneID: cloudflaresdk.F(c.cfg.Zone),
			Body: dns.ARecordParam{
				Name:    cloudflaresdk.F(c.cfg.Domain),
				TTL:     cloudflaresdk.F(dns.TTL(1)),
				Type:    cloudflaresdk.F(dns.ARecordTypeA),
				Content: cloudflaresdk.F(ip.String()),
				Proxied: cloudflaresdk.F(true),
			},
		}
	} else {
		recordUpdateParams = dns.RecordUpdateParams{
			ZoneID: cloudflaresdk.F(c.cfg.Zone),
			Body: dns.AAAARecordParam{
				Name:    cloudflaresdk.F(c.cfg.Domain),
				TTL:     cloudflaresdk.F(dns.TTL(1)),
				Type:    cloudflaresdk.F(dns.AAAARecordTypeAAAA),
				Content: cloudflaresdk.F(ip.String()),
				Proxied: cloudflaresdk.F(true),
			},
		}
	}
	i := c.cfg.Retry + 1
	for range i {
		_, err = c.dCli.Update(
			ctx,
			c.cfg.Record,
			recordUpdateParams,
		)
		if err == nil {
			return nil
		}
	}
	return err
}

func (c *IPPortCli) updateRule(ctx context.Context, port int) (err error) {
	ruleEditParams := rulesets.RuleEditParams{
		ZoneID: cloudflaresdk.F(c.cfg.Zone),
		Body: rulesets.RuleEditParamsBody{
			Expression:  cloudflaresdk.F(c.expression),
			Description: cloudflaresdk.F("nat"),
			Action:      cloudflaresdk.F(rulesets.RuleEditParamsBodyActionRoute),
			ActionParameters: cloudflaresdk.F[interface{}](routeRuleActionParameters{
				Origin: routeRuleActionParametersOrigin{
					Port: port,
				},
			}),
		},
	}

	for range c.cfg.Retry {
		_, err = c.rCli.Edit(ctx, c.cfg.RulesetID, c.cfg.RuleID, ruleEditParams)
		if err == nil {
			return nil
		}
	}
	return err
}

type routeRuleActionParameters struct {
	Origin routeRuleActionParametersOrigin `json:"origin"`
}

type routeRuleActionParametersOrigin struct {
	Port int `json:"port"`
}

type ipPort struct {
	ip   net.IP
	port int
}
