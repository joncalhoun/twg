package urlpath

import (
	stdpath "path"
	"strings"
)

// Clean is a wrapper around the standard library's path.Clean with
// the exception that it makes sure all returned paths start with a
// slash and end with a slash.
func Clean(path string) string {
	path = stdpath.Clean("/" + path)
	if path[len(path)-1] != '/' {
		path = path + "/"
	}
	return path
}

// Split is used to extract the first piece of a path between slashes and
// returns this along with the remainder of the path. For instance, given
// the following path:
//
//   /this/is/my/path/
//
// Split would return:
//
//   "this", "/is/my/path/"
//
// Note: The name of this function kinda sucks, but it works for now.
func Split(path string) (head, tail string) {
	path = Clean(path)
	parts := strings.SplitN(path[1:], "/", 2)
	if len(parts) < 2 {
		parts = append(parts, "/")
	}
	return parts[0], Clean(parts[1])
}
