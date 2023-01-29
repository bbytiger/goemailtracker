package emailtracker

import (
	"net/http"
	"time"
)

type EmailTracker struct {
	WebhookURL    string // support tracking links later
	Encoding      string
	Pixel         []byte
	PixelMimetype string
	ExternalURL   string
}

type TrackingMetadata struct {
	Timestamp time.Time
}

func NewEmailTracker() *EmailTracker {
	return &EmailTracker{
		Pixel: append( // transparent 1x1 pixel
			[]byte("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC"),
			[]byte("0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII=")...,
		),
		PixelMimetype: "image/png",
	}
}

func (tracker *EmailTracker) ExtractMetadata(r *http.Request) *TrackingMetadata {
	return &TrackingMetadata{}
}

func (tracker *EmailTracker) ServeTrackingPixelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	tracker.NotifyExternal(tracker.ExtractMetadata(r))
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "image/png")
	w.Write(tracker.Pixel)
}

func (tracker *EmailTracker) NotifyExternal(metadata *TrackingMetadata) {
	// modify to suit your own implementation
}
