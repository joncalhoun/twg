package di_pkg_func

import (
	"os/exec"
	"testing"
)

func TestGitVersion(t *testing.T) {
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("echo", "git version 2.17.1")
	}
	defer func() {
		execCommand = exec.Command
	}()
	want := "2.17.1"
	got := GitVersion()
	if got != want {
		t.Errorf("GitVersion() = %q; want %q", got, want)
	}
}
