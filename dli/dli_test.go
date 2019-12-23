package dli

import (
	"testing"
)

func TestDLIStrip(t *testing.T) {
	d := NewDLIStrip("1.1.1.1:9999")
	if d.Metadata().Name == "" {
		t.Error("HAL metadata should not have empty name")
	}

}

