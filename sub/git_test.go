package sub

import (
	"os/exec"
	"strings"
	"testing"
)

var (
	testHasGit bool
)

func init() {
	if _, err := exec.LookPath("git"); err == nil {
		testHasGit = true
	}
}

func TestGitStatus_mock(t *testing.T) {
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("echo", "On branch master")
	}
	defer func() {
		execCommand = exec.Command
	}()

	want := "On branch master"
	got, err := GitStatus()
	if err != nil {
		t.Fatalf("GitStatus() err = %s; want nil", err)
	}
	if !strings.HasPrefix(got, want) {
		t.Errorf("GitStatus() = %s; want prefix %s", got, want)
	}
}

func TestGitStatus_actual(t *testing.T) {
	if !testHasGit {
		t.Skip("git not found")
	}
	want := "On branch master"
	got, err := GitStatus()
	if err != nil {
		t.Fatalf("GitStatus() err = %s; want nil", err)
	}
	if !strings.HasPrefix(got, want) {
		t.Errorf("GitStatus() = %s; want prefix %s", got, want)
	}
}
