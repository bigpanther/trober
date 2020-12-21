package grifts

import "testing"

func TestDemoCreateDrop(t *testing.T) {
	err := demoCreate()
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	err = demoDrop()
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
}

func TestDemoDrop(t *testing.T) {
	err := demoDrop()
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
}
