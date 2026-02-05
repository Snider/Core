package ansible

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/host-uk/core/pkg/log"
)

// Executor runs Ansible playbooks.
type Executor struct {
	parser    *Parser
	inventory *Inventory
	vars      map[string]any
	facts     map[string]*Facts
	results   map[string]map[string]*TaskResult // host -> register_name -> result
	handlers  map[string][]Task
	notified  map[string]bool
	clients   map[string]*SSHClient
	mu        sync.RWMutex

	// Callbacks
	OnPlayStart func(play *Play)
	OnTaskStart func(host string, task *Task)
	OnTaskEnd   func(host string, task *Task, result *TaskResult)
	OnPlayEnd   func(play *Play)

	// Options
	Limit     string
	Tags      []string
	SkipTags  []string
	CheckMode bool
	Diff      bool
	Verbose   int
}

// NewExecutor creates a new playbook executor.
func NewExecutor(basePath string) *Executor {
	return &Executor{
		parser:   NewParser(basePath),
		vars:     make(map[string]any),
		facts:    make(map[string]*Facts),
		results:  make(map[string]map[string]*TaskResult),
		handlers: make(map[string][]Task),
		notified: make(map[string]bool),
		clients:  make(map[string]*SSHClient),
	}
}

// SetInventory loads inventory from a file.
func (e *Executor) SetInventory(path string) error {
	inv, err := e.parser.ParseInventory(path)
	if err != nil {
		return err
	}
	e.inventory = inv
	return nil
}

// SetInventoryDirect sets inventory directly.
func (e *Executor) SetInventoryDirect(inv *Inventory) {
	e.inventory = inv
}

// SetVar sets a variable.
func (e *Executor) SetVar(key string, value any) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.vars[key] = value
}

// Run executes a playbook.
func (e *Executor) Run(ctx context.Context, playbookPath string) error {
	plays, err := e.parser.ParsePlaybook(playbookPath)
	if err != nil {
		return fmt.Errorf("parse playbook: %w", err)
	}

	for i := range plays {
		if err := e.runPlay(ctx, &plays[i]); err != nil {
			return fmt.Errorf("play %d (%s): %w", i, plays[i].Name, err)
		}
	}

	return nil
}

// runPlay executes a single play.
func (e *Executor) runPlay(ctx context.Context, play *Play) error {
	if e.OnPlayStart != nil {
		e.OnPlayStart(play)
	}
	defer func() {
		if e.OnPlayEnd != nil {
			e.OnPlayEnd(play)
		}
	}()

	// Get target hosts
	hosts := e.getHosts(play.Hosts)
	if len(hosts) == 0 {
		return nil // No hosts matched
	}

	// Merge play vars
	for k, v := range play.Vars {
		e.vars[k] = v
	}

	// Gather facts if needed
	gatherFacts := play.GatherFacts == nil || *play.GatherFacts
	if gatherFacts {
		for _, host := range hosts {
			if err := e.gatherFacts(ctx, host, play); err != nil {
				// Non-fatal
				if e.Verbose > 0 {
					log.Warn("gather facts failed", "host", host, "err", err)
				}
			}
		}
	}

	// Execute pre_tasks
	for _, task := range play.PreTasks {
		if err := e.runTaskOnHosts(ctx, hosts, &task, play); err != nil {
			return err
		}
	}

	// Execute roles
	for _, roleRef := range play.Roles {
		if err := e.runRole(ctx, hosts, &roleRef, play); err != nil {
			return err
		}
	}

	// Execute tasks
	for _, task := range play.Tasks {
		if err := e.runTaskOnHosts(ctx, hosts, &task, play); err != nil {
			return err
		}
	}

	// Execute post_tasks
	for _, task := range play.PostTasks {
		if err := e.runTaskOnHosts(ctx, hosts, &task, play); err != nil {
			return err
		}
	}

	// Run notified handlers
	for _, handler := range play.Handlers {
		if e.notified[handler.Name] {
			if err := e.runTaskOnHosts(ctx, hosts, &handler, play); err != nil {
				return err
			}
		}
	}

	return nil
}

// runRole executes a role on hosts.
func (e *Executor) runRole(ctx context.Context, hosts []string, roleRef *RoleRef, play *Play) error {
	// Check when condition
	if roleRef.When != nil {
		if !e.evaluateWhen(roleRef.When, "", nil) {
			return nil
		}
	}

	// Parse role tasks
	tasks, err := e.parser.ParseRole(roleRef.Role, roleRef.TasksFrom)
	if err != nil {
		return log.E("executor.runRole", fmt.Sprintf("parse role %s", roleRef.Role), err)
	}

	// Merge role vars
	oldVars := make(map[string]any)
	for k, v := range e.vars {
		oldVars[k] = v
	}
	for k, v := range roleRef.Vars {
		e.vars[k] = v
	}

	// Execute tasks
	for _, task := range tasks {
		if err := e.runTaskOnHosts(ctx, hosts, &task, play); err != nil {
			// Restore vars
			e.vars = oldVars
			return err
		}
	}

	// Restore vars
	e.vars = oldVars
	return nil
}

// runTaskOnHosts runs a task on all hosts.
func (e *Executor) runTaskOnHosts(ctx context.Context, hosts []string, task *Task, play *Play) error {
	// Check tags
	if !e.matchesTags(task.Tags) {
		return nil
	}

	// Handle block tasks
	if len(task.Block) > 0 {
		return e.runBlock(ctx, hosts, task, play)
	}

	// Handle include/import
	if task.IncludeTasks != "" || task.ImportTasks != "" {
		return e.runIncludeTasks(ctx, hosts, task, play)
	}
	if task.IncludeRole != nil || task.ImportRole != nil {
		return e.runIncludeRole(ctx, hosts, task, play)
	}

	for _, host := range hosts {
		if err := e.runTaskOnHost(ctx, host, task, play); err != nil {
			if !task.IgnoreErrors {
				return err
			}
		}
	}

	return nil
}

// runTaskOnHost runs a task on a single host.
func (e *Executor) runTaskOnHost(ctx context.Context, host string, task *Task, play *Play) error {
	start := time.Now()

	if e.OnTaskStart != nil {
		e.OnTaskStart(host, task)
	}

	// Initialize host results
	if e.results[host] == nil {
		e.results[host] = make(map[string]*TaskResult)
	}

	// Check when condition
	if task.When != nil {
		if !e.evaluateWhen(task.When, host, task) {
			result := &TaskResult{Skipped: true, Msg: "Skipped due to when condition"}
			if task.Register != "" {
				e.results[host][task.Register] = result
			}
			if e.OnTaskEnd != nil {
				e.OnTaskEnd(host, task, result)
			}
			return nil
		}
	}

	// Get SSH client
	client, err := e.getClient(host, play)
	if err != nil {
		return fmt.Errorf("get client for %s: %w", host, err)
	}

	// Handle loops
	if task.Loop != nil {
		return e.runLoop(ctx, host, client, task, play)
	}

	// Execute the task
	result, err := e.executeModule(ctx, host, client, task, play)
	if err != nil {
		result = &TaskResult{Failed: true, Msg: err.Error()}
	}
	result.Duration = time.Since(start)

	// Store result
	if task.Register != "" {
		e.results[host][task.Register] = result
	}

	// Handle notify
	if result.Changed && task.Notify != nil {
		e.handleNotify(task.Notify)
	}

	if e.OnTaskEnd != nil {
		e.OnTaskEnd(host, task, result)
	}

	if result.Failed && !task.IgnoreErrors {
		return fmt.Errorf("task failed: %s", result.Msg)
	}

	return nil
}

// runLoop handles task loops.
func (e *Executor) runLoop(ctx context.Context, host string, client *SSHClient, task *Task, play *Play) error {
	items := e.resolveLoop(task.Loop, host)

	loopVar := "item"
	if task.LoopControl != nil && task.LoopControl.LoopVar != "" {
		loopVar = task.LoopControl.LoopVar
	}

	// Save loop state to restore after loop
	savedVars := make(map[string]any)
	if v, ok := e.vars[loopVar]; ok {
		savedVars[loopVar] = v
	}
	indexVar := ""
	if task.LoopControl != nil && task.LoopControl.IndexVar != "" {
		indexVar = task.LoopControl.IndexVar
		if v, ok := e.vars[indexVar]; ok {
			savedVars[indexVar] = v
		}
	}

	var results []TaskResult
	for i, item := range items {
		// Set loop variables
		e.vars[loopVar] = item
		if indexVar != "" {
			e.vars[indexVar] = i
		}

		result, err := e.executeModule(ctx, host, client, task, play)
		if err != nil {
			result = &TaskResult{Failed: true, Msg: err.Error()}
		}
		results = append(results, *result)

		if result.Failed && !task.IgnoreErrors {
			break
		}
	}

	// Restore loop variables
	if v, ok := savedVars[loopVar]; ok {
		e.vars[loopVar] = v
	} else {
		delete(e.vars, loopVar)
	}
	if indexVar != "" {
		if v, ok := savedVars[indexVar]; ok {
			e.vars[indexVar] = v
		} else {
			delete(e.vars, indexVar)
		}
	}

	// Store combined result
	if task.Register != "" {
		combined := &TaskResult{
			Results: results,
			Changed: false,
		}
		for _, r := range results {
			if r.Changed {
				combined.Changed = true
			}
			if r.Failed {
				combined.Failed = true
			}
		}
		e.results[host][task.Register] = combined
	}

	return nil
}

// runBlock handles block/rescue/always.
func (e *Executor) runBlock(ctx context.Context, hosts []string, task *Task, play *Play) error {
	var blockErr error

	// Try block
	for _, t := range task.Block {
		if err := e.runTaskOnHosts(ctx, hosts, &t, play); err != nil {
			blockErr = err
			break
		}
	}

	// Run rescue if block failed
	if blockErr != nil && len(task.Rescue) > 0 {
		for _, t := range task.Rescue {
			if err := e.runTaskOnHosts(ctx, hosts, &t, play); err != nil {
				// Rescue also failed
				break
			}
		}
	}

	// Always run always block
	for _, t := range task.Always {
		if err := e.runTaskOnHosts(ctx, hosts, &t, play); err != nil {
			if blockErr == nil {
				blockErr = err
			}
		}
	}

	if blockErr != nil && len(task.Rescue) == 0 {
		return blockErr
	}

	return nil
}

// runIncludeTasks handles include_tasks/import_tasks.
func (e *Executor) runIncludeTasks(ctx context.Context, hosts []string, task *Task, play *Play) error {
	path := task.IncludeTasks
	if path == "" {
		path = task.ImportTasks
	}

	// Resolve path relative to playbook
	path = e.templateString(path, "", nil)

	tasks, err := e.parser.ParseTasks(path)
	if err != nil {
		return fmt.Errorf("include_tasks %s: %w", path, err)
	}

	for _, t := range tasks {
		if err := e.runTaskOnHosts(ctx, hosts, &t, play); err != nil {
			return err
		}
	}

	return nil
}

// runIncludeRole handles include_role/import_role.
func (e *Executor) runIncludeRole(ctx context.Context, hosts []string, task *Task, play *Play) error {
	var roleName, tasksFrom string
	var roleVars map[string]any

	if task.IncludeRole != nil {
		roleName = task.IncludeRole.Name
		tasksFrom = task.IncludeRole.TasksFrom
		roleVars = task.IncludeRole.Vars
	} else {
		roleName = task.ImportRole.Name
		tasksFrom = task.ImportRole.TasksFrom
		roleVars = task.ImportRole.Vars
	}

	roleRef := &RoleRef{
		Role:      roleName,
		TasksFrom: tasksFrom,
		Vars:      roleVars,
	}

	return e.runRole(ctx, hosts, roleRef, play)
}

// getHosts returns hosts matching the pattern.
func (e *Executor) getHosts(pattern string) []string {
	if e.inventory == nil {
		if pattern == "localhost" {
			return []string{"localhost"}
		}
		return nil
	}

	hosts := GetHosts(e.inventory, pattern)

	// Apply limit - filter to hosts that are also in the limit group
	if e.Limit != "" {
		limitHosts := GetHosts(e.inventory, e.Limit)
		limitSet := make(map[string]bool)
		for _, h := range limitHosts {
			limitSet[h] = true
		}

		var filtered []string
		for _, h := range hosts {
			if limitSet[h] || h == e.Limit || strings.Contains(h, e.Limit) {
				filtered = append(filtered, h)
			}
		}
		hosts = filtered
	}

	return hosts
}

// getClient returns or creates an SSH client for a host.
func (e *Executor) getClient(host string, play *Play) (*SSHClient, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if client, ok := e.clients[host]; ok {
		return client, nil
	}

	// Get host vars
	vars := make(map[string]any)
	if e.inventory != nil {
		vars = GetHostVars(e.inventory, host)
	}

	// Merge with play vars
	for k, v := range e.vars {
		if _, exists := vars[k]; !exists {
			vars[k] = v
		}
	}

	// Build SSH config
	cfg := SSHConfig{
		Host: host,
		Port: 22,
		User: "root",
	}

	if h, ok := vars["ansible_host"].(string); ok {
		cfg.Host = h
	}
	if p, ok := vars["ansible_port"].(int); ok {
		cfg.Port = p
	}
	if u, ok := vars["ansible_user"].(string); ok {
		cfg.User = u
	}
	if p, ok := vars["ansible_password"].(string); ok {
		cfg.Password = p
	}
	if k, ok := vars["ansible_ssh_private_key_file"].(string); ok {
		cfg.KeyFile = k
	}

	// Apply play become settings
	if play.Become {
		cfg.Become = true
		cfg.BecomeUser = play.BecomeUser
		if bp, ok := vars["ansible_become_password"].(string); ok {
			cfg.BecomePass = bp
		} else if cfg.Password != "" {
			// Use SSH password for sudo if no become password specified
			cfg.BecomePass = cfg.Password
		}
	}

	client, err := NewSSHClient(cfg)
	if err != nil {
		return nil, err
	}

	e.clients[host] = client
	return client, nil
}

// gatherFacts collects facts from a host.
func (e *Executor) gatherFacts(ctx context.Context, host string, play *Play) error {
	if play.Connection == "local" || host == "localhost" {
		// Local facts
		e.facts[host] = &Facts{
			Hostname: "localhost",
		}
		return nil
	}

	client, err := e.getClient(host, play)
	if err != nil {
		return err
	}

	// Gather basic facts
	facts := &Facts{}

	// Hostname
	stdout, _, _, err := client.Run(ctx, "hostname -f 2>/dev/null || hostname")
	if err == nil {
		facts.FQDN = strings.TrimSpace(stdout)
	}

	stdout, _, _, err = client.Run(ctx, "hostname -s 2>/dev/null || hostname")
	if err == nil {
		facts.Hostname = strings.TrimSpace(stdout)
	}

	// OS info
	stdout, _, _, _ = client.Run(ctx, "cat /etc/os-release 2>/dev/null | grep -E '^(ID|VERSION_ID)=' | head -2")
	for _, line := range strings.Split(stdout, "\n") {
		if strings.HasPrefix(line, "ID=") {
			facts.Distribution = strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
		}
		if strings.HasPrefix(line, "VERSION_ID=") {
			facts.Version = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
		}
	}

	// Architecture
	stdout, _, _, _ = client.Run(ctx, "uname -m")
	facts.Architecture = strings.TrimSpace(stdout)

	// Kernel
	stdout, _, _, _ = client.Run(ctx, "uname -r")
	facts.Kernel = strings.TrimSpace(stdout)

	e.mu.Lock()
	e.facts[host] = facts
	e.mu.Unlock()

	return nil
}

// evaluateWhen evaluates a when condition.
func (e *Executor) evaluateWhen(when any, host string, task *Task) bool {
	conditions := normalizeConditions(when)

	for _, cond := range conditions {
		cond = e.templateString(cond, host, task)
		if !e.evalCondition(cond, host) {
			return false
		}
	}

	return true
}

func normalizeConditions(when any) []string {
	switch v := when.(type) {
	case string:
		return []string{v}
	case []any:
		var conds []string
		for _, c := range v {
			if s, ok := c.(string); ok {
				conds = append(conds, s)
			}
		}
		return conds
	case []string:
		return v
	}
	return nil
}

// evalCondition evaluates a single condition.
func (e *Executor) evalCondition(cond string, host string) bool {
	cond = strings.TrimSpace(cond)

	// Handle negation
	if strings.HasPrefix(cond, "not ") {
		return !e.evalCondition(strings.TrimPrefix(cond, "not "), host)
	}

	// Handle boolean literals
	if cond == "true" || cond == "True" {
		return true
	}
	if cond == "false" || cond == "False" {
		return false
	}

	// Handle registered variable checks
	// e.g., "result is success", "result.rc == 0"
	if strings.Contains(cond, " is ") {
		parts := strings.SplitN(cond, " is ", 2)
		varName := strings.TrimSpace(parts[0])
		check := strings.TrimSpace(parts[1])

		result := e.getRegisteredVar(host, varName)
		if result == nil {
			return check == "not defined" || check == "undefined"
		}

		switch check {
		case "defined":
			return true
		case "not defined", "undefined":
			return false
		case "success", "succeeded":
			return !result.Failed
		case "failed":
			return result.Failed
		case "changed":
			return result.Changed
		case "skipped":
			return result.Skipped
		}
	}

	// Handle simple var checks
	if strings.Contains(cond, " | default(") {
		// Extract var name and check if defined
		re := regexp.MustCompile(`(\w+)\s*\|\s*default\([^)]*\)`)
		if match := re.FindStringSubmatch(cond); len(match) > 1 {
			// Has default, so condition is satisfied
			return true
		}
	}

	// Check if it's a variable that should be truthy
	if result := e.getRegisteredVar(host, cond); result != nil {
		return !result.Failed && !result.Skipped
	}

	// Check vars
	if val, ok := e.vars[cond]; ok {
		switch v := val.(type) {
		case bool:
			return v
		case string:
			return v != "" && v != "false" && v != "False"
		case int:
			return v != 0
		}
	}

	// Default to true for unknown conditions (be permissive)
	return true
}

// getRegisteredVar gets a registered task result.
func (e *Executor) getRegisteredVar(host string, name string) *TaskResult {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Handle dotted access (e.g., "result.stdout")
	parts := strings.SplitN(name, ".", 2)
	varName := parts[0]

	if hostResults, ok := e.results[host]; ok {
		if result, ok := hostResults[varName]; ok {
			return result
		}
	}

	return nil
}

// templateString applies Jinja2-like templating.
func (e *Executor) templateString(s string, host string, task *Task) string {
	// Handle {{ var }} syntax
	re := regexp.MustCompile(`\{\{\s*([^}]+)\s*\}\}`)

	return re.ReplaceAllStringFunc(s, func(match string) string {
		expr := strings.TrimSpace(match[2 : len(match)-2])
		return e.resolveExpr(expr, host, task)
	})
}

// resolveExpr resolves a template expression.
func (e *Executor) resolveExpr(expr string, host string, task *Task) string {
	// Handle filters
	if strings.Contains(expr, " | ") {
		parts := strings.SplitN(expr, " | ", 2)
		value := e.resolveExpr(parts[0], host, task)
		return e.applyFilter(value, parts[1])
	}

	// Handle lookups
	if strings.HasPrefix(expr, "lookup(") {
		return e.handleLookup(expr)
	}

	// Handle registered vars
	if strings.Contains(expr, ".") {
		parts := strings.SplitN(expr, ".", 2)
		if result := e.getRegisteredVar(host, parts[0]); result != nil {
			switch parts[1] {
			case "stdout":
				return result.Stdout
			case "stderr":
				return result.Stderr
			case "rc":
				return fmt.Sprintf("%d", result.RC)
			case "changed":
				return fmt.Sprintf("%t", result.Changed)
			case "failed":
				return fmt.Sprintf("%t", result.Failed)
			}
		}
	}

	// Check vars
	if val, ok := e.vars[expr]; ok {
		return fmt.Sprintf("%v", val)
	}

	// Check task vars
	if task != nil {
		if val, ok := task.Vars[expr]; ok {
			return fmt.Sprintf("%v", val)
		}
	}

	// Check host vars
	if e.inventory != nil {
		hostVars := GetHostVars(e.inventory, host)
		if val, ok := hostVars[expr]; ok {
			return fmt.Sprintf("%v", val)
		}
	}

	// Check facts
	if facts, ok := e.facts[host]; ok {
		switch expr {
		case "ansible_hostname":
			return facts.Hostname
		case "ansible_fqdn":
			return facts.FQDN
		case "ansible_distribution":
			return facts.Distribution
		case "ansible_distribution_version":
			return facts.Version
		case "ansible_architecture":
			return facts.Architecture
		case "ansible_kernel":
			return facts.Kernel
		}
	}

	return "{{ " + expr + " }}" // Return as-is if unresolved
}

// applyFilter applies a Jinja2 filter.
func (e *Executor) applyFilter(value, filter string) string {
	filter = strings.TrimSpace(filter)

	// Handle default filter
	if strings.HasPrefix(filter, "default(") {
		if value == "" || value == "{{ "+filter+" }}" {
			// Extract default value
			re := regexp.MustCompile(`default\(([^)]*)\)`)
			if match := re.FindStringSubmatch(filter); len(match) > 1 {
				return strings.Trim(match[1], "'\"")
			}
		}
		return value
	}

	// Handle bool filter
	if filter == "bool" {
		lower := strings.ToLower(value)
		if lower == "true" || lower == "yes" || lower == "1" {
			return "true"
		}
		return "false"
	}

	// Handle trim
	if filter == "trim" {
		return strings.TrimSpace(value)
	}

	// Handle b64decode
	if filter == "b64decode" {
		// Would need base64 decode
		return value
	}

	return value
}

// handleLookup handles lookup() expressions.
func (e *Executor) handleLookup(expr string) string {
	// Parse lookup('type', 'arg')
	re := regexp.MustCompile(`lookup\s*\(\s*['"](\w+)['"]\s*,\s*['"]([^'"]+)['"]\s*`)
	match := re.FindStringSubmatch(expr)
	if len(match) < 3 {
		return ""
	}

	lookupType := match[1]
	arg := match[2]

	switch lookupType {
	case "env":
		return os.Getenv(arg)
	case "file":
		if data, err := os.ReadFile(arg); err == nil {
			return string(data)
		}
	}

	return ""
}

// resolveLoop resolves loop items.
func (e *Executor) resolveLoop(loop any, host string) []any {
	switch v := loop.(type) {
	case []any:
		return v
	case []string:
		items := make([]any, len(v))
		for i, s := range v {
			items[i] = s
		}
		return items
	case string:
		// Template the string and see if it's a var reference
		resolved := e.templateString(v, host, nil)
		if val, ok := e.vars[resolved]; ok {
			if items, ok := val.([]any); ok {
				return items
			}
		}
	}
	return nil
}

// matchesTags checks if task tags match execution tags.
func (e *Executor) matchesTags(taskTags []string) bool {
	// If no tags specified, run all
	if len(e.Tags) == 0 && len(e.SkipTags) == 0 {
		return true
	}

	// Check skip tags
	for _, skip := range e.SkipTags {
		for _, tt := range taskTags {
			if skip == tt {
				return false
			}
		}
	}

	// Check include tags
	if len(e.Tags) > 0 {
		for _, tag := range e.Tags {
			for _, tt := range taskTags {
				if tag == tt || tag == "all" {
					return true
				}
			}
		}
		return false
	}

	return true
}

// handleNotify marks handlers as notified.
func (e *Executor) handleNotify(notify any) {
	switch v := notify.(type) {
	case string:
		e.notified[v] = true
	case []any:
		for _, n := range v {
			if s, ok := n.(string); ok {
				e.notified[s] = true
			}
		}
	case []string:
		for _, s := range v {
			e.notified[s] = true
		}
	}
}

// Close closes all SSH connections.
func (e *Executor) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, client := range e.clients {
		_ = client.Close()
	}
	e.clients = make(map[string]*SSHClient)
}

// TemplateFile processes a template file.
func (e *Executor) TemplateFile(src, host string, task *Task) (string, error) {
	content, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}

	// Convert Jinja2 to Go template syntax (basic conversion)
	tmplContent := string(content)
	tmplContent = strings.ReplaceAll(tmplContent, "{{", "{{ .")
	tmplContent = strings.ReplaceAll(tmplContent, "{%", "{{")
	tmplContent = strings.ReplaceAll(tmplContent, "%}", "}}")

	tmpl, err := template.New("template").Parse(tmplContent)
	if err != nil {
		// Fall back to simple replacement
		return e.templateString(string(content), host, task), nil
	}

	// Build context map
	context := make(map[string]any)
	for k, v := range e.vars {
		context[k] = v
	}
	// Add host vars
	if e.inventory != nil {
		hostVars := GetHostVars(e.inventory, host)
		for k, v := range hostVars {
			context[k] = v
		}
	}
	// Add facts
	if facts, ok := e.facts[host]; ok {
		context["ansible_hostname"] = facts.Hostname
		context["ansible_fqdn"] = facts.FQDN
		context["ansible_distribution"] = facts.Distribution
		context["ansible_distribution_version"] = facts.Version
		context["ansible_architecture"] = facts.Architecture
		context["ansible_kernel"] = facts.Kernel
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, context); err != nil {
		return e.templateString(string(content), host, task), nil
	}

	return buf.String(), nil
}
