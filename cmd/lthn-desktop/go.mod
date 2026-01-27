module lthn-desktop

go 1.25.5

require (
	github.com/Snider/Core v0.0.0-00010101000000-000000000000
	github.com/Snider/Core/pkg/display v0.0.0
	github.com/Snider/Core/pkg/mcp v0.0.0-00010101000000-000000000000
	github.com/Snider/Core/pkg/module v0.0.0-00010101000000-000000000000
	github.com/Snider/Core/pkg/webview v0.0.0-00010101000000-000000000000
	github.com/Snider/Core/pkg/ws v0.0.0-00010101000000-000000000000
	github.com/Snider/Mining v0.0.9
	github.com/gin-gonic/gin v1.11.0
	github.com/gorilla/websocket v1.5.3
	github.com/wailsapp/wails/v3 v3.0.0-alpha.41
)

replace (
	github.com/Snider/Core => ../../
	github.com/Snider/Core/pkg/config => ../../pkg/config
	github.com/Snider/Core/pkg/core => ../../pkg/core
	github.com/Snider/Core/pkg/crypt => ../../pkg/crypt
	github.com/Snider/Core/pkg/display => ../../pkg/display
	github.com/Snider/Core/pkg/docs => ../../pkg/docs
	github.com/Snider/Core/pkg/help => ../../pkg/help
	github.com/Snider/Core/pkg/i18n => ../../pkg/i18n
	github.com/Snider/Core/pkg/ide => ../../pkg/ide
	github.com/Snider/Core/pkg/io => ../../pkg/io
	github.com/Snider/Core/pkg/mcp => ../../pkg/mcp
	github.com/Snider/Core/pkg/module => ../../pkg/module
	github.com/Snider/Core/pkg/plugin => ../../pkg/plugin
	github.com/Snider/Core/pkg/process => ../../pkg/process
	github.com/Snider/Core/pkg/runtime => ../../pkg/runtime
	github.com/Snider/Core/pkg/webview => ../../pkg/webview
	github.com/Snider/Core/pkg/workspace => ../../pkg/workspace
	github.com/Snider/Core/pkg/ws => ../../pkg/ws
)

require (
	dario.cat/mergo v1.0.2 // indirect
	git.sr.ht/~jackmordaunt/go-toast/v2 v2.0.3 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/Masterminds/semver/v3 v3.3.1 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.3.0 // indirect
	github.com/Snider/Core/pkg/config v0.0.0-00010101000000-000000000000 // indirect
	github.com/Snider/Core/pkg/core v0.0.0 // indirect
	github.com/Snider/Core/pkg/docs v0.0.0-00010101000000-000000000000 // indirect
	github.com/Snider/Core/pkg/help v0.0.0-00010101000000-000000000000 // indirect
	github.com/Snider/Core/pkg/i18n v0.0.0-00010101000000-000000000000 // indirect
	github.com/Snider/Core/pkg/ide v0.0.0-00010101000000-000000000000 // indirect
	github.com/Snider/Core/pkg/process v0.0.0-00010101000000-000000000000 // indirect
	github.com/Snider/Enchantrix v0.0.2 // indirect
	github.com/adrg/xdg v0.5.3 // indirect
	github.com/bep/debounce v1.2.1 // indirect
	github.com/bytedance/sonic v1.14.0 // indirect
	github.com/bytedance/sonic/loader v0.3.0 // indirect
	github.com/cloudflare/circl v1.6.3 // indirect
	github.com/cloudwego/base64x v0.1.6 // indirect
	github.com/cyphar/filepath-securejoin v0.6.1 // indirect
	github.com/ebitengine/purego v0.9.1 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.9 // indirect
	github.com/gin-contrib/cors v1.7.6 // indirect
	github.com/gin-contrib/sse v1.1.0 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.6.2 // indirect
	github.com/go-git/go-git/v5 v5.16.4 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.22.1 // indirect
	github.com/go-openapi/jsonreference v0.21.2 // indirect
	github.com/go-openapi/spec v0.22.0 // indirect
	github.com/go-openapi/swag/conv v0.25.1 // indirect
	github.com/go-openapi/swag/jsonname v0.25.1 // indirect
	github.com/go-openapi/swag/jsonutils v0.25.1 // indirect
	github.com/go-openapi/swag/loading v0.25.1 // indirect
	github.com/go-openapi/swag/stringutils v0.25.1 // indirect
	github.com/go-openapi/swag/typeutils v0.25.1 // indirect
	github.com/go-openapi/swag/yamlutils v0.25.1 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.27.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/goccy/go-yaml v1.18.0 // indirect
	github.com/godbus/dbus/v5 v5.2.0 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/google/jsonschema-go v0.3.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jchv/go-winloader v0.0.0-20250406163304-c1995be93bd1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kevinburke/ssh_config v1.4.0 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/leaanthony/go-ansi-parser v1.6.1 // indirect
	github.com/leaanthony/u v1.1.1 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lmittmann/tint v1.1.2 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modelcontextprotocol/go-sdk v1.2.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/nicksnyder/go-i18n/v2 v2.6.1 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pjbgf/sha1cd v0.5.0 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/quic-go/quic-go v0.57.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/samber/lo v1.52.0 // indirect
	github.com/sergi/go-diff v1.4.0 // indirect
	github.com/shirou/gopsutil/v4 v4.25.10 // indirect
	github.com/skeema/knownhosts v1.3.2 // indirect
	github.com/swaggo/files v1.0.1 // indirect
	github.com/swaggo/gin-swagger v1.6.0 // indirect
	github.com/swaggo/swag v1.16.6 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.3.0 // indirect
	github.com/wailsapp/go-webview2 v1.0.23 // indirect
	github.com/wailsapp/mimetype v1.4.1 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/arch v0.20.0 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/mod v0.31.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/oauth2 v0.33.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	golang.org/x/tools v0.40.0 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
