package emailtracker

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"
)

type EmailTracker struct {
	BaseURL       *url.URL
	PixelPath     string
	Pixel         []byte
	PixelMimetype string
	Encoder
	ExternalConnector
	Logger
}

type MailMetadata struct {
	Timestamp  time.Time
	UserAgent  string
	UserIP     string
	SenderInfo *MailPII
}

// sender's personal identifying info
type MailPII struct {
	SenderId    string `json:"sender_id"`
	SenderEmail string `json:"sender_email"`
	RecvEmail   string `json:"recv_email"`
	EmailId     string `json:"email_id"`
}

func NewEmailTracker(
	encoder Encoder,
	connector ExternalConnector,
	logger Logger,
) *EmailTracker {
	return &EmailTracker{
		Pixel: append( // transparent 1x1 pixel
			[]byte("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC"),
			[]byte("0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII=")...,
		),
		PixelMimetype:     "image/png",
		Encoder:           encoder,
		ExternalConnector: connector,
		Logger:            logger,
	}
}

func (tracker *EmailTracker) GetPIIFromQueryParams(url *url.URL) (*MailPII, error) {
	encodedData := url.Query().Get("tr")
	if encodedData == "" {
		return nil, errors.New("no tracking info found")
	}

	jsonBytes := []byte(tracker.Encoder.DecodeMessage([]byte(encodedData)))
	var data MailPII
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		tracker.Logger.LogPkgError(err)
		return nil, err
	}
	return &data, nil
}

func (tracker *EmailTracker) GetURLFromPII(pii *MailPII) (*url.URL, error) {
	jsonPII, err := json.Marshal(pii)
	if err != nil {
		tracker.Logger.LogPkgError(err)
		return nil, err
	}

	encodedBytes := tracker.Encoder.EncodeMessage(string(jsonPII))
	params := url.Values{}
	params.Add("tr", string(encodedBytes))

	baseURL, _ := url.Parse(tracker.BaseURL.String())
	baseURL.Path += tracker.PixelPath
	baseURL.RawQuery = params.Encode()

	return baseURL, nil
}

func (tracker *EmailTracker) ExtractMetadata(r *http.Request) (*MailMetadata, error) {
	pii, err := tracker.GetPIIFromQueryParams(r.URL)
	if err != nil {
		tracker.Logger.LogPkgError(err)
		return nil, err
	}
	return &MailMetadata{
		Timestamp:  time.Now(),
		UserAgent:  r.Header.Get("User-Agent"),
		UserIP:     r.Header.Get("X-Forwarded-For"),
		SenderInfo: pii,
	}, nil
}

func (tracker *EmailTracker) ServePixelHandler(w http.ResponseWriter, r *http.Request) {
	// handle logging
	tracker.Logger.LogRequest(r)
	defer tracker.Logger.LogResponse(w)

	// only GET allowed
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// invoke external connector
	if tracker.ExternalConnector != nil {
		metadata, err := tracker.ExtractMetadata(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		tracker.ExternalConnector.NotifyExternal(metadata)
	}

	// write tracking pixel
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "image/png")
	w.Write(tracker.Pixel)
}
