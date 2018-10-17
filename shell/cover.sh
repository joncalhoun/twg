# Via https://twitter.com/davecheney/status/1002384377802735617
go_cover() {
  local t=$(mktemp -t cover)
  go test $COVERFLAGS -coverprofile=$t $@ && go tool cover -func=$t && unlink $t
}
