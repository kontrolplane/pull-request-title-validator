// main_test.go
package main

import (
	"log/slog"
	"os"
	"testing"
)

func TestExtractTypeAndScope(t *testing.T) {
	tests := []struct {
		name          string
		prefix        string
		expectedType  string
		expectedScope string
	}{
		{
			name:          "type with scope",
			prefix:        "feat(api)",
			expectedType:  "feat",
			expectedScope: "api",
		},
		{
			name:          "type without scope",
			prefix:        "fix",
			expectedType:  "fix",
			expectedScope: "",
		},
		{
			name:          "type with complex scope",
			prefix:        "refactor(package/utils)",
			expectedType:  "refactor",
			expectedScope: "package/utils",
		},
		{
			name:          "type with scope containing special chars",
			prefix:        "feat(api/v2)",
			expectedType:  "feat",
			expectedScope: "api/v2",
		},
		{
			name:          "malformed scope - missing closing",
			prefix:        "feat(api",
			expectedType:  "feat(api",
			expectedScope: "",
		},
		{
			name:          "empty prefix",
			prefix:        "",
			expectedType:  "",
			expectedScope: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotScope := extractTypeAndScope(tt.prefix)
			if gotType != tt.expectedType {
				t.Errorf("extractTypeAndScope() type = %v, want %v", gotType, tt.expectedType)
			}
			if gotScope != tt.expectedScope {
				t.Errorf("extractTypeAndScope() scope = %v, want %v", gotScope, tt.expectedScope)
			}
		})
	}
}

func TestValidateType(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	validator := &Validator{logger: logger}

	tests := []struct {
		name         string
		titleType    string
		allowedTypes []string
		shouldPass   bool
	}{
		{
			name:         "valid type",
			titleType:    "feat",
			allowedTypes: []string{"feat", "fix", "chore"},
			shouldPass:   true,
		},
		{
			name:         "invalid type",
			titleType:    "invalid",
			allowedTypes: []string{"feat", "fix", "chore"},
			shouldPass:   false,
		},
		{
			name:         "empty type",
			titleType:    "",
			allowedTypes: []string{"feat", "fix", "chore"},
			shouldPass:   false,
		},
		{
			name:         "case sensitive",
			titleType:    "FEAT",
			allowedTypes: []string{"feat", "fix", "chore"},
			shouldPass:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateType(tt.titleType, tt.allowedTypes)
			if tt.shouldPass && err != nil {
				t.Errorf("validateType() should pass but got error: %v", err)
			}
			if !tt.shouldPass && err == nil {
				t.Errorf("validateType() should fail but passed")
			}
		})
	}
}

func TestValidateScope(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	validator := &Validator{logger: logger}

	tests := []struct {
		name          string
		titleScope    string
		allowedScopes []string
		shouldPass    bool
	}{
		{
			name:          "valid exact scope",
			titleScope:    "api",
			allowedScopes: []string{"api", "ui", "core"},
			shouldPass:    true,
		},
		{
			name:          "invalid scope",
			titleScope:    "database",
			allowedScopes: []string{"api", "ui", "core"},
			shouldPass:    false,
		},
		{
			name:          "regex pattern match",
			titleScope:    "package/utils",
			allowedScopes: []string{"package/.+", "api/v[0-9]+"},
			shouldPass:    true,
		},
		{
			name:          "regex pattern no match",
			titleScope:    "invalid/path",
			allowedScopes: []string{"package/.+", "api/v[0-9]+"},
			shouldPass:    false,
		},
		{
			name:          "case insensitive match",
			titleScope:    "API",
			allowedScopes: []string{"api", "ui", "core"},
			shouldPass:    true,
		},
		{
			name:          "empty scope with allowed scopes",
			titleScope:    "",
			allowedScopes: []string{"api", "ui", "core"},
			shouldPass:    false,
		},
		{
			name:          "empty scope with no restrictions",
			titleScope:    "",
			allowedScopes: []string{},
			shouldPass:    false, // Empty scope should not match empty pattern
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateScope(tt.titleScope, tt.allowedScopes)
			if tt.shouldPass && err != nil {
				t.Errorf("validateScope() should pass but got error: %v", err)
			}
			if !tt.shouldPass && err == nil {
				t.Errorf("validateScope() should fail but passed")
			}
		})
	}
}

func TestParseCommaSeparatedList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "normal list",
			input:    "feat,fix,chore",
			expected: []string{"feat", "fix", "chore"},
		},
		{
			name:     "list with spaces",
			input:    " feat , fix , chore ",
			expected: []string{"feat", "fix", "chore"},
		},
		{
			name:     "single item",
			input:    "feat",
			expected: []string{"feat"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{""},
		},
		{
			name:     "trailing comma",
			input:    "feat,fix,",
			expected: []string{"feat", "fix", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCommaSeparatedList(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseCommaSeparatedList() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("parseCommaSeparatedList()[%d] = %v, want %v", i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestParseTitle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	validator := &Validator{logger: logger}

	tests := []struct {
		name          string
		title         string
		expectedType  string
		expectedScope string
		expectedMsg   string
		shouldPass    bool
	}{
		{
			name:          "valid title with scope",
			title:         "feat(api): add new endpoint",
			expectedType:  "feat",
			expectedScope: "api",
			expectedMsg:   "add new endpoint",
			shouldPass:    true,
		},
		{
			name:          "valid title without scope",
			title:         "fix: resolve memory leak",
			expectedType:  "fix",
			expectedScope: "",
			expectedMsg:   "resolve memory leak",
			shouldPass:    true,
		},
		{
			name:       "invalid title - no colon",
			title:      "feat add new feature",
			shouldPass: false,
		},
		{
			name:       "invalid title - no type",
			title:      ": add new feature",
			shouldPass: false,
		},
		{
			name:          "title with complex scope",
			title:         "refactor(package/utils): optimize helper functions",
			expectedType:  "refactor",
			expectedScope: "package/utils",
			expectedMsg:   "optimize helper functions",
			shouldPass:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components, err := validator.parseTitle(tt.title)

			if tt.shouldPass {
				if err != nil {
					t.Errorf("parseTitle() should pass but got error: %v", err)
					return
				}
				if components.Type != tt.expectedType {
					t.Errorf("parseTitle() type = %v, want %v", components.Type, tt.expectedType)
				}
				if components.Scope != tt.expectedScope {
					t.Errorf("parseTitle() scope = %v, want %v", components.Scope, tt.expectedScope)
				}
				if components.Message != tt.expectedMsg {
					t.Errorf("parseTitle() message = %v, want %v", components.Message, tt.expectedMsg)
				}
			} else {
				if err == nil {
					t.Errorf("parseTitle() should fail but passed")
				}
			}
		})
	}
}
