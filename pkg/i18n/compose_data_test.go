// Package i18n provides internationalization for the CLI.
package i18n

import (
	"sync"
)

// coreIntents defines the built-in semantic intents for common operations.
// These are accessed via the "core.*" namespace in T() and C() calls.
//
// Each intent provides templates for all output forms:
//   - Question: Initial prompt to the user
//   - Confirm: Secondary confirmation (for dangerous actions)
//   - Success: Message shown on successful completion
//   - Failure: Message shown on failure
//
// Templates use Go text/template syntax with the following data available:
//   - .Subject: Display value of the subject
//   - .Noun: The noun type (e.g., "file", "repo")
//   - .Count: Count for pluralization
//   - .Location: Location context
//
// Template functions available:
//   - title, lower, upper: Case transformations
//   - past, gerund: Verb conjugations
//   - plural, pluralForm: Noun pluralization
//   - article: Indefinite article selection (a/an)
//   - quote: Wrap in double quotes
var coreIntents = map[string]Intent{
	// --- Destructive Actions ---

	"core.delete": {
		Meta: IntentMeta{
			Type:      "action",
			Verb:      "delete",
			Dangerous: true,
			Default:   "no",
		},
		Question: "Delete {{.Subject}}?",
		Confirm:  "Really delete {{.Subject}}? This cannot be undone.",
		Success:  "{{.Subject | title}} deleted",
		Failure:  "Failed to delete {{.Subject}}",
	},

	"core.remove": {
		Meta: IntentMeta{
			Type:      "action",
			Verb:      "remove",
			Dangerous: true,
			Default:   "no",
		},
		Question: "Remove {{.Subject}}?",
		Confirm:  "Really remove {{.Subject}}?",
		Success:  "{{.Subject | title}} removed",
		Failure:  "Failed to remove {{.Subject}}",
	},

	"core.discard": {
		Meta: IntentMeta{
			Type:      "action",
			Verb:      "discard",
			Dangerous: true,
			Default:   "no",
		},
		Question: "Discard {{.Subject}}?",
		Confirm:  "Really discard {{.Subject}}? All changes will be lost.",
		Success:  "{{.Subject | title}} discarded",
		Failure:  "Failed to discard {{.Subject}}",
	},

	"core.reset": {
		Meta: IntentMeta{
			Type:      "action",
			Verb:      "reset",
			Dangerous: true,
			Default:   "no",
		},
		Question: "Reset {{.Subject}}?",
		Confirm:  "Really reset {{.Subject}}? This cannot be undone.",
		Success:  "{{.Subject | title}} reset",
		Failure:  "Failed to reset {{.Subject}}",
	},

	"core.overwrite": {
		Meta: IntentMeta{
			Type:      "action",
			Verb:      "overwrite",
			Dangerous: true,
			Default:   "no",
		},
		Question: "Overwrite {{.Subject}}?",
		Confirm:  "Really overwrite {{.Subject}}? Existing content will be lost.",
		Success:  "{{.Subject | title}} overwritten",
		Failure:  "Failed to overwrite {{.Subject}}",
	},

	// --- Creation Actions ---

	"core.create": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "create",
			Default: "yes",
		},
		Question: "Create {{.Subject}}?",
		Confirm:  "Create {{.Subject}}?",
		Success:  "{{.Subject | title}} created",
		Failure:  "Failed to create {{.Subject}}",
	},

	"core.add": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "add",
			Default: "yes",
		},
		Question: "Add {{.Subject}}?",
		Confirm:  "Add {{.Subject}}?",
		Success:  "{{.Subject | title}} added",
		Failure:  "Failed to add {{.Subject}}",
	},

	"core.clone": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "clone",
			Default: "yes",
		},
		Question: "Clone {{.Subject}}?",
		Confirm:  "Clone {{.Subject}}?",
		Success:  "{{.Subject | title}} cloned",
		Failure:  "Failed to clone {{.Subject}}",
	},

	"core.copy": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "copy",
			Default: "yes",
		},
		Question: "Copy {{.Subject}}?",
		Confirm:  "Copy {{.Subject}}?",
		Success:  "{{.Subject | title}} copied",
		Failure:  "Failed to copy {{.Subject}}",
	},

	// --- Modification Actions ---

	"core.save": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "save",
			Default: "yes",
		},
		Question: "Save {{.Subject}}?",
		Confirm:  "Save {{.Subject}}?",
		Success:  "{{.Subject | title}} saved",
		Failure:  "Failed to save {{.Subject}}",
	},

	"core.update": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "update",
			Default: "yes",
		},
		Question: "Update {{.Subject}}?",
		Confirm:  "Update {{.Subject}}?",
		Success:  "{{.Subject | title}} updated",
		Failure:  "Failed to update {{.Subject}}",
	},

	"core.rename": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "rename",
			Default: "yes",
		},
		Question: "Rename {{.Subject}}?",
		Confirm:  "Rename {{.Subject}}?",
		Success:  "{{.Subject | title}} renamed",
		Failure:  "Failed to rename {{.Subject}}",
	},

	"core.move": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "move",
			Default: "yes",
		},
		Question: "Move {{.Subject}}?",
		Confirm:  "Move {{.Subject}}?",
		Success:  "{{.Subject | title}} moved",
		Failure:  "Failed to move {{.Subject}}",
	},

	// --- Git Actions ---

	"core.commit": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "commit",
			Default: "yes",
		},
		Question: "Commit {{.Subject}}?",
		Confirm:  "Commit {{.Subject}}?",
		Success:  "{{.Subject | title}} committed",
		Failure:  "Failed to commit {{.Subject}}",
	},

	"core.push": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "push",
			Default: "yes",
		},
		Question: "Push {{.Subject}}?",
		Confirm:  "Push {{.Subject}}?",
		Success:  "{{.Subject | title}} pushed",
		Failure:  "Failed to push {{.Subject}}",
	},

	"core.pull": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "pull",
			Default: "yes",
		},
		Question: "Pull {{.Subject}}?",
		Confirm:  "Pull {{.Subject}}?",
		Success:  "{{.Subject | title}} pulled",
		Failure:  "Failed to pull {{.Subject}}",
	},

	"core.merge": {
		Meta: IntentMeta{
			Type:      "action",
			Verb:      "merge",
			Dangerous: true,
			Default:   "no",
		},
		Question: "Merge {{.Subject}}?",
		Confirm:  "Really merge {{.Subject}}?",
		Success:  "{{.Subject | title}} merged",
		Failure:  "Failed to merge {{.Subject}}",
	},

	"core.rebase": {
		Meta: IntentMeta{
			Type:      "action",
			Verb:      "rebase",
			Dangerous: true,
			Default:   "no",
		},
		Question: "Rebase {{.Subject}}?",
		Confirm:  "Really rebase {{.Subject}}? This rewrites history.",
		Success:  "{{.Subject | title}} rebased",
		Failure:  "Failed to rebase {{.Subject}}",
	},

	// --- Network Actions ---

	"core.install": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "install",
			Default: "yes",
		},
		Question: "Install {{.Subject}}?",
		Confirm:  "Install {{.Subject}}?",
		Success:  "{{.Subject | title}} installed",
		Failure:  "Failed to install {{.Subject}}",
	},

	"core.download": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "download",
			Default: "yes",
		},
		Question: "Download {{.Subject}}?",
		Confirm:  "Download {{.Subject}}?",
		Success:  "{{.Subject | title}} downloaded",
		Failure:  "Failed to download {{.Subject}}",
	},

	"core.upload": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "upload",
			Default: "yes",
		},
		Question: "Upload {{.Subject}}?",
		Confirm:  "Upload {{.Subject}}?",
		Success:  "{{.Subject | title}} uploaded",
		Failure:  "Failed to upload {{.Subject}}",
	},

	"core.publish": {
		Meta: IntentMeta{
			Type:      "action",
			Verb:      "publish",
			Dangerous: true,
			Default:   "no",
		},
		Question: "Publish {{.Subject}}?",
		Confirm:  "Really publish {{.Subject}}? This will be publicly visible.",
		Success:  "{{.Subject | title}} published",
		Failure:  "Failed to publish {{.Subject}}",
	},

	"core.deploy": {
		Meta: IntentMeta{
			Type:      "action",
			Verb:      "deploy",
			Dangerous: true,
			Default:   "no",
		},
		Question: "Deploy {{.Subject}}?",
		Confirm:  "Really deploy {{.Subject}}?",
		Success:  "{{.Subject | title}} deployed",
		Failure:  "Failed to deploy {{.Subject}}",
	},

	// --- Process Actions ---

	"core.start": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "start",
			Default: "yes",
		},
		Question: "Start {{.Subject}}?",
		Confirm:  "Start {{.Subject}}?",
		Success:  "{{.Subject | title}} started",
		Failure:  "Failed to start {{.Subject}}",
	},

	"core.stop": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "stop",
			Default: "yes",
		},
		Question: "Stop {{.Subject}}?",
		Confirm:  "Stop {{.Subject}}?",
		Success:  "{{.Subject | title}} stopped",
		Failure:  "Failed to stop {{.Subject}}",
	},

	"core.restart": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "restart",
			Default: "yes",
		},
		Question: "Restart {{.Subject}}?",
		Confirm:  "Restart {{.Subject}}?",
		Success:  "{{.Subject | title}} restarted",
		Failure:  "Failed to restart {{.Subject}}",
	},

	"core.run": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "run",
			Default: "yes",
		},
		Question: "Run {{.Subject}}?",
		Confirm:  "Run {{.Subject}}?",
		Success:  "{{.Subject | title}} completed",
		Failure:  "Failed to run {{.Subject}}",
	},

	"core.build": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "build",
			Default: "yes",
		},
		Question: "Build {{.Subject}}?",
		Confirm:  "Build {{.Subject}}?",
		Success:  "{{.Subject | title}} built",
		Failure:  "Failed to build {{.Subject}}",
	},

	"core.test": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "test",
			Default: "yes",
		},
		Question: "Test {{.Subject}}?",
		Confirm:  "Test {{.Subject}}?",
		Success:  "{{.Subject | title}} passed",
		Failure:  "{{.Subject | title}} failed",
	},

	// --- Information Actions ---

	"core.continue": {
		Meta: IntentMeta{
			Type:    "question",
			Verb:    "continue",
			Default: "yes",
		},
		Question: "Continue?",
		Confirm:  "Continue?",
		Success:  "Continuing",
		Failure:  "Aborted",
	},

	"core.proceed": {
		Meta: IntentMeta{
			Type:    "question",
			Verb:    "proceed",
			Default: "yes",
		},
		Question: "Proceed?",
		Confirm:  "Proceed?",
		Success:  "Proceeding",
		Failure:  "Aborted",
	},

	"core.confirm": {
		Meta: IntentMeta{
			Type:    "question",
			Verb:    "confirm",
			Default: "no",
		},
		Question: "Are you sure?",
		Confirm:  "Are you sure?",
		Success:  "Confirmed",
		Failure:  "Cancelled",
	},

	// --- Additional Actions ---

	"core.sync": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "sync",
			Default: "yes",
		},
		Question: "Sync {{.Subject}}?",
		Confirm:  "Sync {{.Subject}}?",
		Success:  "{{.Subject | title}} synced",
		Failure:  "Failed to sync {{.Subject}}",
	},

	"core.boot": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "boot",
			Default: "yes",
		},
		Question: "Boot {{.Subject}}?",
		Confirm:  "Boot {{.Subject}}?",
		Success:  "{{.Subject | title}} booted",
		Failure:  "Failed to boot {{.Subject}}",
	},

	"core.format": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "format",
			Default: "yes",
		},
		Question: "Format {{.Subject}}?",
		Confirm:  "Format {{.Subject}}?",
		Success:  "{{.Subject | title}} formatted",
		Failure:  "Failed to format {{.Subject}}",
	},

	"core.analyse": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "analyse",
			Default: "yes",
		},
		Question: "Analyse {{.Subject}}?",
		Confirm:  "Analyse {{.Subject}}?",
		Success:  "{{.Subject | title}} analysed",
		Failure:  "Failed to analyse {{.Subject}}",
	},

	"core.link": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "link",
			Default: "yes",
		},
		Question: "Link {{.Subject}}?",
		Confirm:  "Link {{.Subject}}?",
		Success:  "{{.Subject | title}} linked",
		Failure:  "Failed to link {{.Subject}}",
	},

	"core.unlink": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "unlink",
			Default: "yes",
		},
		Question: "Unlink {{.Subject}}?",
		Confirm:  "Unlink {{.Subject}}?",
		Success:  "{{.Subject | title}} unlinked",
		Failure:  "Failed to unlink {{.Subject}}",
	},

	"core.fetch": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "fetch",
			Default: "yes",
		},
		Question: "Fetch {{.Subject}}?",
		Confirm:  "Fetch {{.Subject}}?",
		Success:  "{{.Subject | title}} fetched",
		Failure:  "Failed to fetch {{.Subject}}",
	},

	"core.generate": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "generate",
			Default: "yes",
		},
		Question: "Generate {{.Subject}}?",
		Confirm:  "Generate {{.Subject}}?",
		Success:  "{{.Subject | title}} generated",
		Failure:  "Failed to generate {{.Subject}}",
	},

	"core.validate": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "validate",
			Default: "yes",
		},
		Question: "Validate {{.Subject}}?",
		Confirm:  "Validate {{.Subject}}?",
		Success:  "{{.Subject | title}} valid",
		Failure:  "{{.Subject | title}} invalid",
	},

	"core.check": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "check",
			Default: "yes",
		},
		Question: "Check {{.Subject}}?",
		Confirm:  "Check {{.Subject}}?",
		Success:  "{{.Subject | title}} OK",
		Failure:  "{{.Subject | title}} failed",
	},

	"core.scan": {
		Meta: IntentMeta{
			Type:    "action",
			Verb:    "scan",
			Default: "yes",
		},
		Question: "Scan {{.Subject}}?",
		Confirm:  "Scan {{.Subject}}?",
		Success:  "{{.Subject | title}} scanned",
		Failure:  "Failed to scan {{.Subject}}",
	},
}

// customIntents holds user-registered intents.
// Separated from coreIntents to allow thread-safe registration.
var (
	customIntents   = make(map[string]Intent)
	customIntentsMu sync.RWMutex
)

// getIntent retrieves an intent by its key.
// Checks custom intents first, then falls back to core intents.
// Returns nil if the intent is not found.
func getIntent(key string) *Intent {
	// Check custom intents first (thread-safe)
	customIntentsMu.RLock()
	if intent, ok := customIntents[key]; ok {
		customIntentsMu.RUnlock()
		return &intent
	}
	customIntentsMu.RUnlock()

	// Fall back to core intents
	if intent, ok := coreIntents[key]; ok {
		return &intent
	}
	return nil
}

// RegisterIntent adds a custom intent at runtime.
// Use this to extend the built-in intents with application-specific ones.
// This function is thread-safe.
//
//	i18n.RegisterIntent("myapp.archive", i18n.Intent{
//	    Meta: i18n.IntentMeta{Type: "action", Verb: "archive", Default: "yes"},
//	    Question: "Archive {{.Subject}}?",
//	    Success: "{{.Subject | title}} archived",
//	    Failure: "Failed to archive {{.Subject}}",
//	})
func RegisterIntent(key string, intent Intent) {
	customIntentsMu.Lock()
	defer customIntentsMu.Unlock()
	customIntents[key] = intent
}

// RegisterIntents adds multiple custom intents at runtime.
// This is more efficient than calling RegisterIntent multiple times.
// This function is thread-safe.
//
//	i18n.RegisterIntents(map[string]i18n.Intent{
//	    "myapp.archive": {
//	        Meta: i18n.IntentMeta{Type: "action", Verb: "archive"},
//	        Question: "Archive {{.Subject}}?",
//	    },
//	    "myapp.export": {
//	        Meta: i18n.IntentMeta{Type: "action", Verb: "export"},
//	        Question: "Export {{.Subject}}?",
//	    },
//	})
func RegisterIntents(intents map[string]Intent) {
	customIntentsMu.Lock()
	defer customIntentsMu.Unlock()
	for k, v := range intents {
		customIntents[k] = v
	}
}

// UnregisterIntent removes a custom intent by key.
// This only affects custom intents, not core intents.
// This function is thread-safe.
func UnregisterIntent(key string) {
	customIntentsMu.Lock()
	defer customIntentsMu.Unlock()
	delete(customIntents, key)
}

// IntentKeys returns all registered intent keys (both core and custom).
func IntentKeys() []string {
	customIntentsMu.RLock()
	defer customIntentsMu.RUnlock()

	keys := make([]string, 0, len(coreIntents)+len(customIntents))
	for key := range coreIntents {
		keys = append(keys, key)
	}
	for key := range customIntents {
		// Avoid duplicates if custom overrides core
		found := false
		for _, k := range keys {
			if k == key {
				found = true
				break
			}
		}
		if !found {
			keys = append(keys, key)
		}
	}
	return keys
}

// HasIntent returns true if an intent with the given key exists.
func HasIntent(key string) bool {
	return getIntent(key) != nil
}

// GetIntent returns the intent for a key, or nil if not found.
// This is the public API for retrieving intents.
func GetIntent(key string) *Intent {
	return getIntent(key)
}
