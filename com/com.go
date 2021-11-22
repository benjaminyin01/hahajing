package com

import (
	"os"
	"strings"

	"github.com/op/go-logging"
)

// HhjLog is HHJ system log
var HhjLog = logging.MustGetLogger("hhj")
var logformat = logging.MustStringFormatter(
	`%{color}%{time:2006-01-02 15:04:05.000} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func init() {
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, logformat)

	backendLeveled := logging.AddModuleLevel(backend)
	backendLeveled.SetLevel(logging.CRITICAL, "")

	// Set the backends to be used.
	logging.SetBackend(backendLeveled, backendFormatter)
}

// GetConfigPath x
func GetConfigPath() string {
	path := os.Args[0]

	i := strings.LastIndex(path, "\\")
	if i == -1 {
		i = strings.LastIndex(path, "/")
	}

	if i == -1 {
		HhjLog.Fatalf("Config path error: %s", path)
	}

	path = string(path[0:i])

	return path
}
