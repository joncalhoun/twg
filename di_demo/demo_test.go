package di_demo_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/joncalhoun/twg/di_demo"
)

type fakeLogger struct {
	sb strings.Builder
}

func (fl *fakeLogger) Println(v ...interface{}) {
	fmt.Fprintln(&fl.sb, v...)
}

func TestDemo(t *testing.T) {
	var fl fakeLogger
	di_demo.Demo(&fl)
	want := "failed to do the thing:"
	got := fl.sb.String()
	if !strings.Contains(got, want) {
		t.Errorf("logs = %q; want substring %q", got, want)
	}
}
