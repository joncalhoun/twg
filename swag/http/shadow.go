package http

import "net/http"

var ListenAndServe = http.ListenAndServe

type Handler http.Handler
