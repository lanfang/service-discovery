package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegistDomain(t *testing.T) {
	domain := RegistDomain(filepath.Base(os.Args[0]), "1.2.3.4", ":8080")
	if domain == "" {
		t.Errorf("RegistDomain err")
	} else {
		t.Logf("RegistDomain success %v", domain)
	}
}
