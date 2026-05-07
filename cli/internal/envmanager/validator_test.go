package envmanager

import "testing"

func TestValidateScopeName(t *testing.T) {
	valid := []string{"global", "global/aws", "myapp/production", "my_app/staging"}
	for _, s := range valid {
		if err := ValidateScope(s); err != nil {
			t.Fatalf("expected valid scope %q, got err %v", s, err)
		}
	}

	invalid := []string{"", "../escape", "myapp production", "myapp#prod"}
	for _, s := range invalid {
		if err := ValidateScope(s); err == nil {
			t.Fatalf("expected invalid scope %q", s)
		}
	}
}

func TestScopeURLConversion(t *testing.T) {
	out := ScopeToURL("myapp/production")
	if out != "myapp-production" {
		t.Fatalf("expected myapp-production, got %q", out)
	}

	back := ScopeFromURL("myapp-production")
	if back != "myapp/production" {
		t.Fatalf("expected myapp/production, got %q", back)
	}
}
