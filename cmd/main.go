package main

import (
	"github.com/kubex-ecosystem/domus/internal/module"
	gl "github.com/kubex-ecosystem/logz"
)

var (
	logger = gl.GetLoggerZ("domus")
)

// main initializes the logger and creates a new Domus instance.
func main() {
	if err := module.RegX().Command().Execute(); err != nil {
		logger.Fatal(err.Error())
	}
}
