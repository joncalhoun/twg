package di_pkg_func

import (
	"os/exec"
	"strings"
)

var execCommand = exec.Command

func GitVersion() string {
	cmd := execCommand("git", "version")
	stdout, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	n := len("git version ")
	version := strings.TrimSpace(string(stdout[n:]))
	return version
}
