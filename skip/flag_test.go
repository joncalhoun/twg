package skip

import (
	"flag"
	"os"
	"testing"
)

var integration = false

func init() {
	flag.BoolVar(&integration, "integration", false, "run database integration tests")
}

func TestMain(m *testing.M) {
	flag.Parse()
	if integration {
		// setup integration stuff if you need to
	}
	result := m.Run()
	if integration {
		// teardown integration stuff if you need to
	}
	os.Exit(result)
}

func TestWithFlag(t *testing.T) {
	if !integration {
		t.Skip()
	}
	t.Log("Running the integration test...")
}
