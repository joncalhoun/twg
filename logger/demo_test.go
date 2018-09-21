package logger_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/joncalhoun/twg/logger"
)

type fakeLogger struct {
	sb strings.Builder
}

func (fl *fakeLogger) Println(v ...interface{}) {
	fmt.Fprintln(&fl.sb, v...)
}

func TestDemo(t *testing.T) {
	var fl fakeLogger
	logger.Demo(&fl)
	want := "something went wrong!"
	got := fl.sb.String()
	if !strings.Contains(got, want) {
		t.Errorf("Logs = %q; want substring %q", got, want)
	}
}
