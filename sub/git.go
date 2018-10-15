package sub

import (
	"os/exec"
)

var execCommand = exec.Command

func GitStatus() (string, error) {
	cmd := execCommand("git", "status")
	status, err := cmd.Output()
	return string(status), err
}
