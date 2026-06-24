package secretlogic

import (
	"strings"
	"testing"
)

func TestSecretsAreEqual(t *testing.T) {
	tests := []struct {
		name     string
		local    interface{}
		remote   interface{}
		expected bool
	}{
		{
			name:     "identical structs",
			local:    struct{ Name string }{"test"},
			remote:   struct{ Name string }{"test"},
			expected: true,
		},
		{
			name:     "different structs",
			local:    struct{ Name string }{"test1"},
			remote:   struct{ Name string }{"test2"},
			expected: false,
		},
		{
			name:     "identical maps",
			local:    map[string]string{"key": "value"},
			remote:   map[string]string{"key": "value"},
			expected: true,
		},
		{
			name:     "different maps",
			local:    map[string]string{"key": "value1"},
			remote:   map[string]string{"key": "value2"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SecretsAreEqual(tt.local, tt.remote)
			if result != tt.expected {
				t.Errorf("SecretsAreEqual() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateJSONDiff(t *testing.T) {
	local := map[string]string{
		"field1": "value1",
		"field2": "new-value",
	}

	remote := map[string]string{
		"field1": "value1",
		"field2": "old-value",
	}

	diff := GenerateJSONDiff(local, remote)

	// Check that diff contains indicators
	if !strings.Contains(diff, "-") {
		t.Error("Diff should contain '-' for removed lines")
	}
	if !strings.Contains(diff, "+") {
		t.Error("Diff should contain '+' for added lines")
	}
	if !strings.Contains(diff, "old-value") {
		t.Error("Diff should contain old value")
	}
	if !strings.Contains(diff, "new-value") {
		t.Error("Diff should contain new value")
	}
}

func TestGenerateJSONDiff_NoDifference(t *testing.T) {
	data := map[string]string{
		"field1": "value1",
		"field2": "value2",
	}

	diff := GenerateJSONDiff(data, data)

	// All lines should be unchanged (prefixed with spaces)
	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if line != "" && !strings.HasPrefix(line, " ") {
			t.Errorf("Expected all lines to start with space, got: %s", line)
		}
	}
}
