package envmanager

import (
	"fmt"
	"regexp"
	"strings"
)

var scopePattern = regexp.MustCompile(`^[a-zA-Z0-9_\-/]+$`)

func ValidateScope(scope string) error {
	if strings.TrimSpace(scope) == "" {
		return fmt.Errorf("invalid scope name. Use alphanumeric, dash, underscore, slash only")
	}
	if strings.Contains(scope, "..") || !scopePattern.MatchString(scope) {
		return fmt.Errorf("invalid scope name. Use alphanumeric, dash, underscore, slash only")
	}
	return nil
}

func ScopeToURL(scope string) string {
	return strings.ReplaceAll(scope, "/", "-")
}

func ScopeFromURL(scope string) string {
	return strings.ReplaceAll(scope, "-", "/")
}
