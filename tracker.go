package emailtracker

import (
	"net/http"
)

type EmailTracker struct {
	WebhookURL    string // support tracking links later
	Encoding      string
	Pixel         []byte
	PixelMimetype string
	ExternalURL   string
	ExternalConnector
}

type MailMetadata struct {
}

type ExternalConnector interface {
	// interface for connecting your external service
	NotifyExternal(metadata *MailMetadata)
}

func NewEmailTracker(connector ExternalConnector) *EmailTracker {
	return &EmailTracker{
		Pixel: append( // transparent 1x1 pixel
			[]byte("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC"),
			[]byte("0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII=")...,
		),
		PixelMimetype:     "image/png",
		ExternalConnector: connector,
	}
}

func (tracker *EmailTracker) ExtractMetadata(r *http.Request) *MailMetadata {
	return &MailMetadata{}
}

func (tracker *EmailTracker) ServeTrackingPixelHandler(w http.ResponseWriter, r *http.Request) {
	// handle logging
	tracker.LogRequest(r)
	defer tracker.LogResponse(w)

	// only GET allowed
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// invoke external connector
	if tracker.ExternalConnector != nil {
		tracker.ExternalConnector.NotifyExternal(tracker.ExtractMetadata(r))
	}

	// write tracking pixel
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "image/png")
	w.Write(tracker.Pixel)
}
