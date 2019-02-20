package urlpath_test

import (
	"testing"

	"github.com/joncalhoun/twg/swag/urlpath"
)

func TestClean(t *testing.T) {
	tests := map[string]struct {
		path string
		want string
	}{
		"empty":            {"", "/"},
		"no slashes":       {"abc", "/abc/"},
		"leading slash":    {"/abc", "/abc/"},
		"trailing slash":   {"abc/", "/abc/"},
		"multiple slashes": {"a/b/c/123", "/a/b/c/123/"},
		"extra slashes":    {"//a///b//c", "/a/b/c/"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := urlpath.Clean(tc.path); got != tc.want {
				t.Fatalf("Clean() = %v; want %v", got, tc.want)
			}
		})
	}
}

func TestSplit(t *testing.T) {
	tests := map[string]struct {
		path     string
		wantHead string
		wantTail string
	}{
		"one element":      {"/thing/", "thing", "/"},
		"two elements":     {"/one/two/", "one", "/two/"},
		"leading slash":    {"/abc", "abc", "/"},
		"trailing slash":   {"abc/", "abc", "/"},
		"multiple slashes": {"a/b/c/123", "a", "/b/c/123/"},
		"extra slashes":    {"//a///b//c", "a", "/b/c/"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gotHead, gotTail := urlpath.Split(tc.path)
			if gotHead != tc.wantHead {
				t.Errorf("Split() head = %v; want %v", gotHead, tc.wantHead)
			}
			if gotTail != tc.wantTail {
				t.Errorf("Split() tail = %v; want %v", gotTail, tc.wantTail)
			}
		})
	}
}
