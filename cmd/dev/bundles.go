package dev

import (
	"context"

	"github.com/host-uk/core/pkg/agentic"
	devpkg "github.com/host-uk/core/pkg/dev"
	"github.com/host-uk/core/pkg/framework"
	"github.com/host-uk/core/pkg/git"
)

// WorkBundle contains the Core instance for dev work operations.
type WorkBundle struct {
	Core *framework.Core
}

// WorkBundleOptions configures the work bundle.
type WorkBundleOptions struct {
	RegistryPath string
	AllowEdit    bool // Allow agentic to use Write/Edit tools
}

// NewWorkBundle creates a bundle for dev work operations.
// Includes: dev (orchestration), git, agentic services.
func NewWorkBundle(opts WorkBundleOptions) (*WorkBundle, error) {
	c, err := framework.New(
		framework.WithService(devpkg.NewService(devpkg.ServiceOptions{
			RegistryPath: opts.RegistryPath,
		})),
		framework.WithService(git.NewService(git.ServiceOptions{})),
		framework.WithService(agentic.NewService(agentic.ServiceOptions{
			AllowEdit: opts.AllowEdit,
		})),
		framework.WithServiceLock(),
	)
	if err != nil {
		return nil, err
	}

	return &WorkBundle{Core: c}, nil
}

// Start initialises the bundle services.
func (b *WorkBundle) Start(ctx context.Context) error {
	return b.Core.ServiceStartup(ctx, nil)
}

// Stop shuts down the bundle services.
func (b *WorkBundle) Stop(ctx context.Context) error {
	return b.Core.ServiceShutdown(ctx)
}

// StatusBundle contains the Core instance for status-only operations.
type StatusBundle struct {
	Core *framework.Core
}

// StatusBundleOptions configures the status bundle.
type StatusBundleOptions struct {
	RegistryPath string
}

// NewStatusBundle creates a bundle for status-only operations.
// Includes: dev (orchestration), git services. No agentic - commits not available.
func NewStatusBundle(opts StatusBundleOptions) (*StatusBundle, error) {
	c, err := framework.New(
		framework.WithService(devpkg.NewService(devpkg.ServiceOptions{
			RegistryPath: opts.RegistryPath,
		})),
		framework.WithService(git.NewService(git.ServiceOptions{})),
		// No agentic service - TaskCommit will be unhandled
		framework.WithServiceLock(),
	)
	if err != nil {
		return nil, err
	}

	return &StatusBundle{Core: c}, nil
}

// Start initialises the bundle services.
func (b *StatusBundle) Start(ctx context.Context) error {
	return b.Core.ServiceStartup(ctx, nil)
}

// Stop shuts down the bundle services.
func (b *StatusBundle) Stop(ctx context.Context) error {
	return b.Core.ServiceShutdown(ctx)
}
