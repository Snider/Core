package php

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFrankenPHPService_Good(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		dir := "/tmp/test"
		service := NewFrankenPHPService(dir, FrankenPHPOptions{})

		assert.Equal(t, "FrankenPHP", service.Name())
		assert.Equal(t, 8000, service.port)
		assert.Equal(t, 443, service.httpsPort)
		assert.False(t, service.https)
	})

	t.Run("custom options", func(t *testing.T) {
		dir := "/tmp/test"
		opts := FrankenPHPOptions{
			Port:      9000,
			HTTPSPort: 8443,
			HTTPS:     true,
			CertFile:  "cert.pem",
			KeyFile:   "key.pem",
		}
		service := NewFrankenPHPService(dir, opts)

		assert.Equal(t, 9000, service.port)
		assert.Equal(t, 8443, service.httpsPort)
		assert.True(t, service.https)
		assert.Equal(t, "cert.pem", service.certFile)
		assert.Equal(t, "key.pem", service.keyFile)
	})
}

func TestNewViteService_Good(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		dir := t.TempDir()
		service := NewViteService(dir, ViteOptions{})

		assert.Equal(t, "Vite", service.Name())
		assert.Equal(t, 5173, service.port)
		assert.Equal(t, "npm", service.packageManager) // default when no lock file
	})

	t.Run("custom package manager", func(t *testing.T) {
		dir := t.TempDir()
		service := NewViteService(dir, ViteOptions{PackageManager: "pnpm"})

		assert.Equal(t, "pnpm", service.packageManager)
	})
}

func TestNewHorizonService_Good(t *testing.T) {
	service := NewHorizonService("/tmp/test")
	assert.Equal(t, "Horizon", service.Name())
	assert.Equal(t, 0, service.port)
}

func TestNewReverbService_Good(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		service := NewReverbService("/tmp/test", ReverbOptions{})
		assert.Equal(t, "Reverb", service.Name())
		assert.Equal(t, 8080, service.port)
	})

	t.Run("custom port", func(t *testing.T) {
		service := NewReverbService("/tmp/test", ReverbOptions{Port: 9090})
		assert.Equal(t, 9090, service.port)
	})
}

func TestNewRedisService_Good(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		service := NewRedisService("/tmp/test", RedisOptions{})
		assert.Equal(t, "Redis", service.Name())
		assert.Equal(t, 6379, service.port)
	})

	t.Run("custom config", func(t *testing.T) {
		service := NewRedisService("/tmp/test", RedisOptions{ConfigFile: "redis.conf"})
		assert.Equal(t, "redis.conf", service.configFile)
	})
}

func TestBaseService_Status(t *testing.T) {
	s := &baseService{
		name:    "TestService",
		port:    1234,
		running: true,
	}

	status := s.Status()
	assert.Equal(t, "TestService", status.Name)
	assert.Equal(t, 1234, status.Port)
	assert.True(t, status.Running)
}
