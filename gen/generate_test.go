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

	spec, iTracker, err := GetInterfaceSpec(filepath.Join(usr.HomeDir, "github/fnt/testing"), "TestInterface", "testing")
	if err != nil {
		t.Fatalf("err: %e", err)
	}

	err = GenerateFile(spec, iTracker, "./../testing/gen.go")
	if err != nil {
		t.Fatalf("err: %e", err)
	}
}
