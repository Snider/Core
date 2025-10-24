package core

import (
	"embed"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type Contract struct {
	DontPanic      bool
	DisableLogging bool
}
type Ipc struct {
	Target string
}

var allowedModules = map[string]bool{
	"docs":    true,
	"display": true,
	// add more names here if you want to restrict what can be loaded
}

type Message interface{}
type Core struct {
	once    sync.Once
	initErr error
	App     *application.App
	assets  embed.FS

	modMu          sync.RWMutex
	mods           map[string]any
	ipcMu          sync.RWMutex
	ipcHandlers    []func(*Core, Message) error
	serviceLock    bool
	servicesLocked bool
}

type Option func(*Core) error

var (
	instance *Core
	once     sync.Once
	initErr  error
)
