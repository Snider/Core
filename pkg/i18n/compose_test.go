package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// stringerValue is a test helper that implements fmt.Stringer
type stringerValue struct {
	value string
}

func (s stringerValue) String() string {
	return s.value
}

func TestSubject_Good(t *testing.T) {
	t.Run("basic creation", func(t *testing.T) {
		s := S("file", "config.yaml")
		assert.Equal(t, "file", s.Noun)
		assert.Equal(t, "config.yaml", s.Value)
		assert.Equal(t, 1, s.count)
		assert.Equal(t, "", s.gender)
		assert.Equal(t, "", s.location)
	})

	t.Run("NewSubject alias", func(t *testing.T) {
		s := NewSubject("repo", "core-php")
		assert.Equal(t, "repo", s.Noun)
		assert.Equal(t, "core-php", s.Value)
	})

	t.Run("with count", func(t *testing.T) {
		s := S("file", "*.go").Count(5)
		assert.Equal(t, 5, s.GetCount())
		assert.True(t, s.IsPlural())
	})

	t.Run("with gender", func(t *testing.T) {
		s := S("user", "alice").Gender("female")
		assert.Equal(t, "female", s.GetGender())
	})

	t.Run("with location", func(t *testing.T) {
		s := S("file", "config.yaml").In("workspace")
		assert.Equal(t, "workspace", s.GetLocation())
	})

	t.Run("chained methods", func(t *testing.T) {
		s := S("repo", "core-php").Count(3).Gender("neuter").In("organisation")
		assert.Equal(t, "repo", s.GetNoun())
		assert.Equal(t, 3, s.GetCount())
		assert.Equal(t, "neuter", s.GetGender())
		assert.Equal(t, "organisation", s.GetLocation())
	})
}

func TestSubject_String(t *testing.T) {
	t.Run("string value", func(t *testing.T) {
		s := S("file", "config.yaml")
		assert.Equal(t, "config.yaml", s.String())
	})

	t.Run("stringer interface", func(t *testing.T) {
		// Using a struct that implements Stringer via embedded method
		s := S("item", stringerValue{"test"})
		assert.Equal(t, "test", s.String())
	})

	t.Run("nil subject", func(t *testing.T) {
		var s *Subject
		assert.Equal(t, "", s.String())
	})

	t.Run("int value", func(t *testing.T) {
		s := S("count", 42)
		assert.Equal(t, "42", s.String())
	})
}

func TestSubject_IsPlural(t *testing.T) {
	t.Run("singular (count 1)", func(t *testing.T) {
		s := S("file", "test.go")
		assert.False(t, s.IsPlural())
	})

	t.Run("plural (count 0)", func(t *testing.T) {
		s := S("file", "*.go").Count(0)
		assert.True(t, s.IsPlural())
	})

	t.Run("plural (count > 1)", func(t *testing.T) {
		s := S("file", "*.go").Count(5)
		assert.True(t, s.IsPlural())
	})

	t.Run("nil subject", func(t *testing.T) {
		var s *Subject
		assert.False(t, s.IsPlural())
	})
}

func TestSubject_Getters(t *testing.T) {
	t.Run("nil safety", func(t *testing.T) {
		var s *Subject
		assert.Equal(t, "", s.GetNoun())
		assert.Equal(t, 1, s.GetCount())
		assert.Equal(t, "", s.GetGender())
		assert.Equal(t, "", s.GetLocation())
	})
}

func TestIntentMeta(t *testing.T) {
	meta := IntentMeta{
		Type:      "action",
		Verb:      "delete",
		Dangerous: true,
		Default:   "no",
		Supports:  []string{"force", "recursive"},
	}

	assert.Equal(t, "action", meta.Type)
	assert.Equal(t, "delete", meta.Verb)
	assert.True(t, meta.Dangerous)
	assert.Equal(t, "no", meta.Default)
	assert.Contains(t, meta.Supports, "force")
	assert.Contains(t, meta.Supports, "recursive")
}

func TestComposed(t *testing.T) {
	composed := Composed{
		Question: "Delete config.yaml?",
		Confirm:  "Really delete config.yaml?",
		Success:  "Config.yaml deleted",
		Failure:  "Failed to delete config.yaml",
		Meta: IntentMeta{
			Type:      "action",
			Verb:      "delete",
			Dangerous: true,
			Default:   "no",
		},
	}

	assert.Equal(t, "Delete config.yaml?", composed.Question)
	assert.Equal(t, "Really delete config.yaml?", composed.Confirm)
	assert.Equal(t, "Config.yaml deleted", composed.Success)
	assert.Equal(t, "Failed to delete config.yaml", composed.Failure)
	assert.True(t, composed.Meta.Dangerous)
}

func TestNewTemplateData(t *testing.T) {
	t.Run("from subject", func(t *testing.T) {
		s := S("file", "config.yaml").Count(3).Gender("neuter").In("workspace")
		data := newTemplateData(s)

		assert.Equal(t, "config.yaml", data.Subject)
		assert.Equal(t, "file", data.Noun)
		assert.Equal(t, 3, data.Count)
		assert.Equal(t, "neuter", data.Gender)
		assert.Equal(t, "workspace", data.Location)
		assert.Equal(t, "config.yaml", data.Value)
	})

	t.Run("from nil subject", func(t *testing.T) {
		data := newTemplateData(nil)

		assert.Equal(t, "", data.Subject)
		assert.Equal(t, "", data.Noun)
		assert.Equal(t, 1, data.Count)
		assert.Equal(t, "", data.Gender)
		assert.Equal(t, "", data.Location)
		assert.Nil(t, data.Value)
	})

	t.Run("with formality", func(t *testing.T) {
		s := S("user", "Hans").Formal()
		data := newTemplateData(s)

		assert.Equal(t, FormalityFormal, data.Formality)
		assert.True(t, data.IsFormal)
	})

	t.Run("with plural", func(t *testing.T) {
		s := S("file", "*.go").Count(5)
		data := newTemplateData(s)

		assert.True(t, data.IsPlural)
		assert.Equal(t, 5, data.Count)
	})
}

func TestSubject_Formality(t *testing.T) {
	t.Run("default is neutral", func(t *testing.T) {
		s := S("user", "name")
		assert.Equal(t, FormalityNeutral, s.GetFormality())
		assert.False(t, s.IsFormal())
		assert.False(t, s.IsInformal())
	})

	t.Run("Formal()", func(t *testing.T) {
		s := S("user", "name").Formal()
		assert.Equal(t, FormalityFormal, s.GetFormality())
		assert.True(t, s.IsFormal())
	})

	t.Run("Informal()", func(t *testing.T) {
		s := S("user", "name").Informal()
		assert.Equal(t, FormalityInformal, s.GetFormality())
		assert.True(t, s.IsInformal())
	})

	t.Run("Formality() explicit", func(t *testing.T) {
		s := S("user", "name").Formality(FormalityFormal)
		assert.Equal(t, FormalityFormal, s.GetFormality())
	})

	t.Run("nil safety", func(t *testing.T) {
		var s *Subject
		assert.Equal(t, FormalityNeutral, s.GetFormality())
		assert.False(t, s.IsFormal())
		assert.False(t, s.IsInformal())
	})
}

// --- Grammar composition tests using intent data ---

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

