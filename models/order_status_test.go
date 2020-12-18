package models_test

// AUTOGENERATED BY: HSM GEN

import(
	"testing"

	m "github.com/bigpanther/trober/models"
)

func TestIsValidOrderStatus(t *testing.T) {
	var validVal = "Open"
	var inValidVal = "_someInvalidval_"
	if !m.IsValidOrderStatus(validVal) {
		t.Fatalf("IsValidOrderStatus(%q) should be true", validVal)
	}
	if m.IsValidOrderStatus(inValidVal) {
		t.Fatalf("IsValidOrderStatus(%q) should be false", inValidVal)
	}
}
