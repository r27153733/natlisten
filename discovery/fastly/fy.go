package fastly

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	fastlysdk "github.com/fastly/go-fastly/v11/fastly"
)

type Config struct {
	APIKey      string        `json:"apiKey"`
	ServiceID   string        `json:"service_id"`
	BackendName string        `json:"backend_name"`
	Retry       int           `json:"retry"`
	Timeout     time.Duration `json:"timeout"`
}

type IPPortCli struct {
	client      fastlysdk.Client
	cfg         Config
	cacheIPPort atomic.Pointer[ipPort]
}

type ipPort struct {
	ip   net.IP
	port int
}

func GetCli(cfg Config) (IPPortCli, error) {
	if cfg.Retry == 0 {
		cfg.Retry = 10
	}

	client, err := fastlysdk.NewClient(cfg.APIKey)
	if err != nil {
		return IPPortCli{}, err
	}
	client.HTTPClient = &http.Client{
		Timeout: cfg.Timeout,
	}

	return IPPortCli{
		client: *client,
		cfg:    cfg,
	}, nil
}

func (c *IPPortCli) UpdateIPPort(ip net.IP, port int) error {
	ctx := context.Background()

	service, err := retry(ctx, &fastlysdk.GetServiceInput{
		ServiceID: c.cfg.ServiceID,
	}, c.cfg.Retry, c.client.GetService)
	if err != nil {
		return err
	}

	var activeVersion *fastlysdk.Version
	for _, version := range service.Versions {
		if version.Active != nil && *version.Active {
			activeVersion = version
			break
		}
	}
	if activeVersion == nil {
		return errors.New("no active version found")
	}

	backend, err := retry(ctx, &fastlysdk.GetBackendInput{
		ServiceID:      c.cfg.ServiceID,
		ServiceVersion: *activeVersion.Number,
		Name:           c.cfg.BackendName,
	}, c.cfg.Retry, c.client.GetBackend)
	if err != nil {
		return err
	}
	if backend.Port != nil && *backend.Port == port && backend.Address != nil && net.ParseIP(*backend.Address).Equal(ip) {
		return nil
	}

	clonedVersion, err := retry(ctx, &fastlysdk.CloneVersionInput{
		ServiceID:      c.cfg.ServiceID,
		ServiceVersion: *activeVersion.Number,
	}, c.cfg.Retry, c.client.CloneVersion)
	if err != nil {
		return err
	}

	_, err = retry(ctx, &fastlysdk.UpdateBackendInput{
		ServiceID:      c.cfg.ServiceID,
		ServiceVersion: *clonedVersion.Number,
		Name:           c.cfg.BackendName,
		Address:        fastlysdk.ToPointer(ip.String()),
		Port:           fastlysdk.ToPointer(port),
	}, c.cfg.Retry, c.client.UpdateBackend)
	if err != nil {
		return err
	}

	valid, err := retry(ctx, &fastlysdk.ValidateVersionInput{
		ServiceID:      c.cfg.ServiceID,
		ServiceVersion: *clonedVersion.Number,
	}, c.cfg.Retry, c.validateVersion)
	if err != nil || !valid {
		return err
	}

	_, err = retry(ctx, &fastlysdk.ActivateVersionInput{
		ServiceID:      c.cfg.ServiceID,
		ServiceVersion: *clonedVersion.Number,
	}, c.cfg.Retry, c.client.ActivateVersion)
	if err != nil {
		return err
	}

	return nil
}

// UpdateIPPortCache checks cache before updating to avoid unnecessary API calls
func (c *IPPortCli) UpdateIPPortCache(ip net.IP, port int) error {
	load := c.cacheIPPort.Load()
	if load != nil && load.port == port && ip.Equal(load.ip) {
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

func (c *IPPortCli) validateVersion(ctx context.Context, i *fastlysdk.ValidateVersionInput) (bool, error) {
	version, _, err := c.client.ValidateVersion(ctx, i)
	return version, err
}

func retry[Req any, Resp any](ctx context.Context, req Req, retry int, f func(ctx context.Context, req Req) (Resp, error)) (resp Resp, err error) {
	i := retry + 1
	for range i {
		resp, err = f(ctx, req)
		if err == nil {
			return resp, nil
		}
	}
	return resp, err
}
