package model

import "testing"

func TestDBImportScopesValidateRejectsEmptyObject(t *testing.T) {
	if err := (&DBImportScopes{}).Validate(); err == nil {
		t.Fatal("Validate() error = nil, want at least one import scope must be enabled")
	} else if err.Error() != "at least one import scope must be enabled" {
		t.Fatalf("Validate() error = %v, want at least one import scope must be enabled", err)
	}
}

func TestDBImportScopesValidateAllowsSelectedScopes(t *testing.T) {
	if err := (&DBImportScopes{Settings: true}).Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}
