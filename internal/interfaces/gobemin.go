// Package interfaces provides the interfaces for the GNyx application
package interfaces

import (
	"net/http"

	logz "github.com/kubex-ecosystem/logz"
)

type IGNyx interface {
	StartGNyx()
	HandleValidate(w http.ResponseWriter, r *http.Request)
	HandleContact(w http.ResponseWriter, r *http.Request)
	RateLimit(w http.ResponseWriter, r *http.Request) bool
	Initialize() error
	GetLogFilePath() string
	GetConfigFilePath() string
	GetLogger() *logz.LoggerZ
	Mu() IMutexes
	GetReference() IReference
	Environment() IEnvironment
}
