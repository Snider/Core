// Package i18n provides internationalization for the CLI.
package i18n

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
}

// getIntent retrieves an intent by its key from the core intents.
// Returns nil if the intent is not found.
func getIntent(key string) *Intent {
	if intent, ok := coreIntents[key]; ok {
		return &intent
	}
	return nil
}

// RegisterIntent adds a custom intent to the core intents.
// Use this to extend the built-in intents with application-specific ones.
//
//	i18n.RegisterIntent("myapp.archive", i18n.Intent{
//	    Meta: i18n.IntentMeta{Type: "action", Verb: "archive", Default: "yes"},
//	    Question: "Archive {{.Subject}}?",
//	    Success: "{{.Subject | title}} archived",
//	    Failure: "Failed to archive {{.Subject}}",
//	})
func RegisterIntent(key string, intent Intent) {
	coreIntents[key] = intent
}

// IntentKeys returns all registered intent keys.
func IntentKeys() []string {
	keys := make([]string, 0, len(coreIntents))
	for key := range coreIntents {
		keys = append(keys, key)
	}
	return keys
}
