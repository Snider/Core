package php

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Options configures the development server.
type Options struct {
	// Dir is the Laravel project directory.
	Dir string

	// Services specifies which services to start.
	// If empty, services are auto-detected.
	Services []DetectedService

	// NoVite disables the Vite dev server.
	NoVite bool

	// NoHorizon disables Laravel Horizon.
	NoHorizon bool

	// NoReverb disables Laravel Reverb.
	NoReverb bool

	// NoRedis disables the Redis server.
	NoRedis bool

	// HTTPS enables HTTPS with mkcert certificates.
	HTTPS bool

	// Domain is the domain for SSL certificates.
	// Defaults to APP_URL from .env or "localhost".
	Domain string

	// Ports for each service
	FrankenPHPPort int
	HTTPSPort      int
	VitePort       int
	ReverbPort     int
	RedisPort      int
}

// DevServer manages all development services.
type DevServer struct {
	opts     Options
	services []Service
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.RWMutex
	running  bool
}

// NewDevServer creates a new development server manager.
func NewDevServer(opts Options) *DevServer {
	return &DevServer{
		opts:     opts,
		services: make([]Service, 0),
	}
}

// Start starts all detected/configured services.
func (d *DevServer) Start(ctx context.Context, opts Options) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.running {
		return fmt.Errorf("dev server is already running")
	}

	// Merge options
	if opts.Dir != "" {
		d.opts.Dir = opts.Dir
	}
	if d.opts.Dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		d.opts.Dir = cwd
	}

	// Verify this is a Laravel project
	if !IsLaravelProject(d.opts.Dir) {
		return fmt.Errorf("not a Laravel project: %s", d.opts.Dir)
	}

	// Create cancellable context
	d.ctx, d.cancel = context.WithCancel(ctx)

	// Detect or use provided services
	services := opts.Services
	if len(services) == 0 {
		services = DetectServices(d.opts.Dir)
	}

	// Filter out disabled services
	services = d.filterServices(services, opts)

	// Setup SSL if HTTPS is enabled
	var certFile, keyFile string
	if opts.HTTPS {
		domain := opts.Domain
		if domain == "" {
			// Try to get domain from APP_URL
			appURL := GetLaravelAppURL(d.opts.Dir)
			if appURL != "" {
				domain = ExtractDomainFromURL(appURL)
			}
		}
		if domain == "" {
			domain = "localhost"
		}

		var err error
		certFile, keyFile, err = SetupSSLIfNeeded(domain, SSLOptions{})
		if err != nil {
			return fmt.Errorf("failed to setup SSL: %w", err)
		}
	}

	// Create services
	d.services = make([]Service, 0)

	for _, svc := range services {
		var service Service

		switch svc {
		case ServiceFrankenPHP:
			port := opts.FrankenPHPPort
			if port == 0 {
				port = 8000
			}
			httpsPort := opts.HTTPSPort
			if httpsPort == 0 {
				httpsPort = 443
			}
			service = NewFrankenPHPService(d.opts.Dir, FrankenPHPOptions{
				Port:      port,
				HTTPSPort: httpsPort,
				HTTPS:     opts.HTTPS,
				CertFile:  certFile,
				KeyFile:   keyFile,
			})

		case ServiceVite:
			port := opts.VitePort
			if port == 0 {
				port = 5173
			}
			service = NewViteService(d.opts.Dir, ViteOptions{
				Port: port,
			})

		case ServiceHorizon:
			service = NewHorizonService(d.opts.Dir)

		case ServiceReverb:
			port := opts.ReverbPort
			if port == 0 {
				port = 8080
			}
			service = NewReverbService(d.opts.Dir, ReverbOptions{
				Port: port,
			})

		case ServiceRedis:
			port := opts.RedisPort
			if port == 0 {
				port = 6379
			}
			service = NewRedisService(d.opts.Dir, RedisOptions{
				Port: port,
			})
		}

		if service != nil {
			d.services = append(d.services, service)
		}
	}

	// Start all services
	var startErrors []error
	for _, svc := range d.services {
		if err := svc.Start(d.ctx); err != nil {
			startErrors = append(startErrors, fmt.Errorf("%s: %w", svc.Name(), err))
		}
	}

	if len(startErrors) > 0 {
		// Stop any services that did start
		for _, svc := range d.services {
			svc.Stop()
		}
		return fmt.Errorf("failed to start services: %v", startErrors)
	}

	d.running = true
	return nil
}

// filterServices removes disabled services from the list.
func (d *DevServer) filterServices(services []DetectedService, opts Options) []DetectedService {
	filtered := make([]DetectedService, 0)

	for _, svc := range services {
		switch svc {
		case ServiceVite:
			if !opts.NoVite {
				filtered = append(filtered, svc)
			}
		case ServiceHorizon:
			if !opts.NoHorizon {
				filtered = append(filtered, svc)
			}
		case ServiceReverb:
			if !opts.NoReverb {
				filtered = append(filtered, svc)
			}
		case ServiceRedis:
			if !opts.NoRedis {
				filtered = append(filtered, svc)
			}
		default:
			filtered = append(filtered, svc)
		}
	}

	return filtered
}

// Stop stops all services gracefully.
func (d *DevServer) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.running {
		return nil
	}

	// Cancel context first
	if d.cancel != nil {
		d.cancel()
	}

	// Stop all services in reverse order
	var stopErrors []error
	for i := len(d.services) - 1; i >= 0; i-- {
		svc := d.services[i]
		if err := svc.Stop(); err != nil {
			stopErrors = append(stopErrors, fmt.Errorf("%s: %w", svc.Name(), err))
		}
	}

	d.running = false

	if len(stopErrors) > 0 {
		return fmt.Errorf("errors stopping services: %v", stopErrors)
	}

	return nil
}

// Logs returns a reader for the specified service's logs.
// If service is empty, returns unified logs from all services.
func (d *DevServer) Logs(service string, follow bool) (io.ReadCloser, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if service == "" {
		// Return unified logs
		return d.unifiedLogs(follow)
	}

	// Find specific service
	for _, svc := range d.services {
		if svc.Name() == service {
			return svc.Logs(follow)
		}
	}

	return nil, fmt.Errorf("service not found: %s", service)
}

// unifiedLogs creates a reader that combines logs from all services.
func (d *DevServer) unifiedLogs(follow bool) (io.ReadCloser, error) {
	readers := make([]io.ReadCloser, 0)

	for _, svc := range d.services {
		reader, err := svc.Logs(follow)
		if err != nil {
			// Close any readers we already opened
			for _, r := range readers {
				r.Close()
			}
			return nil, fmt.Errorf("failed to get logs for %s: %w", svc.Name(), err)
		}
		readers = append(readers, reader)
	}

	return newMultiServiceReader(d.services, readers, follow), nil
}

// Status returns the status of all services.
func (d *DevServer) Status() []ServiceStatus {
	d.mu.RLock()
	defer d.mu.RUnlock()

	statuses := make([]ServiceStatus, 0, len(d.services))
	for _, svc := range d.services {
		statuses = append(statuses, svc.Status())
	}

	return statuses
}

// IsRunning returns true if the dev server is running.
func (d *DevServer) IsRunning() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.running
}

// Services returns the list of managed services.
func (d *DevServer) Services() []Service {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.services
}

// multiServiceReader combines multiple service log readers.
type multiServiceReader struct {
	services []Service
	readers  []io.ReadCloser
	follow   bool
	closed   bool
	mu       sync.RWMutex
}

func newMultiServiceReader(services []Service, readers []io.ReadCloser, follow bool) *multiServiceReader {
	return &multiServiceReader{
		services: services,
		readers:  readers,
		follow:   follow,
	}
}

func (m *multiServiceReader) Read(p []byte) (n int, err error) {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return 0, io.EOF
	}
	m.mu.RUnlock()

	// Round-robin read from all readers
	for i, reader := range m.readers {
		buf := make([]byte, len(p))
		n, err := reader.Read(buf)
		if n > 0 {
			// Prefix with service name
			prefix := fmt.Sprintf("[%s] ", m.services[i].Name())
			copy(p, prefix)
			copy(p[len(prefix):], buf[:n])
			return n + len(prefix), nil
		}
		if err != nil && err != io.EOF {
			return 0, err
		}
	}

	if m.follow {
		time.Sleep(100 * time.Millisecond)
		return 0, nil
	}

	return 0, io.EOF
}

func (m *multiServiceReader) Close() error {
	m.mu.Lock()
	m.closed = true
	m.mu.Unlock()

	var closeErr error
	for _, reader := range m.readers {
		if err := reader.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	return closeErr
}
