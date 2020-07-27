package main

import (
	"net/http"
	_ "net/http/pprof"
)

func launchpprof() error {
	return http.ListenAndServe(":6060", nil)
}
