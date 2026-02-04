package ansible

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/log"
	"gopkg.in/yaml.v3"
)

// Parser handles Ansible YAML parsing.
type Parser struct {
	basePath string
	vars     map[string]any
}

// NewParser creates a new Ansible parser.
func NewParser(basePath string) *Parser {
	return &Parser{
		basePath: basePath,
		vars:     make(map[string]any),
	}
}

// ParsePlaybook parses an Ansible playbook file.
func (p *Parser) ParsePlaybook(path string) ([]Play, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read playbook: %w", err)
	}

	var plays []Play
	if err := yaml.Unmarshal(data, &plays); err != nil {
		return nil, fmt.Errorf("parse playbook: %w", err)
	}

	// Process each play
	for i := range plays {
		if err := p.processPlay(&plays[i]); err != nil {
			return nil, fmt.Errorf("process play %d: %w", i, err)
		}
	}

	return plays, nil
}

// ParseInventory parses an Ansible inventory file.
func (p *Parser) ParseInventory(path string) (*Inventory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read inventory: %w", err)
	}

	var inv Inventory
	if err := yaml.Unmarshal(data, &inv); err != nil {
		return nil, fmt.Errorf("parse inventory: %w", err)
	}

	return &inv, nil
}

// ParseTasks parses a tasks file (used by include_tasks).
func (p *Parser) ParseTasks(path string) ([]Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read tasks: %w", err)
	}

	var tasks []Task
	if err := yaml.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("parse tasks: %w", err)
	}

	for i := range tasks {
		if err := p.extractModule(&tasks[i]); err != nil {
			return nil, fmt.Errorf("task %d: %w", i, err)
		}
	}

	return tasks, nil
}

// ParseRole parses a role and returns its tasks.
func (p *Parser) ParseRole(name string, tasksFrom string) ([]Task, error) {
	if tasksFrom == "" {
		tasksFrom = "main.yml"
	}

	// Search paths for roles (in order of precedence)
	searchPaths := []string{
		// Relative to playbook
		filepath.Join(p.basePath, "roles", name, "tasks", tasksFrom),
		// Parent directory roles
		filepath.Join(filepath.Dir(p.basePath), "roles", name, "tasks", tasksFrom),
		// Sibling roles directory
		filepath.Join(p.basePath, "..", "roles", name, "tasks", tasksFrom),
		// playbooks/roles pattern
		filepath.Join(p.basePath, "playbooks", "roles", name, "tasks", tasksFrom),
		// Common DevOps structure
		filepath.Join(filepath.Dir(filepath.Dir(p.basePath)), "roles", name, "tasks", tasksFrom),
	}

	var tasksPath string
	for _, sp := range searchPaths {
		// Clean the path to resolve .. segments
		sp = filepath.Clean(sp)
		if _, err := os.Stat(sp); err == nil {
			tasksPath = sp
			break
		}
	}

	if tasksPath == "" {
		return nil, log.E("parser.ParseRole", fmt.Sprintf("role %s not found in search paths: %v", name, searchPaths), nil)
	}

	// Load role defaults
	defaultsPath := filepath.Join(filepath.Dir(filepath.Dir(tasksPath)), "defaults", "main.yml")
	if data, err := os.ReadFile(defaultsPath); err == nil {
		var defaults map[string]any
		if yaml.Unmarshal(data, &defaults) == nil {
			for k, v := range defaults {
				if _, exists := p.vars[k]; !exists {
					p.vars[k] = v
				}
			}
		}
	}

	// Load role vars
	varsPath := filepath.Join(filepath.Dir(filepath.Dir(tasksPath)), "vars", "main.yml")
	if data, err := os.ReadFile(varsPath); err == nil {
		var roleVars map[string]any
		if yaml.Unmarshal(data, &roleVars) == nil {
			for k, v := range roleVars {
				p.vars[k] = v
			}
		}
	}

	return p.ParseTasks(tasksPath)
}

// processPlay processes a play and extracts modules from tasks.
func (p *Parser) processPlay(play *Play) error {
	// Merge play vars
	for k, v := range play.Vars {
		p.vars[k] = v
	}

	for i := range play.PreTasks {
		if err := p.extractModule(&play.PreTasks[i]); err != nil {
			return fmt.Errorf("pre_task %d: %w", i, err)
		}
	}

	for i := range play.Tasks {
		if err := p.extractModule(&play.Tasks[i]); err != nil {
			return fmt.Errorf("task %d: %w", i, err)
		}
	}

	for i := range play.PostTasks {
		if err := p.extractModule(&play.PostTasks[i]); err != nil {
			return fmt.Errorf("post_task %d: %w", i, err)
		}
	}

	for i := range play.Handlers {
		if err := p.extractModule(&play.Handlers[i]); err != nil {
			return fmt.Errorf("handler %d: %w", i, err)
		}
	}

	return nil
}

// extractModule extracts the module name and args from a task.
func (p *Parser) extractModule(task *Task) error {
	// First, unmarshal the raw YAML to get all keys
	// This is a workaround since we need to find the module key dynamically

	// Handle block tasks
	for i := range task.Block {
		if err := p.extractModule(&task.Block[i]); err != nil {
			return err
		}
	}
	for i := range task.Rescue {
		if err := p.extractModule(&task.Rescue[i]); err != nil {
			return err
		}
	}
	for i := range task.Always {
		if err := p.extractModule(&task.Always[i]); err != nil {
			return err
		}
	}

	return nil
}

// UnmarshalYAML implements custom YAML unmarshaling for Task.
func (t *Task) UnmarshalYAML(node *yaml.Node) error {
	// First decode known fields
	type rawTask Task
	var raw rawTask

	// Create a map to capture all fields
	var m map[string]any
	if err := node.Decode(&m); err != nil {
		return err
	}

	// Decode into struct
	if err := node.Decode(&raw); err != nil {
		return err
	}
	*t = Task(raw)
	t.raw = m

	// Find the module key
	knownKeys := map[string]bool{
		"name": true, "register": true, "when": true, "loop": true,
		"loop_control": true, "vars": true, "environment": true,
		"changed_when": true, "failed_when": true, "ignore_errors": true,
		"no_log": true, "become": true, "become_user": true,
		"delegate_to": true, "run_once": true, "tags": true,
		"block": true, "rescue": true, "always": true, "notify": true,
		"retries": true, "delay": true, "until": true,
		"include_tasks": true, "import_tasks": true,
		"include_role": true, "import_role": true,
		"with_items": true, "with_dict": true, "with_file": true,
	}

	for key, val := range m {
		if knownKeys[key] {
			continue
		}

		// Check if this is a module
		if isModule(key) {
			t.Module = key
			t.Args = make(map[string]any)

			switch v := val.(type) {
			case string:
				// Free-form args (e.g., shell: echo hello)
				t.Args["_raw_params"] = v
			case map[string]any:
				t.Args = v
			case nil:
				// Module with no args
			default:
				t.Args["_raw_params"] = v
			}
			break
		}
	}

	// Handle with_items as loop
	if items, ok := m["with_items"]; ok && t.Loop == nil {
		t.Loop = items
	}

	return nil
}

// isModule checks if a key is a known module.
func isModule(key string) bool {
	for _, m := range KnownModules {
		if key == m {
			return true
		}
		// Also check without ansible.builtin. prefix
		if strings.HasPrefix(m, "ansible.builtin.") {
			if key == strings.TrimPrefix(m, "ansible.builtin.") {
				return true
			}
		}
	}
	// Accept any key with dots (likely a module)
	return strings.Contains(key, ".")
}

// NormalizeModule normalizes a module name to its canonical form.
func NormalizeModule(name string) string {
	// Add ansible.builtin. prefix if missing
	if !strings.Contains(name, ".") {
		return "ansible.builtin." + name
	}
	return name
}

// GetHosts returns hosts matching a pattern from inventory.
func GetHosts(inv *Inventory, pattern string) []string {
	if pattern == "all" {
		return getAllHosts(inv.All)
	}
	if pattern == "localhost" {
		return []string{"localhost"}
	}

	// Check if it's a group name
	hosts := getGroupHosts(inv.All, pattern)
	if len(hosts) > 0 {
		return hosts
	}

	// Check if it's a specific host
	if hasHost(inv.All, pattern) {
		return []string{pattern}
	}

	// Handle patterns with : (intersection/union)
	// For now, just return empty
	return nil
}

func getAllHosts(group *InventoryGroup) []string {
	if group == nil {
		return nil
	}

	var hosts []string
	for name := range group.Hosts {
		hosts = append(hosts, name)
	}
	for _, child := range group.Children {
		hosts = append(hosts, getAllHosts(child)...)
	}
	return hosts
}

func getGroupHosts(group *InventoryGroup, name string) []string {
	if group == nil {
		return nil
	}

	// Check children for the group name
	if child, ok := group.Children[name]; ok {
		return getAllHosts(child)
	}

	// Recurse
	for _, child := range group.Children {
		if hosts := getGroupHosts(child, name); len(hosts) > 0 {
			return hosts
		}
	}

	return nil
}

func hasHost(group *InventoryGroup, name string) bool {
	if group == nil {
		return false
	}

	if _, ok := group.Hosts[name]; ok {
		return true
	}

	for _, child := range group.Children {
		if hasHost(child, name) {
			return true
		}
	}

	return false
}

// GetHostVars returns variables for a specific host.
func GetHostVars(inv *Inventory, hostname string) map[string]any {
	vars := make(map[string]any)

	// Collect vars from all levels
	collectHostVars(inv.All, hostname, vars)

	return vars
}

func collectHostVars(group *InventoryGroup, hostname string, vars map[string]any) bool {
	if group == nil {
		return false
	}

	// Check if host is in this group
	found := false
	if host, ok := group.Hosts[hostname]; ok {
		found = true
		// Apply group vars first
		for k, v := range group.Vars {
			vars[k] = v
		}
		// Then host vars
		if host != nil {
			if host.AnsibleHost != "" {
				vars["ansible_host"] = host.AnsibleHost
			}
			if host.AnsiblePort != 0 {
				vars["ansible_port"] = host.AnsiblePort
			}
			if host.AnsibleUser != "" {
				vars["ansible_user"] = host.AnsibleUser
			}
			if host.AnsiblePassword != "" {
				vars["ansible_password"] = host.AnsiblePassword
			}
			if host.AnsibleSSHPrivateKeyFile != "" {
				vars["ansible_ssh_private_key_file"] = host.AnsibleSSHPrivateKeyFile
			}
			if host.AnsibleConnection != "" {
				vars["ansible_connection"] = host.AnsibleConnection
			}
			for k, v := range host.Vars {
				vars[k] = v
			}
		}
	}

	// Check children
	for _, child := range group.Children {
		if collectHostVars(child, hostname, vars) {
			// Apply this group's vars (parent vars)
			for k, v := range group.Vars {
				if _, exists := vars[k]; !exists {
					vars[k] = v
				}
			}
			found = true
		}
	}

	return found
}
