package secretlogic

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// GenerateJSONDiff generates a human-readable diff between two JSON structures
func GenerateJSONDiff(local, remote interface{}) string {
	localJSON, _ := json.MarshalIndent(local, "", "  ")
	remoteJSON, _ := json.MarshalIndent(remote, "", "  ")

	localLines := strings.Split(string(localJSON), "\n")
	remoteLines := strings.Split(string(remoteJSON), "\n")

	var diff strings.Builder

	// Simple line-by-line comparison
	maxLines := len(localLines)
	if len(remoteLines) > maxLines {
		maxLines = len(remoteLines)
	}

	for i := 0; i < maxLines; i++ {
		var localLine, remoteLine string
		if i < len(localLines) {
			localLine = localLines[i]
		}
		if i < len(remoteLines) {
			remoteLine = remoteLines[i]
		}

		if localLine != remoteLine {
			if remoteLine != "" && localLine == "" {
				diff.WriteString(fmt.Sprintf("- %s\n", remoteLine))
			} else if localLine != "" && remoteLine == "" {
				diff.WriteString(fmt.Sprintf("+ %s\n", localLine))
			} else {
				diff.WriteString(fmt.Sprintf("- %s\n", remoteLine))
				diff.WriteString(fmt.Sprintf("+ %s\n", localLine))
			}
		} else {
			diff.WriteString(fmt.Sprintf("  %s\n", localLine))
		}
	}

	return diff.String()
}

// SecretsAreEqual compares two secrets and returns true if they are equal
func SecretsAreEqual(local, remote interface{}) bool {
	return reflect.DeepEqual(local, remote)
}
