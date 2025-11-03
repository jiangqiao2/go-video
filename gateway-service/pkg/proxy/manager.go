package proxy

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"gateway-service/pkg/config"
)

// Manager maintains reverse proxies for each upstream service.
type Manager struct {
	proxies  map[string]*httputil.ReverseProxy
	services map[string]config.ServiceConfig
	logger   *logrus.Entry
}

// NewManager creates a proxy manager for the configured services.
func NewManager(services map[string]config.ServiceConfig, logger *logrus.Logger) (*Manager, error) {
	if len(services) == 0 {
		return nil, fmt.Errorf("no upstream services configured")
	}

	entry := logrus.NewEntry(logger).WithField("component", "proxy")
	proxies := make(map[string]*httputil.ReverseProxy, len(services))

	for name, svc := range services {
		target, err := url.Parse(svc.BaseURL)
		if err != nil {
			return nil, fmt.Errorf("parse service %q base url: %w", name, err)
		}

		proxy := newReverseProxy(target, svc, entry.WithField("service", name))
		proxies[name] = proxy
	}

	return &Manager{
		proxies:  proxies,
		services: services,
		logger:   entry,
	}, nil
}

// Handler returns a gin handler that forwards traffic according to route config.
func (m *Manager) Handler(route config.RouteConfig) (gin.HandlerFunc, error) {
	proxy, ok := m.proxies[route.TargetService]
	if !ok {
		return nil, fmt.Errorf("route %q references unknown service %q", route.Name, route.TargetService)
	}

	routeStrip := route.StripPrefix
	serviceCfg := m.services[route.TargetService]

	return func(c *gin.Context) {
		start := time.Now()
		originalPath := c.Request.URL.Path

		if routeStrip != "" && strings.HasPrefix(c.Request.URL.Path, routeStrip) {
			c.Request.URL.Path = ensureLeadingSlash(strings.TrimPrefix(c.Request.URL.Path, routeStrip))
			c.Request.URL.RawPath = c.Request.URL.Path
		}

		if serviceCfg.StripPrefix != "" && strings.HasPrefix(c.Request.URL.Path, serviceCfg.StripPrefix) {
			c.Request.URL.Path = ensureLeadingSlash(strings.TrimPrefix(c.Request.URL.Path, serviceCfg.StripPrefix))
			c.Request.URL.RawPath = c.Request.URL.Path
		}

		for k, v := range serviceCfg.StaticHeaders {
			c.Request.Header.Set(k, v)
		}
		for k, v := range route.StaticHeaders {
			c.Request.Header.Set(k, v)
		}

		proxy.ServeHTTP(c.Writer, c.Request)

		m.logger.WithFields(logrus.Fields{
			"route":        route.Name,
			"method":       c.Request.Method,
			"path":         originalPath,
			"target":       route.TargetService,
			"status":       c.Writer.Status(),
			"duration_ms":  time.Since(start).Milliseconds(),
			"client_ip":    c.ClientIP(),
			"user_agent":   c.Request.UserAgent(),
		}).Debug("proxied request")
	}, nil
}

func newReverseProxy(target *url.URL, svc config.ServiceConfig, logger *logrus.Entry) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
		if svc.PreserveHost {
			req.Header.Set("X-Forwarded-Host", req.Host)
			req.Host = target.Host
		}
		req.Header.Set("X-Forwarded-Proto", target.Scheme)
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: svc.Timeout, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	if target.Scheme == "https" {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: false,
		}
	}
	proxy.Transport = transport

	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		logger.WithError(err).Error("proxy error")
		http.Error(rw, "upstream service unavailable", http.StatusBadGateway)
	}

	return proxy
}

func ensureLeadingSlash(path string) string {
	if path == "" {
		return "/"
	}
	if path[0] != '/' {
		return "/" + path
	}
	return path
}

