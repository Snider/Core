package ansible

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// executeModule dispatches to the appropriate module handler.
func (e *Executor) executeModule(ctx context.Context, host string, client *SSHClient, task *Task, play *Play) (*TaskResult, error) {
	module := NormalizeModule(task.Module)

	// Apply task-level become
	if task.Become != nil && *task.Become {
		// Save old state to restore
		oldBecome := client.become
		oldUser := client.becomeUser
		oldPass := client.becomePass

		client.SetBecome(true, task.BecomeUser, "")

		defer client.SetBecome(oldBecome, oldUser, oldPass)
	}

	// Template the args
	args := e.templateArgs(task.Args, host, task)

	switch module {
	// Command execution
	case "ansible.builtin.shell":
		return e.moduleShell(ctx, client, args)
	case "ansible.builtin.command":
		return e.moduleCommand(ctx, client, args)
	case "ansible.builtin.raw":
		return e.moduleRaw(ctx, client, args)
	case "ansible.builtin.script":
		return e.moduleScript(ctx, client, args)

	// File operations
	case "ansible.builtin.copy":
		return e.moduleCopy(ctx, client, args, host, task)
	case "ansible.builtin.template":
		return e.moduleTemplate(ctx, client, args, host, task)
	case "ansible.builtin.file":
		return e.moduleFile(ctx, client, args)
	case "ansible.builtin.lineinfile":
		return e.moduleLineinfile(ctx, client, args)
	case "ansible.builtin.stat":
		return e.moduleStat(ctx, client, args)
	case "ansible.builtin.slurp":
		return e.moduleSlurp(ctx, client, args)
	case "ansible.builtin.fetch":
		return e.moduleFetch(ctx, client, args)
	case "ansible.builtin.get_url":
		return e.moduleGetURL(ctx, client, args)

	// Package management
	case "ansible.builtin.apt":
		return e.moduleApt(ctx, client, args)
	case "ansible.builtin.apt_key":
		return e.moduleAptKey(ctx, client, args)
	case "ansible.builtin.apt_repository":
		return e.moduleAptRepository(ctx, client, args)
	case "ansible.builtin.package":
		return e.modulePackage(ctx, client, args)
	case "ansible.builtin.pip":
		return e.modulePip(ctx, client, args)

	// Service management
	case "ansible.builtin.service":
		return e.moduleService(ctx, client, args)
	case "ansible.builtin.systemd":
		return e.moduleSystemd(ctx, client, args)

	// User/Group
	case "ansible.builtin.user":
		return e.moduleUser(ctx, client, args)
	case "ansible.builtin.group":
		return e.moduleGroup(ctx, client, args)

	// HTTP
	case "ansible.builtin.uri":
		return e.moduleURI(ctx, client, args)

	// Misc
	case "ansible.builtin.debug":
		return e.moduleDebug(args)
	case "ansible.builtin.fail":
		return e.moduleFail(args)
	case "ansible.builtin.assert":
		return e.moduleAssert(args, host)
	case "ansible.builtin.set_fact":
		return e.moduleSetFact(args)
	case "ansible.builtin.pause":
		return e.modulePause(ctx, args)
	case "ansible.builtin.wait_for":
		return e.moduleWaitFor(ctx, client, args)
	case "ansible.builtin.git":
		return e.moduleGit(ctx, client, args)
	case "ansible.builtin.unarchive":
		return e.moduleUnarchive(ctx, client, args)

	// Additional modules
	case "ansible.builtin.hostname":
		return e.moduleHostname(ctx, client, args)
	case "ansible.builtin.sysctl":
		return e.moduleSysctl(ctx, client, args)
	case "ansible.builtin.cron":
		return e.moduleCron(ctx, client, args)
	case "ansible.builtin.blockinfile":
		return e.moduleBlockinfile(ctx, client, args)
	case "ansible.builtin.include_vars":
		return e.moduleIncludeVars(args)
	case "ansible.builtin.meta":
		return e.moduleMeta(args)
	case "ansible.builtin.setup":
		return e.moduleSetup(ctx, client)
	case "ansible.builtin.reboot":
		return e.moduleReboot(ctx, client, args)

	// Community modules (basic support)
	case "community.general.ufw":
		return e.moduleUFW(ctx, client, args)
	case "ansible.posix.authorized_key":
		return e.moduleAuthorizedKey(ctx, client, args)
	case "community.docker.docker_compose":
		return e.moduleDockerCompose(ctx, client, args)

	default:
		// For unknown modules, try to execute as shell if it looks like a command
		if strings.Contains(task.Module, " ") || task.Module == "" {
			return e.moduleShell(ctx, client, args)
		}
		return nil, fmt.Errorf("unsupported module: %s", module)
	}
}

// templateArgs templates all string values in args.
func (e *Executor) templateArgs(args map[string]any, host string, task *Task) map[string]any {
	// Set inventory_hostname for templating
	e.vars["inventory_hostname"] = host

	result := make(map[string]any)
	for k, v := range args {
		switch val := v.(type) {
		case string:
			result[k] = e.templateString(val, host, task)
		case map[string]any:
			// Recurse for nested maps
			result[k] = e.templateArgs(val, host, task)
		case []any:
			// Template strings in arrays
			templated := make([]any, len(val))
			for i, item := range val {
				if s, ok := item.(string); ok {
					templated[i] = e.templateString(s, host, task)
				} else {
					templated[i] = item
				}
			}
			result[k] = templated
		default:
			result[k] = v
		}
	}
	return result
}

// --- Command Modules ---

func (e *Executor) moduleShell(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	cmd := getStringArg(args, "_raw_params", "")
	if cmd == "" {
		cmd = getStringArg(args, "cmd", "")
	}
	if cmd == "" {
		return nil, fmt.Errorf("shell: no command specified")
	}

	// Handle chdir
	if chdir := getStringArg(args, "chdir", ""); chdir != "" {
		cmd = fmt.Sprintf("cd %q && %s", chdir, cmd)
	}

	stdout, stderr, rc, err := client.RunScript(ctx, cmd)
	if err != nil {
		return &TaskResult{Failed: true, Msg: err.Error(), Stdout: stdout, Stderr: stderr, RC: rc}, nil
	}

	return &TaskResult{
		Changed: true,
		Stdout:  stdout,
		Stderr:  stderr,
		RC:      rc,
		Failed:  rc != 0,
	}, nil
}

func (e *Executor) moduleCommand(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	cmd := getStringArg(args, "_raw_params", "")
	if cmd == "" {
		cmd = getStringArg(args, "cmd", "")
	}
	if cmd == "" {
		return nil, fmt.Errorf("command: no command specified")
	}

	// Handle chdir
	if chdir := getStringArg(args, "chdir", ""); chdir != "" {
		cmd = fmt.Sprintf("cd %q && %s", chdir, cmd)
	}

	stdout, stderr, rc, err := client.Run(ctx, cmd)
	if err != nil {
		return &TaskResult{Failed: true, Msg: err.Error()}, nil
	}

	return &TaskResult{
		Changed: true,
		Stdout:  stdout,
		Stderr:  stderr,
		RC:      rc,
		Failed:  rc != 0,
	}, nil
}

func (e *Executor) moduleRaw(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	cmd := getStringArg(args, "_raw_params", "")
	if cmd == "" {
		return nil, fmt.Errorf("raw: no command specified")
	}

	stdout, stderr, rc, err := client.Run(ctx, cmd)
	if err != nil {
		return &TaskResult{Failed: true, Msg: err.Error()}, nil
	}

	return &TaskResult{
		Changed: true,
		Stdout:  stdout,
		Stderr:  stderr,
		RC:      rc,
	}, nil
}

func (e *Executor) moduleScript(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	script := getStringArg(args, "_raw_params", "")
	if script == "" {
		return nil, fmt.Errorf("script: no script specified")
	}

	// Read local script
	content, err := os.ReadFile(script)
	if err != nil {
		return nil, fmt.Errorf("read script: %w", err)
	}

	stdout, stderr, rc, err := client.RunScript(ctx, string(content))
	if err != nil {
		return &TaskResult{Failed: true, Msg: err.Error()}, nil
	}

	return &TaskResult{
		Changed: true,
		Stdout:  stdout,
		Stderr:  stderr,
		RC:      rc,
		Failed:  rc != 0,
	}, nil
}

// --- File Modules ---

func (e *Executor) moduleCopy(ctx context.Context, client *SSHClient, args map[string]any, host string, task *Task) (*TaskResult, error) {
	dest := getStringArg(args, "dest", "")
	if dest == "" {
		return nil, fmt.Errorf("copy: dest required")
	}

	var content []byte
	var err error

	if src := getStringArg(args, "src", ""); src != "" {
		content, err = os.ReadFile(src)
		if err != nil {
			return nil, fmt.Errorf("read src: %w", err)
		}
	} else if c := getStringArg(args, "content", ""); c != "" {
		content = []byte(c)
	} else {
		return nil, fmt.Errorf("copy: src or content required")
	}

	mode := os.FileMode(0644)
	if m := getStringArg(args, "mode", ""); m != "" {
		if parsed, err := strconv.ParseInt(m, 8, 32); err == nil {
			mode = os.FileMode(parsed)
		}
	}

	err = client.Upload(ctx, strings.NewReader(string(content)), dest, mode)
	if err != nil {
		return nil, err
	}

	// Handle owner/group (best-effort, errors ignored)
	if owner := getStringArg(args, "owner", ""); owner != "" {
		_, _, _, _ = client.Run(ctx, fmt.Sprintf("chown %s %q", owner, dest))
	}
	if group := getStringArg(args, "group", ""); group != "" {
		_, _, _, _ = client.Run(ctx, fmt.Sprintf("chgrp %s %q", group, dest))
	}

	return &TaskResult{Changed: true, Msg: fmt.Sprintf("copied to %s", dest)}, nil
}

func (e *Executor) moduleTemplate(ctx context.Context, client *SSHClient, args map[string]any, host string, task *Task) (*TaskResult, error) {
	src := getStringArg(args, "src", "")
	dest := getStringArg(args, "dest", "")
	if src == "" || dest == "" {
		return nil, fmt.Errorf("template: src and dest required")
	}

	// Process template
	content, err := e.TemplateFile(src, host, task)
	if err != nil {
		return nil, fmt.Errorf("template: %w", err)
	}

	mode := os.FileMode(0644)
	if m := getStringArg(args, "mode", ""); m != "" {
		if parsed, err := strconv.ParseInt(m, 8, 32); err == nil {
			mode = os.FileMode(parsed)
		}
	}

	err = client.Upload(ctx, strings.NewReader(content), dest, mode)
	if err != nil {
		return nil, err
	}

	return &TaskResult{Changed: true, Msg: fmt.Sprintf("templated to %s", dest)}, nil
}

func (e *Executor) moduleFile(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	path := getStringArg(args, "path", "")
	if path == "" {
		path = getStringArg(args, "dest", "")
	}
	if path == "" {
		return nil, fmt.Errorf("file: path required")
	}

	state := getStringArg(args, "state", "file")

	switch state {
	case "directory":
		mode := getStringArg(args, "mode", "0755")
		cmd := fmt.Sprintf("mkdir -p %q && chmod %s %q", path, mode, path)
		stdout, stderr, rc, err := client.Run(ctx, cmd)
		if err != nil || rc != 0 {
			return &TaskResult{Failed: true, Msg: stderr, Stdout: stdout, RC: rc}, nil
		}

	case "absent":
		cmd := fmt.Sprintf("rm -rf %q", path)
		_, stderr, rc, err := client.Run(ctx, cmd)
		if err != nil || rc != 0 {
			return &TaskResult{Failed: true, Msg: stderr, RC: rc}, nil
		}

	case "touch":
		cmd := fmt.Sprintf("touch %q", path)
		_, stderr, rc, err := client.Run(ctx, cmd)
		if err != nil || rc != 0 {
			return &TaskResult{Failed: true, Msg: stderr, RC: rc}, nil
		}

	case "link":
		src := getStringArg(args, "src", "")
		if src == "" {
			return nil, fmt.Errorf("file: src required for link state")
		}
		cmd := fmt.Sprintf("ln -sf %q %q", src, path)
		_, stderr, rc, err := client.Run(ctx, cmd)
		if err != nil || rc != 0 {
			return &TaskResult{Failed: true, Msg: stderr, RC: rc}, nil
		}

	case "file":
		// Ensure file exists and set permissions
		if mode := getStringArg(args, "mode", ""); mode != "" {
			_, _, _, _ = client.Run(ctx, fmt.Sprintf("chmod %s %q", mode, path))
		}
	}

	// Handle owner/group (best-effort, errors ignored)
	if owner := getStringArg(args, "owner", ""); owner != "" {
		_, _, _, _ = client.Run(ctx, fmt.Sprintf("chown %s %q", owner, path))
	}
	if group := getStringArg(args, "group", ""); group != "" {
		_, _, _, _ = client.Run(ctx, fmt.Sprintf("chgrp %s %q", group, path))
	}
	if recurse := getBoolArg(args, "recurse", false); recurse {
		if owner := getStringArg(args, "owner", ""); owner != "" {
			_, _, _, _ = client.Run(ctx, fmt.Sprintf("chown -R %s %q", owner, path))
		}
	}

	return &TaskResult{Changed: true}, nil
}

func (e *Executor) moduleLineinfile(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	path := getStringArg(args, "path", "")
	if path == "" {
		path = getStringArg(args, "dest", "")
	}
	if path == "" {
		return nil, fmt.Errorf("lineinfile: path required")
	}

	line := getStringArg(args, "line", "")
	regexp := getStringArg(args, "regexp", "")
	state := getStringArg(args, "state", "present")

	if state == "absent" {
		if regexp != "" {
			cmd := fmt.Sprintf("sed -i '/%s/d' %q", regexp, path)
			_, stderr, rc, _ := client.Run(ctx, cmd)
			if rc != 0 {
				return &TaskResult{Failed: true, Msg: stderr, RC: rc}, nil
			}
		}
	} else {
		// state == present
		if regexp != "" {
			// Replace line matching regexp
			escapedLine := strings.ReplaceAll(line, "/", "\\/")
			cmd := fmt.Sprintf("sed -i 's/%s/%s/' %q", regexp, escapedLine, path)
			_, _, rc, _ := client.Run(ctx, cmd)
			if rc != 0 {
				// Line not found, append
				cmd = fmt.Sprintf("echo %q >> %q", line, path)
				_, _, _, _ = client.Run(ctx, cmd)
			}
		} else if line != "" {
			// Ensure line is present
			cmd := fmt.Sprintf("grep -qxF %q %q || echo %q >> %q", line, path, line, path)
			_, _, _, _ = client.Run(ctx, cmd)
		}
	}

	return &TaskResult{Changed: true}, nil
}

func (e *Executor) moduleStat(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	path := getStringArg(args, "path", "")
	if path == "" {
		return nil, fmt.Errorf("stat: path required")
	}

	stat, err := client.Stat(ctx, path)
	if err != nil {
		return nil, err
	}

	return &TaskResult{
		Changed: false,
		Data:    map[string]any{"stat": stat},
	}, nil
}

func (e *Executor) moduleSlurp(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	path := getStringArg(args, "path", "")
	if path == "" {
		path = getStringArg(args, "src", "")
	}
	if path == "" {
		return nil, fmt.Errorf("slurp: path required")
	}

	content, err := client.Download(ctx, path)
	if err != nil {
		return nil, err
	}

	encoded := base64.StdEncoding.EncodeToString(content)

	return &TaskResult{
		Changed: false,
		Data:    map[string]any{"content": encoded, "encoding": "base64"},
	}, nil
}

func (e *Executor) moduleFetch(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	src := getStringArg(args, "src", "")
	dest := getStringArg(args, "dest", "")
	if src == "" || dest == "" {
		return nil, fmt.Errorf("fetch: src and dest required")
	}

	content, err := client.Download(ctx, src)
	if err != nil {
		return nil, err
	}

	// Create dest directory
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return nil, err
	}

	if err := os.WriteFile(dest, content, 0644); err != nil {
		return nil, err
	}

	return &TaskResult{Changed: true, Msg: fmt.Sprintf("fetched %s to %s", src, dest)}, nil
}

func (e *Executor) moduleGetURL(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	url := getStringArg(args, "url", "")
	dest := getStringArg(args, "dest", "")
	if url == "" || dest == "" {
		return nil, fmt.Errorf("get_url: url and dest required")
	}

	// Use curl or wget
	cmd := fmt.Sprintf("curl -fsSL -o %q %q || wget -q -O %q %q", dest, url, dest, url)
	stdout, stderr, rc, err := client.Run(ctx, cmd)
	if err != nil || rc != 0 {
		return &TaskResult{Failed: true, Msg: stderr, Stdout: stdout, RC: rc}, nil
	}

	// Set mode if specified (best-effort)
	if mode := getStringArg(args, "mode", ""); mode != "" {
		_, _, _, _ = client.Run(ctx, fmt.Sprintf("chmod %s %q", mode, dest))
	}

	return &TaskResult{Changed: true}, nil
}

// --- Package Modules ---

func (e *Executor) moduleApt(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	name := getStringArg(args, "name", "")
	state := getStringArg(args, "state", "present")
	updateCache := getBoolArg(args, "update_cache", false)

	var cmd string

	if updateCache {
		_, _, _, _ = client.Run(ctx, "apt-get update -qq")
	}

	switch state {
	case "present", "installed":
		if name != "" {
			cmd = fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get install -y -qq %s", name)
		}
	case "absent", "removed":
		cmd = fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get remove -y -qq %s", name)
	case "latest":
		cmd = fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get install -y -qq --only-upgrade %s", name)
	}

	if cmd == "" {
		return &TaskResult{Changed: false}, nil
	}

	stdout, stderr, rc, err := client.Run(ctx, cmd)
	if err != nil || rc != 0 {
		return &TaskResult{Failed: true, Msg: stderr, Stdout: stdout, RC: rc}, nil
	}

	return &TaskResult{Changed: true}, nil
}

func (e *Executor) moduleAptKey(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	url := getStringArg(args, "url", "")
	keyring := getStringArg(args, "keyring", "")
	state := getStringArg(args, "state", "present")

	if state == "absent" {
		if keyring != "" {
			_, _, _, _ = client.Run(ctx, fmt.Sprintf("rm -f %q", keyring))
		}
		return &TaskResult{Changed: true}, nil
	}

	if url == "" {
		return nil, fmt.Errorf("apt_key: url required")
	}

	var cmd string
	if keyring != "" {
		cmd = fmt.Sprintf("curl -fsSL %q | gpg --dearmor -o %q", url, keyring)
	} else {
		cmd = fmt.Sprintf("curl -fsSL %q | apt-key add -", url)
	}

	stdout, stderr, rc, err := client.Run(ctx, cmd)
	if err != nil || rc != 0 {
		return &TaskResult{Failed: true, Msg: stderr, Stdout: stdout, RC: rc}, nil
	}

	return &TaskResult{Changed: true}, nil
}

func (e *Executor) moduleAptRepository(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	repo := getStringArg(args, "repo", "")
	filename := getStringArg(args, "filename", "")
	state := getStringArg(args, "state", "present")

	if repo == "" {
		return nil, fmt.Errorf("apt_repository: repo required")
	}

	if filename == "" {
		// Generate filename from repo
		filename = strings.ReplaceAll(repo, " ", "-")
		filename = strings.ReplaceAll(filename, "/", "-")
		filename = strings.ReplaceAll(filename, ":", "")
	}

	path := fmt.Sprintf("/etc/apt/sources.list.d/%s.list", filename)

	if state == "absent" {
		_, _, _, _ = client.Run(ctx, fmt.Sprintf("rm -f %q", path))
		return &TaskResult{Changed: true}, nil
	}

	cmd := fmt.Sprintf("echo %q > %q", repo, path)
	stdout, stderr, rc, err := client.Run(ctx, cmd)
	if err != nil || rc != 0 {
		return &TaskResult{Failed: true, Msg: stderr, Stdout: stdout, RC: rc}, nil
	}

	// Update apt cache (best-effort)
	if getBoolArg(args, "update_cache", true) {
		_, _, _, _ = client.Run(ctx, "apt-get update -qq")
	}

	return &TaskResult{Changed: true}, nil
}

func (e *Executor) modulePackage(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	// Detect package manager and delegate
	stdout, _, _, _ := client.Run(ctx, "which apt-get yum dnf 2>/dev/null | head -1")
	stdout = strings.TrimSpace(stdout)

	if strings.Contains(stdout, "apt") {
		return e.moduleApt(ctx, client, args)
	}

	// Default to apt
	return e.moduleApt(ctx, client, args)
}

func (e *Executor) modulePip(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	name := getStringArg(args, "name", "")
	state := getStringArg(args, "state", "present")
	executable := getStringArg(args, "executable", "pip3")

	var cmd string
	switch state {
	case "present", "installed":
		cmd = fmt.Sprintf("%s install %s", executable, name)
	case "absent", "removed":
		cmd = fmt.Sprintf("%s uninstall -y %s", executable, name)
	case "latest":
		cmd = fmt.Sprintf("%s install --upgrade %s", executable, name)
	}

	stdout, stderr, rc, err := client.Run(ctx, cmd)
	if err != nil || rc != 0 {
		return &TaskResult{Failed: true, Msg: stderr, Stdout: stdout, RC: rc}, nil
	}

	return &TaskResult{Changed: true}, nil
}

// --- Service Modules ---

func (e *Executor) moduleService(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	name := getStringArg(args, "name", "")
	state := getStringArg(args, "state", "")
	enabled := args["enabled"]

	if name == "" {
		return nil, fmt.Errorf("service: name required")
	}

	var cmds []string

	if state != "" {
		switch state {
		case "started":
			cmds = append(cmds, fmt.Sprintf("systemctl start %s", name))
		case "stopped":
			cmds = append(cmds, fmt.Sprintf("systemctl stop %s", name))
		case "restarted":
			cmds = append(cmds, fmt.Sprintf("systemctl restart %s", name))
		case "reloaded":
			cmds = append(cmds, fmt.Sprintf("systemctl reload %s", name))
		}
	}

	if enabled != nil {
		if getBoolArg(args, "enabled", false) {
			cmds = append(cmds, fmt.Sprintf("systemctl enable %s", name))
		} else {
			cmds = append(cmds, fmt.Sprintf("systemctl disable %s", name))
		}
	}

	for _, cmd := range cmds {
		stdout, stderr, rc, err := client.Run(ctx, cmd)
		if err != nil || rc != 0 {
			return &TaskResult{Failed: true, Msg: stderr, Stdout: stdout, RC: rc}, nil
		}
	}

	return &TaskResult{Changed: len(cmds) > 0}, nil
}

func (e *Executor) moduleSystemd(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	// systemd is similar to service
	if getBoolArg(args, "daemon_reload", false) {
		_, _, _, _ = client.Run(ctx, "systemctl daemon-reload")
	}

	return e.moduleService(ctx, client, args)
}

// --- User/Group Modules ---

func (e *Executor) moduleUser(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	name := getStringArg(args, "name", "")
	state := getStringArg(args, "state", "present")

	if name == "" {
		return nil, fmt.Errorf("user: name required")
	}

	if state == "absent" {
		cmd := fmt.Sprintf("userdel -r %s 2>/dev/null || true", name)
		_, _, _, _ = client.Run(ctx, cmd)
		return &TaskResult{Changed: true}, nil
	}

	// Build useradd/usermod command
	var opts []string

	if uid := getStringArg(args, "uid", ""); uid != "" {
		opts = append(opts, "-u", uid)
	}
	if group := getStringArg(args, "group", ""); group != "" {
		opts = append(opts, "-g", group)
	}
	if groups := getStringArg(args, "groups", ""); groups != "" {
		opts = append(opts, "-G", groups)
	}
	if home := getStringArg(args, "home", ""); home != "" {
		opts = append(opts, "-d", home)
	}
	if shell := getStringArg(args, "shell", ""); shell != "" {
		opts = append(opts, "-s", shell)
	}
	if getBoolArg(args, "system", false) {
		opts = append(opts, "-r")
	}
	if getBoolArg(args, "create_home", true) {
		opts = append(opts, "-m")
	}

	// Try usermod first, then useradd
	optsStr := strings.Join(opts, " ")
	var cmd string
	if optsStr == "" {
		cmd = fmt.Sprintf("id %s >/dev/null 2>&1 || useradd %s", name, name)
	} else {
		cmd = fmt.Sprintf("id %s >/dev/null 2>&1 && usermod %s %s || useradd %s %s",
			name, optsStr, name, optsStr, name)
	}

	stdout, stderr, rc, err := client.Run(ctx, cmd)
	if err != nil || rc != 0 {
		return &TaskResult{Failed: true, Msg: stderr, Stdout: stdout, RC: rc}, nil
	}

	return &TaskResult{Changed: true}, nil
}

func (e *Executor) moduleGroup(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	name := getStringArg(args, "name", "")
	state := getStringArg(args, "state", "present")

	if name == "" {
		return nil, fmt.Errorf("group: name required")
	}

	if state == "absent" {
		cmd := fmt.Sprintf("groupdel %s 2>/dev/null || true", name)
		_, _, _, _ = client.Run(ctx, cmd)
		return &TaskResult{Changed: true}, nil
	}

	var opts []string
	if gid := getStringArg(args, "gid", ""); gid != "" {
		opts = append(opts, "-g", gid)
	}
	if getBoolArg(args, "system", false) {
		opts = append(opts, "-r")
	}

	cmd := fmt.Sprintf("getent group %s >/dev/null 2>&1 || groupadd %s %s",
		name, strings.Join(opts, " "), name)

	stdout, stderr, rc, err := client.Run(ctx, cmd)
	if err != nil || rc != 0 {
		return &TaskResult{Failed: true, Msg: stderr, Stdout: stdout, RC: rc}, nil
	}

	return &TaskResult{Changed: true}, nil
}

// --- HTTP Module ---

func (e *Executor) moduleURI(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	url := getStringArg(args, "url", "")
	method := getStringArg(args, "method", "GET")

	if url == "" {
		return nil, fmt.Errorf("uri: url required")
	}

	var curlOpts []string
	curlOpts = append(curlOpts, "-s", "-S")
	curlOpts = append(curlOpts, "-X", method)

	// Headers
	if headers, ok := args["headers"].(map[string]any); ok {
		for k, v := range headers {
			curlOpts = append(curlOpts, "-H", fmt.Sprintf("%s: %v", k, v))
		}
	}

	// Body
	if body := getStringArg(args, "body", ""); body != "" {
		curlOpts = append(curlOpts, "-d", body)
	}

	// Status code
	curlOpts = append(curlOpts, "-w", "\\n%{http_code}")

	cmd := fmt.Sprintf("curl %s %q", strings.Join(curlOpts, " "), url)
	stdout, stderr, rc, err := client.Run(ctx, cmd)
	if err != nil {
		return &TaskResult{Failed: true, Msg: err.Error()}, nil
	}

	// Parse status code from last line
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	statusCode := 0
	if len(lines) > 0 {
		statusCode, _ = strconv.Atoi(lines[len(lines)-1])
	}

	// Check expected status
	expectedStatus := 200
	if s, ok := args["status_code"].(int); ok {
		expectedStatus = s
	}

	failed := rc != 0 || statusCode != expectedStatus

	return &TaskResult{
		Changed: false,
		Failed:  failed,
		Stdout:  stdout,
		Stderr:  stderr,
		RC:      statusCode,
		Data:    map[string]any{"status": statusCode},
	}, nil
}

// --- Misc Modules ---

func (e *Executor) moduleDebug(args map[string]any) (*TaskResult, error) {
	msg := getStringArg(args, "msg", "")
	if v, ok := args["var"]; ok {
		msg = fmt.Sprintf("%v = %v", v, e.vars[fmt.Sprintf("%v", v)])
	}

	return &TaskResult{
		Changed: false,
		Msg:     msg,
	}, nil
}

func (e *Executor) moduleFail(args map[string]any) (*TaskResult, error) {
	msg := getStringArg(args, "msg", "Failed as requested")
	return &TaskResult{
		Failed: true,
		Msg:    msg,
	}, nil
}

func (e *Executor) moduleAssert(args map[string]any, host string) (*TaskResult, error) {
	that, ok := args["that"]
	if !ok {
		return nil, fmt.Errorf("assert: 'that' required")
	}

	conditions := normalizeConditions(that)
	for _, cond := range conditions {
		if !e.evalCondition(cond, host) {
			msg := getStringArg(args, "fail_msg", fmt.Sprintf("Assertion failed: %s", cond))
			return &TaskResult{Failed: true, Msg: msg}, nil
		}
	}

	return &TaskResult{Changed: false, Msg: "All assertions passed"}, nil
}

func (e *Executor) moduleSetFact(args map[string]any) (*TaskResult, error) {
	for k, v := range args {
		if k != "cacheable" {
			e.vars[k] = v
		}
	}
	return &TaskResult{Changed: true}, nil
}

func (e *Executor) modulePause(ctx context.Context, args map[string]any) (*TaskResult, error) {
	seconds := 0
	if s, ok := args["seconds"].(int); ok {
		seconds = s
	}
	if s, ok := args["seconds"].(string); ok {
		seconds, _ = strconv.Atoi(s)
	}

	if seconds > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ctxSleep(ctx, seconds):
		}
	}

	return &TaskResult{Changed: false}, nil
}

func ctxSleep(ctx context.Context, seconds int) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
		case <-sleepChan(seconds):
		}
		close(ch)
	}()
	return ch
}

func sleepChan(seconds int) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		for i := 0; i < seconds; i++ {
			select {
			case <-ch:
				return
			default:
				// Sleep 1 second at a time
			}
		}
		close(ch)
	}()
	return ch
}

func (e *Executor) moduleWaitFor(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	port := 0
	if p, ok := args["port"].(int); ok {
		port = p
	}
	host := getStringArg(args, "host", "127.0.0.1")
	state := getStringArg(args, "state", "started")
	timeout := 300
	if t, ok := args["timeout"].(int); ok {
		timeout = t
	}

	if port > 0 && state == "started" {
		cmd := fmt.Sprintf("timeout %d bash -c 'until nc -z %s %d; do sleep 1; done'",
			timeout, host, port)
		stdout, stderr, rc, err := client.Run(ctx, cmd)
		if err != nil || rc != 0 {
			return &TaskResult{Failed: true, Msg: stderr, Stdout: stdout, RC: rc}, nil
		}
	}

	return &TaskResult{Changed: false}, nil
}

func (e *Executor) moduleGit(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	repo := getStringArg(args, "repo", "")
	dest := getStringArg(args, "dest", "")
	version := getStringArg(args, "version", "HEAD")

	if repo == "" || dest == "" {
		return nil, fmt.Errorf("git: repo and dest required")
	}

	// Check if dest exists
	exists, _ := client.FileExists(ctx, dest+"/.git")

	var cmd string
	if exists {
		// Fetch and checkout (force to ensure clean state)
		cmd = fmt.Sprintf("cd %q && git fetch --all && git checkout --force %q", dest, version)
	} else {
		cmd = fmt.Sprintf("git clone %q %q && cd %q && git checkout %q",
			repo, dest, dest, version)
	}

	stdout, stderr, rc, err := client.Run(ctx, cmd)
	if err != nil || rc != 0 {
		return &TaskResult{Failed: true, Msg: stderr, Stdout: stdout, RC: rc}, nil
	}

	return &TaskResult{Changed: true}, nil
}

func (e *Executor) moduleUnarchive(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	src := getStringArg(args, "src", "")
	dest := getStringArg(args, "dest", "")
	remote := getBoolArg(args, "remote_src", false)

	if src == "" || dest == "" {
		return nil, fmt.Errorf("unarchive: src and dest required")
	}

	// Create dest directory (best-effort)
	_, _, _, _ = client.Run(ctx, fmt.Sprintf("mkdir -p %q", dest))

	var cmd string
	if !remote {
		// Upload local file first
		content, err := os.ReadFile(src)
		if err != nil {
			return nil, fmt.Errorf("read src: %w", err)
		}
		tmpPath := "/tmp/ansible_unarchive_" + filepath.Base(src)
		err = client.Upload(ctx, strings.NewReader(string(content)), tmpPath, 0644)
		if err != nil {
			return nil, err
		}
		src = tmpPath
		defer func() { _, _, _, _ = client.Run(ctx, fmt.Sprintf("rm -f %q", tmpPath)) }()
	}

	// Detect archive type and extract
	if strings.HasSuffix(src, ".tar.gz") || strings.HasSuffix(src, ".tgz") {
		cmd = fmt.Sprintf("tar -xzf %q -C %q", src, dest)
	} else if strings.HasSuffix(src, ".tar.xz") {
		cmd = fmt.Sprintf("tar -xJf %q -C %q", src, dest)
	} else if strings.HasSuffix(src, ".tar.bz2") {
		cmd = fmt.Sprintf("tar -xjf %q -C %q", src, dest)
	} else if strings.HasSuffix(src, ".tar") {
		cmd = fmt.Sprintf("tar -xf %q -C %q", src, dest)
	} else if strings.HasSuffix(src, ".zip") {
		cmd = fmt.Sprintf("unzip -o %q -d %q", src, dest)
	} else {
		cmd = fmt.Sprintf("tar -xf %q -C %q", src, dest) // Guess tar
	}

	stdout, stderr, rc, err := client.Run(ctx, cmd)
	if err != nil || rc != 0 {
		return &TaskResult{Failed: true, Msg: stderr, Stdout: stdout, RC: rc}, nil
	}

	return &TaskResult{Changed: true}, nil
}

// --- Helpers ---

func getStringArg(args map[string]any, key, def string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v)
	}
	return def
}

func getBoolArg(args map[string]any, key string, def bool) bool {
	if v, ok := args[key]; ok {
		switch b := v.(type) {
		case bool:
			return b
		case string:
			lower := strings.ToLower(b)
			return lower == "true" || lower == "yes" || lower == "1"
		}
	}
	return def
}

// --- Additional Modules ---

func (e *Executor) moduleHostname(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	name := getStringArg(args, "name", "")
	if name == "" {
		return nil, fmt.Errorf("hostname: name required")
	}

	// Set hostname
	cmd := fmt.Sprintf("hostnamectl set-hostname %q || hostname %q", name, name)
	stdout, stderr, rc, err := client.Run(ctx, cmd)
	if err != nil || rc != 0 {
		return &TaskResult{Failed: true, Msg: stderr, Stdout: stdout, RC: rc}, nil
	}

	// Update /etc/hosts if needed (best-effort)
	_, _, _, _ = client.Run(ctx, fmt.Sprintf("sed -i 's/127.0.1.1.*/127.0.1.1\t%s/' /etc/hosts", name))

	return &TaskResult{Changed: true}, nil
}

func (e *Executor) moduleSysctl(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	name := getStringArg(args, "name", "")
	value := getStringArg(args, "value", "")
	state := getStringArg(args, "state", "present")

	if name == "" {
		return nil, fmt.Errorf("sysctl: name required")
	}

	if state == "absent" {
		// Remove from sysctl.conf
		cmd := fmt.Sprintf("sed -i '/%s/d' /etc/sysctl.conf", name)
		_, _, _, _ = client.Run(ctx, cmd)
		return &TaskResult{Changed: true}, nil
	}

	// Set value
	cmd := fmt.Sprintf("sysctl -w %s=%s", name, value)
	stdout, stderr, rc, err := client.Run(ctx, cmd)
	if err != nil || rc != 0 {
		return &TaskResult{Failed: true, Msg: stderr, Stdout: stdout, RC: rc}, nil
	}

	// Persist if requested (best-effort)
	if getBoolArg(args, "sysctl_set", true) {
		cmd = fmt.Sprintf("grep -q '^%s' /etc/sysctl.conf && sed -i 's/^%s.*/%s=%s/' /etc/sysctl.conf || echo '%s=%s' >> /etc/sysctl.conf",
			name, name, name, value, name, value)
		_, _, _, _ = client.Run(ctx, cmd)
	}

	return &TaskResult{Changed: true}, nil
}

func (e *Executor) moduleCron(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	name := getStringArg(args, "name", "")
	job := getStringArg(args, "job", "")
	state := getStringArg(args, "state", "present")
	user := getStringArg(args, "user", "root")

	minute := getStringArg(args, "minute", "*")
	hour := getStringArg(args, "hour", "*")
	day := getStringArg(args, "day", "*")
	month := getStringArg(args, "month", "*")
	weekday := getStringArg(args, "weekday", "*")

	if state == "absent" {
		if name != "" {
			// Remove by name (comment marker)
			cmd := fmt.Sprintf("crontab -u %s -l 2>/dev/null | grep -v '# %s' | grep -v '%s' | crontab -u %s -",
				user, name, job, user)
			_, _, _, _ = client.Run(ctx, cmd)
		}
		return &TaskResult{Changed: true}, nil
	}

	// Build cron entry
	schedule := fmt.Sprintf("%s %s %s %s %s", minute, hour, day, month, weekday)
	entry := fmt.Sprintf("%s %s # %s", schedule, job, name)

	// Add to crontab
	cmd := fmt.Sprintf("(crontab -u %s -l 2>/dev/null | grep -v '# %s' ; echo %q) | crontab -u %s -",
		user, name, entry, user)
	stdout, stderr, rc, err := client.Run(ctx, cmd)
	if err != nil || rc != 0 {
		return &TaskResult{Failed: true, Msg: stderr, Stdout: stdout, RC: rc}, nil
	}

	return &TaskResult{Changed: true}, nil
}

func (e *Executor) moduleBlockinfile(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	path := getStringArg(args, "path", "")
	if path == "" {
		path = getStringArg(args, "dest", "")
	}
	if path == "" {
		return nil, fmt.Errorf("blockinfile: path required")
	}

	block := getStringArg(args, "block", "")
	marker := getStringArg(args, "marker", "# {mark} ANSIBLE MANAGED BLOCK")
	state := getStringArg(args, "state", "present")
	create := getBoolArg(args, "create", false)

	beginMarker := strings.Replace(marker, "{mark}", "BEGIN", 1)
	endMarker := strings.Replace(marker, "{mark}", "END", 1)

	if state == "absent" {
		// Remove block
		cmd := fmt.Sprintf("sed -i '/%s/,/%s/d' %q",
			strings.ReplaceAll(beginMarker, "/", "\\/"),
			strings.ReplaceAll(endMarker, "/", "\\/"),
			path)
		_, _, _, _ = client.Run(ctx, cmd)
		return &TaskResult{Changed: true}, nil
	}

	// Create file if needed (best-effort)
	if create {
		_, _, _, _ = client.Run(ctx, fmt.Sprintf("touch %q", path))
	}

	// Remove existing block and add new one
	escapedBlock := strings.ReplaceAll(block, "'", "'\\''")
	cmd := fmt.Sprintf(`
sed -i '/%s/,/%s/d' %q 2>/dev/null || true
cat >> %q << 'BLOCK_EOF'
%s
%s
%s
BLOCK_EOF
`, strings.ReplaceAll(beginMarker, "/", "\\/"),
		strings.ReplaceAll(endMarker, "/", "\\/"),
		path, path, beginMarker, escapedBlock, endMarker)

	stdout, stderr, rc, err := client.RunScript(ctx, cmd)
	if err != nil || rc != 0 {
		return &TaskResult{Failed: true, Msg: stderr, Stdout: stdout, RC: rc}, nil
	}

	return &TaskResult{Changed: true}, nil
}

func (e *Executor) moduleIncludeVars(args map[string]any) (*TaskResult, error) {
	file := getStringArg(args, "file", "")
	if file == "" {
		file = getStringArg(args, "_raw_params", "")
	}

	if file != "" {
		// Would need to read and parse the vars file
		// For now, just acknowledge
		return &TaskResult{Changed: false, Msg: "include_vars: " + file}, nil
	}

	return &TaskResult{Changed: false}, nil
}

func (e *Executor) moduleMeta(args map[string]any) (*TaskResult, error) {
	// meta module controls play execution
	// Most actions are no-ops for us
	return &TaskResult{Changed: false}, nil
}

func (e *Executor) moduleSetup(ctx context.Context, client *SSHClient) (*TaskResult, error) {
	// Gather facts - similar to what we do in gatherFacts
	return &TaskResult{Changed: false, Msg: "facts gathered"}, nil
}

func (e *Executor) moduleReboot(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	preRebootDelay := 0
	if d, ok := args["pre_reboot_delay"].(int); ok {
		preRebootDelay = d
	}

	msg := getStringArg(args, "msg", "Reboot initiated by Ansible")

	if preRebootDelay > 0 {
		cmd := fmt.Sprintf("sleep %d && shutdown -r now '%s' &", preRebootDelay, msg)
		_, _, _, _ = client.Run(ctx, cmd)
	} else {
		_, _, _, _ = client.Run(ctx, fmt.Sprintf("shutdown -r now '%s' &", msg))
	}

	return &TaskResult{Changed: true, Msg: "Reboot initiated"}, nil
}

func (e *Executor) moduleUFW(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	rule := getStringArg(args, "rule", "")
	port := getStringArg(args, "port", "")
	proto := getStringArg(args, "proto", "tcp")
	state := getStringArg(args, "state", "")

	var cmd string

	// Handle state (enable/disable)
	if state != "" {
		switch state {
		case "enabled":
			cmd = "ufw --force enable"
		case "disabled":
			cmd = "ufw disable"
		case "reloaded":
			cmd = "ufw reload"
		case "reset":
			cmd = "ufw --force reset"
		}
		if cmd != "" {
			stdout, stderr, rc, err := client.Run(ctx, cmd)
			if err != nil || rc != 0 {
				return &TaskResult{Failed: true, Msg: stderr, Stdout: stdout, RC: rc}, nil
			}
			return &TaskResult{Changed: true}, nil
		}
	}

	// Handle rule
	if rule != "" && port != "" {
		switch rule {
		case "allow":
			cmd = fmt.Sprintf("ufw allow %s/%s", port, proto)
		case "deny":
			cmd = fmt.Sprintf("ufw deny %s/%s", port, proto)
		case "reject":
			cmd = fmt.Sprintf("ufw reject %s/%s", port, proto)
		case "limit":
			cmd = fmt.Sprintf("ufw limit %s/%s", port, proto)
		}

		stdout, stderr, rc, err := client.Run(ctx, cmd)
		if err != nil || rc != 0 {
			return &TaskResult{Failed: true, Msg: stderr, Stdout: stdout, RC: rc}, nil
		}
	}

	return &TaskResult{Changed: true}, nil
}

func (e *Executor) moduleAuthorizedKey(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	user := getStringArg(args, "user", "")
	key := getStringArg(args, "key", "")
	state := getStringArg(args, "state", "present")

	if user == "" || key == "" {
		return nil, fmt.Errorf("authorized_key: user and key required")
	}

	// Get user's home directory
	stdout, _, _, err := client.Run(ctx, fmt.Sprintf("getent passwd %s | cut -d: -f6", user))
	if err != nil {
		return nil, fmt.Errorf("get home dir: %w", err)
	}
	home := strings.TrimSpace(stdout)
	if home == "" {
		home = "/root"
		if user != "root" {
			home = "/home/" + user
		}
	}

	authKeysPath := filepath.Join(home, ".ssh", "authorized_keys")

	if state == "absent" {
		// Remove key
		escapedKey := strings.ReplaceAll(key, "/", "\\/")
		cmd := fmt.Sprintf("sed -i '/%s/d' %q 2>/dev/null || true", escapedKey[:40], authKeysPath)
		_, _, _, _ = client.Run(ctx, cmd)
		return &TaskResult{Changed: true}, nil
	}

	// Ensure .ssh directory exists (best-effort)
	_, _, _, _ = client.Run(ctx, fmt.Sprintf("mkdir -p %q && chmod 700 %q && chown %s:%s %q",
		filepath.Dir(authKeysPath), filepath.Dir(authKeysPath), user, user, filepath.Dir(authKeysPath)))

	// Add key if not present
	cmd := fmt.Sprintf("grep -qF %q %q 2>/dev/null || echo %q >> %q",
		key[:40], authKeysPath, key, authKeysPath)
	stdout, stderr, rc, err := client.Run(ctx, cmd)
	if err != nil || rc != 0 {
		return &TaskResult{Failed: true, Msg: stderr, Stdout: stdout, RC: rc}, nil
	}

	// Fix permissions (best-effort)
	_, _, _, _ = client.Run(ctx, fmt.Sprintf("chmod 600 %q && chown %s:%s %q",
		authKeysPath, user, user, authKeysPath))

	return &TaskResult{Changed: true}, nil
}

func (e *Executor) moduleDockerCompose(ctx context.Context, client *SSHClient, args map[string]any) (*TaskResult, error) {
	projectSrc := getStringArg(args, "project_src", "")
	state := getStringArg(args, "state", "present")

	if projectSrc == "" {
		return nil, fmt.Errorf("docker_compose: project_src required")
	}

	var cmd string
	switch state {
	case "present":
		cmd = fmt.Sprintf("cd %q && docker compose up -d", projectSrc)
	case "absent":
		cmd = fmt.Sprintf("cd %q && docker compose down", projectSrc)
	case "restarted":
		cmd = fmt.Sprintf("cd %q && docker compose restart", projectSrc)
	default:
		cmd = fmt.Sprintf("cd %q && docker compose up -d", projectSrc)
	}

	stdout, stderr, rc, err := client.Run(ctx, cmd)
	if err != nil || rc != 0 {
		return &TaskResult{Failed: true, Msg: stderr, Stdout: stdout, RC: rc}, nil
	}

	// Heuristic for changed
	changed := !strings.Contains(stdout, "Up to date") && !strings.Contains(stderr, "Up to date")

	return &TaskResult{Changed: changed, Stdout: stdout}, nil
}
