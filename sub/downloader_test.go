package sub_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/joncalhoun/twg/sub"
)

func TestDemo(t *testing.T) {
	fmt.Println(os.Args)
	cmd := exec.Command(os.Args[0], "-test.run=Test_GitCloneSubprocess")
	_, err := cmd.Output()
	if err != nil {
		// fmt.Printf("err = %s\n", err)
	}
	// fmt.Println(string(out))
}

// This is not a real test - it used to simulate the git subprocess
func Test_GitCloneSubprocess(t *testing.T) {
	if os.Getenv("GO_RUNNING_SUBPROCESS") != "1" {
		t.Skip("not a subprocess")
	}
	defer os.Exit(0)

	var args []string
	for i := range os.Args {
		if os.Args[i] == "--" {
			args = os.Args[i+1:]
		}
	}
	dir := args[3]
	os.Mkdir(dir, 777)
}

func TestDownloader_Download(t *testing.T) {
	wantDir := "test-123"
	// Teardown
	defer func() {
		os.Remove(wantDir)
	}()

	d := sub.NewDownloader()
	// test starts here
	d.CloneCmd = exec.Command(os.Args[0], append([]string{"-test.run=Test_GitCloneSubprocess", "--"}, d.CloneCmd.Args...)...)
	d.CloneCmd.Env = append(os.Environ(), "GO_RUNNING_SUBPROCESS=1")
	msg, err := d.Download("https://github.com/joncalhoun/form.git", wantDir)
	if err != nil {
		t.Errorf("Download() err = %s; want nil", err)
		t.Errorf("Download() output: %s", msg)
	}
	if _, err := os.Stat(wantDir); os.IsNotExist(err) {
		t.Errorf("Download() didn't create dir %s", wantDir)
		t.Errorf("Download() output: %s", msg)
	}
}
