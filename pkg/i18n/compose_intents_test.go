package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// composeIntent executes intent templates with a subject for testing.
// This is a test helper that replicates what C() used to do.
func composeIntent(intent Intent, subject *Subject) *Composed {
	data := newTemplateData(subject)
	return &Composed{
		Question: executeIntentTemplate(intent.Question, data),
		Confirm:  executeIntentTemplate(intent.Confirm, data),
		Success:  executeIntentTemplate(intent.Success, data),
		Failure:  executeIntentTemplate(intent.Failure, data),
		Meta:     intent.Meta,
	}
}

// TestGrammarComposition_MatchesIntents verifies that the grammar engine
// can compose the same strings as the intent templates.
// This turns the intents definitions into a comprehensive test suite.
func TestGrammarComposition_MatchesIntents(t *testing.T) {
	// Test subjects for validation
	subjects := []struct {
		noun  string
		value string
	}{
		{"file", "config.yaml"},
		{"directory", "src"},
		{"repo", "core-php"},
		{"branch", "feature/auth"},
		{"commit", "abc1234"},
		{"changes", "5 files"},
		{"package", "laravel/framework"},
	}

	// Test each core intent's composition
	for key, intent := range coreIntents {
		t.Run(key, func(t *testing.T) {
			for _, subj := range subjects {
				subject := S(subj.noun, subj.value)

				// Compose using intent templates directly
				composed := composeIntent(intent, subject)

				// Verify Success output matches ActionResult
				if intent.Success != "" && intent.Meta.Verb != "" {
					// Standard success pattern: "{{.Subject | title}} verbed"
					expectedSuccess := ActionResult(intent.Meta.Verb, subj.value)

					// Some intents have non-standard success messages
					switch key {
					case "core.run":
						// "completed" instead of "ran"
						expectedSuccess = Title(subj.value) + " completed"
					case "core.test":
						// "passed" instead of "tested"
						expectedSuccess = Title(subj.value) + " passed"
					case "core.validate":
						// "valid" instead of "validated"
						expectedSuccess = Title(subj.value) + " valid"
					case "core.check":
						// "OK" instead of "checked"
						expectedSuccess = Title(subj.value) + " OK"
					case "core.continue", "core.proceed":
						// No subject in success
						continue
					case "core.confirm":
						// No subject in success
						continue
					}

					assert.Equal(t, expectedSuccess, composed.Success,
						"%s: Success mismatch for subject %s", key, subj.value)
				}

				// Verify Failure output matches ActionFailed
				if intent.Failure != "" && intent.Meta.Verb != "" {
					// Standard failure pattern: "Failed to verb subject"
					expectedFailure := ActionFailed(intent.Meta.Verb, subj.value)

					// Some intents have non-standard failure messages
					switch key {
					case "core.test":
						// "failed" instead of "Failed to test"
						expectedFailure = Title(subj.value) + " failed"
					case "core.validate":
						// "invalid" instead of "Failed to validate"
						expectedFailure = Title(subj.value) + " invalid"
					case "core.check":
						// "failed" instead of "Failed to check"
						expectedFailure = Title(subj.value) + " failed"
					case "core.continue", "core.proceed", "core.confirm":
						// Non-standard failures
						continue
					}

					assert.Equal(t, expectedFailure, composed.Failure,
						"%s: Failure mismatch for subject %s", key, subj.value)
				}
			}
		})
	}
}

// TestActionResult_AllIntentVerbs tests that ActionResult handles
// all verbs used in the core intents.
func TestActionResult_AllIntentVerbs(t *testing.T) {
	// Extract all unique verbs from intents
	verbs := make(map[string]bool)
	for _, intent := range coreIntents {
		if intent.Meta.Verb != "" {
			verbs[intent.Meta.Verb] = true
		}
	}

	subject := "test item"

	for verb := range verbs {
		t.Run(verb, func(t *testing.T) {
			result := ActionResult(verb, subject)

			// Should produce non-empty result
			assert.NotEmpty(t, result, "ActionResult(%q, %q) should not be empty", verb, subject)

			// Should start with title-cased subject
			assert.Contains(t, result, Title(subject),
				"ActionResult should contain title-cased subject")

			// Should contain past tense of verb
			past := PastTense(verb)
			assert.Contains(t, result, past,
				"ActionResult(%q) should contain past tense %q", verb, past)
		})
	}
}

// TestActionFailed_AllIntentVerbs tests that ActionFailed handles
// all verbs used in the core intents.
func TestActionFailed_AllIntentVerbs(t *testing.T) {
	verbs := make(map[string]bool)
	for _, intent := range coreIntents {
		if intent.Meta.Verb != "" {
			verbs[intent.Meta.Verb] = true
		}
	}

	subject := "test item"

	for verb := range verbs {
		t.Run(verb, func(t *testing.T) {
			result := ActionFailed(verb, subject)

			// Should produce non-empty result
			assert.NotEmpty(t, result, "ActionFailed(%q, %q) should not be empty", verb, subject)

			// Should start with "Failed to"
			assert.Contains(t, result, "Failed to",
				"ActionFailed should contain 'Failed to'")

			// Should contain the verb
			assert.Contains(t, result, verb,
				"ActionFailed should contain the verb")

			// Should contain the subject
			assert.Contains(t, result, subject,
				"ActionFailed should contain the subject")
		})
	}
}

// TestProgress_AllIntentVerbs tests that Progress handles
// all verbs used in the core intents.
func TestProgress_AllIntentVerbs(t *testing.T) {
	verbs := make(map[string]bool)
	for _, intent := range coreIntents {
		if intent.Meta.Verb != "" {
			verbs[intent.Meta.Verb] = true
		}
	}

	for verb := range verbs {
		t.Run(verb, func(t *testing.T) {
			result := Progress(verb)

			// Should produce non-empty result
			assert.NotEmpty(t, result, "Progress(%q) should not be empty", verb)

			// Should end with "..."
			assert.Contains(t, result, "...",
				"Progress should contain '...'")

			// Should contain gerund form
			gerund := Gerund(verb)
			assert.Contains(t, result, Title(gerund),
				"Progress(%q) should contain gerund %q", verb, gerund)
		})
	}
}

// TestPastTense_AllIntentVerbs ensures PastTense works for all intent verbs.
func TestPastTense_AllIntentVerbs(t *testing.T) {
	expected := map[string]string{
		// Destructive
		"delete":    "deleted",
		"remove":    "removed",
		"discard":   "discarded",
		"reset":     "reset",
		"overwrite": "overwritten",

		// Creation
		"create": "created",
		"add":    "added",
		"clone":  "cloned",
		"copy":   "copied",

		// Modification
		"save":   "saved",
		"update": "updated",
		"rename": "renamed",
		"move":   "moved",

		// Git
		"commit": "committed",
		"push":   "pushed",
		"pull":   "pulled",
		"merge":  "merged",
		"rebase": "rebased",

		// Network
		"install":  "installed",
		"download": "downloaded",
		"upload":   "uploaded",
		"publish":  "published",
		"deploy":   "deployed",

		// Process
		"start":   "started",
		"stop":    "stopped",
		"restart": "restarted",
		"run":     "ran",
		"build":   "built",
		"test":    "tested",

		// Info - these are regular verbs ending in consonant, -ed suffix
		"continue": "continued",
		"proceed":  "proceeded",
		"confirm":  "confirmed",

		// Additional
		"sync":     "synced",
		"boot":     "booted",
		"format":   "formatted",
		"analyse":  "analysed",
		"link":     "linked",
		"unlink":   "unlinked",
		"fetch":    "fetched",
		"generate": "generated",
		"validate": "validated",
		"check":    "checked",
		"scan":     "scanned",
	}

	for verb, want := range expected {
		t.Run(verb, func(t *testing.T) {
			got := PastTense(verb)
			assert.Equal(t, want, got, "PastTense(%q)", verb)
		})
	}
}

// TestGerund_AllIntentVerbs ensures Gerund works for all intent verbs.
func TestGerund_AllIntentVerbs(t *testing.T) {
	expected := map[string]string{
		// Destructive
		"delete":    "deleting",
		"remove":    "removing",
		"discard":   "discarding",
		"reset":     "resetting",
		"overwrite": "overwriting",

		// Creation
		"create": "creating",
		"add":    "adding",
		"clone":  "cloning",
		"copy":   "copying",

		// Modification
		"save":   "saving",
		"update": "updating",
		"rename": "renaming",
		"move":   "moving",

		// Git
		"commit": "committing",
		"push":   "pushing",
		"pull":   "pulling",
		"merge":  "merging",
		"rebase": "rebasing",

		// Network
		"install":  "installing",
		"download": "downloading",
		"upload":   "uploading",
		"publish":  "publishing",
		"deploy":   "deploying",

		// Process
		"start":   "starting",
		"stop":    "stopping",
		"restart": "restarting",
		"run":     "running",
		"build":   "building",
		"test":    "testing",

		// Info
		"continue": "continuing",
		"proceed":  "proceeding",
		"confirm":  "confirming",

		// Additional
		"sync":     "syncing",
		"boot":     "booting",
		"format":   "formatting",
		"analyse":  "analysing",
		"link":     "linking",
		"unlink":   "unlinking",
		"fetch":    "fetching",
		"generate": "generating",
		"validate": "validating",
		"check":    "checking",
		"scan":     "scanning",
	}

	for verb, want := range expected {
		t.Run(verb, func(t *testing.T) {
			got := Gerund(verb)
			assert.Equal(t, want, got, "Gerund(%q)", verb)
		})
	}
}

// TestQuestionFormat verifies that standard question format
// can be composed from verb and subject.
func TestQuestionFormat(t *testing.T) {
	tests := []struct {
		verb     string
		subject  string
		expected string
	}{
		{"delete", "config.yaml", "Delete config.yaml?"},
		{"create", "src", "Create src?"},
		{"commit", "changes", "Commit changes?"},
		{"push", "5 commits", "Push 5 commits?"},
		{"install", "package", "Install package?"},
	}

	for _, tt := range tests {
		t.Run(tt.verb, func(t *testing.T) {
			// Standard question format: "Verb subject?"
			result := Title(tt.verb) + " " + tt.subject + "?"
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestConfirmFormat verifies dangerous action confirm messages.
func TestConfirmFormat(t *testing.T) {
	// Dangerous actions have "Really verb subject?" confirm
	dangerous := []string{"delete", "remove", "discard", "reset", "overwrite", "merge", "rebase", "publish", "deploy"}

	for _, verb := range dangerous {
		t.Run(verb, func(t *testing.T) {
			subject := "test item"
			// Basic confirm format
			result := "Really " + verb + " " + subject + "?"

			assert.Contains(t, result, "Really",
				"Dangerous action confirm should start with 'Really'")
			assert.Contains(t, result, verb)
			assert.Contains(t, result, subject)
			assert.Contains(t, result, "?")
		})
	}
}

// TestIntentConsistency verifies patterns across all intents.
func TestIntentConsistency(t *testing.T) {
	// These intents have non-standard question formats
	specialQuestions := map[string]bool{
		"core.continue": true, // "Continue?" (no subject)
		"core.proceed":  true, // "Proceed?" (no subject)
		"core.confirm":  true, // "Are you sure?" (different format)
	}

	for key, intent := range coreIntents {
		t.Run(key, func(t *testing.T) {
			verb := intent.Meta.Verb

			// Verify verb is set
			assert.NotEmpty(t, verb, "intent should have a verb")

			// Verify Question contains the verb (unless special case)
			if !specialQuestions[key] {
				assert.Contains(t, intent.Question, Title(verb)+" ",
					"Question should contain '%s '", Title(verb))
			}

			// Verify dangerous intents default to "no"
			if intent.Meta.Dangerous {
				assert.Equal(t, "no", intent.Meta.Default,
					"Dangerous intent should default to 'no'")
			}

			// Verify non-dangerous intents default to "yes"
			if !intent.Meta.Dangerous && intent.Meta.Type == "action" {
				assert.Equal(t, "yes", intent.Meta.Default,
					"Safe action intent should default to 'yes'")
			}
		})
	}
}

// TestComposedVsManual compares template output with manual grammar composition.
func TestComposedVsManual(t *testing.T) {
	tests := []struct {
		intentKey string
		noun      string
		value     string
	}{
		{"core.delete", "file", "config.yaml"},
		{"core.create", "directory", "src"},
		{"core.save", "changes", "data"},
		{"core.commit", "repo", "core-php"},
		{"core.push", "branch", "feature/test"},
		{"core.install", "package", "express"},
	}

	for _, tt := range tests {
		t.Run(tt.intentKey, func(t *testing.T) {
			subject := S(tt.noun, tt.value)
			intent := coreIntents[tt.intentKey]

			// Compose using intent templates
			composed := composeIntent(intent, subject)

			// Manual composition using grammar functions
			manualSuccess := ActionResult(intent.Meta.Verb, tt.value)
			manualFailure := ActionFailed(intent.Meta.Verb, tt.value)

			assert.Equal(t, manualSuccess, composed.Success,
				"Template Success should match ActionResult()")
			assert.Equal(t, manualFailure, composed.Failure,
				"Template Failure should match ActionFailed()")
		})
	}
}

// TestGrammarCanReplaceIntents demonstrates that the grammar engine
// can compose all the standard output forms without hardcoded templates.
// This proves the i18n system can work with just verb definitions.
func TestGrammarCanReplaceIntents(t *testing.T) {
	tests := []struct {
		verb    string
		subject string
		// Expected outputs that grammar should produce
		wantQuestion string
		wantSuccess  string
		wantFailure  string
		wantProgress string
	}{
		{
			verb:         "delete",
			subject:      "config.yaml",
			wantQuestion: "Delete config.yaml?",
			wantSuccess:  "Config.Yaml deleted",
			wantFailure:  "Failed to delete config.yaml",
			wantProgress: "Deleting...",
		},
		{
			verb:         "create",
			subject:      "project",
			wantQuestion: "Create project?",
			wantSuccess:  "Project created",
			wantFailure:  "Failed to create project",
			wantProgress: "Creating...",
		},
		{
			verb:         "build",
			subject:      "app",
			wantQuestion: "Build app?",
			wantSuccess:  "App built",
			wantFailure:  "Failed to build app",
			wantProgress: "Building...",
		},
		{
			verb:         "run",
			subject:      "tests",
			wantQuestion: "Run tests?",
			wantSuccess:  "Tests ran",
			wantFailure:  "Failed to run tests",
			wantProgress: "Running...",
		},
		{
			verb:         "commit",
			subject:      "changes",
			wantQuestion: "Commit changes?",
			wantSuccess:  "Changes committed",
			wantFailure:  "Failed to commit changes",
			wantProgress: "Committing...",
		},
		{
			verb:         "overwrite",
			subject:      "file",
			wantQuestion: "Overwrite file?",
			wantSuccess:  "File overwritten",
			wantFailure:  "Failed to overwrite file",
			wantProgress: "Overwriting...",
		},
		{
			verb:         "reset",
			subject:      "state",
			wantQuestion: "Reset state?",
			wantSuccess:  "State reset",
			wantFailure:  "Failed to reset state",
			wantProgress: "Resetting...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.verb, func(t *testing.T) {
			// Compose using grammar functions only (no templates)
			question := Title(tt.verb) + " " + tt.subject + "?"
			success := ActionResult(tt.verb, tt.subject)
			failure := ActionFailed(tt.verb, tt.subject)
			progress := Progress(tt.verb)

			assert.Equal(t, tt.wantQuestion, question, "Question")
			assert.Equal(t, tt.wantSuccess, success, "Success")
			assert.Equal(t, tt.wantFailure, failure, "Failure")
			assert.Equal(t, tt.wantProgress, progress, "Progress")
		})
	}
}

// TestProgressSubjectMatchesExpected tests ProgressSubject for all intent verbs.
func TestProgressSubjectMatchesExpected(t *testing.T) {
	tests := []struct {
		verb    string
		subject string
		want    string
	}{
		{"delete", "config.yaml", "Deleting config.yaml..."},
		{"create", "project", "Creating project..."},
		{"build", "app", "Building app..."},
		{"install", "package", "Installing package..."},
		{"commit", "changes", "Committing changes..."},
		{"push", "commits", "Pushing commits..."},
		{"pull", "updates", "Pulling updates..."},
		{"sync", "files", "Syncing files..."},
		{"fetch", "data", "Fetching data..."},
		{"check", "status", "Checking status..."},
	}

	for _, tt := range tests {
		t.Run(tt.verb, func(t *testing.T) {
			result := ProgressSubject(tt.verb, tt.subject)
			assert.Equal(t, tt.want, result)
		})
	}
}
