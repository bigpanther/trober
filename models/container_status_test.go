package models_test

// AUTOGENERATED BY: HSM GEN

import(
	"testing"

	m "github.com/bigpanther/trober/models"
)

func TestIsValidContainerStatus(t *testing.T) {
	var validVal = "Unassigned"
	var inValidVal = "_someInvalidval_"
	if !m.IsValidContainerStatus(validVal) {
		t.Fatalf("IsValidContainerStatus(%q) should be true", validVal)
	}
	if m.IsValidContainerStatus(inValidVal) {
		t.Fatalf("IsValidContainerStatus(%q) should be false", inValidVal)
	}
}
