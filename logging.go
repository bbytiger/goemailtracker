package emailtracker

import "net/http"

// extend to implement your own logging
type Logger interface {
	LogRequest(r *http.Request)
	LogResponse(w http.ResponseWriter)
	LogEndpointError(err error)
	LogPkgError(err error)
}

type DefaultLogger struct{}

func (d DefaultLogger) LogRequest(r *http.Request) {

}

func (d DefaultLogger) LogResponse(w http.ResponseWriter) {

}

func (d DefaultLogger) LogEndpointError(err error) {

}
