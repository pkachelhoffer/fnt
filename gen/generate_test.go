package gen

import (
	"os/user"
	"path/filepath"
	"testing"
)

func TestGetInterface(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		return
	}

	err = PerformTypeGeneration(filepath.Join(usr.HomeDir, "github/fnt/testing"), "TestInterface", "testing", "./../testing/gen.go")
	if err != nil {
		t.Fatalf("err: %e", err)
	}
}
