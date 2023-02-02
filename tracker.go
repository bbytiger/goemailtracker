package emailtracker

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"
)

type Action string

const (
	AppendPixel Action = "appendPixel"
	ServePixel  Action = "servePixel"
)

type Status string

const (
	Untracked Status = "untracked"
	Attached  Status = "attached"
	Opened    Status = "opened"
	Responded Status = "responded"
)

type MailMetadata struct {
	Timestamp    time.Time
	UserAgent    string
	UserIP       string
	Action       Action
	StatusUpdate Status
	HTML         string
	SenderInfo   *MailPII
}

// sender's personal identifying info
type MailPII struct {
	SenderId    string `json:"sender_id"`
	SenderEmail string `json:"sender_email"`
	RecvEmail   string `json:"recv_email"`
	EmailId     string `json:"email_id"`
}

type EmailTracker struct {
	BaseURL         *url.URL
	ActionToURLPath map[Action]string
	Pixel           []byte
	PixelMimetype   string
	Encoder
	ExternalConnector
	Logger
}

func NewEmailTracker(
	baseUrl string,
	appendPath string,
	servePath string,
	encoder Encoder,
	connector ExternalConnector,
	logger Logger,
) *EmailTracker {
	baseURL, err := url.Parse(baseUrl)
	if err != nil {
		panic(err)
	}
	return &EmailTracker{
		BaseURL: baseURL,
		ActionToURLPath: map[Action]string{
			AppendPixel: appendPath,
			ServePixel:  servePath,
		},
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
	baseURL.Path += tracker.ActionToURLPath[ServePixel]
	baseURL.RawQuery = params.Encode()

	return baseURL, nil
}

func (tracker *EmailTracker) ExtractMetadata(r *http.Request) (*MailMetadata, error) {
	// get PII
	if tracker.ActionToURLPath[AppendPixel] == r.URL.Path {
		return nil, errors.New("ExtractMetadata cannot be called on AppendPixel action")
	}
	pii, err := tracker.GetPIIFromQueryParams(r.URL)
	if err != nil {
		tracker.Logger.LogPkgError(err)
		return nil, err
	}

	// find action based on path
	action := Action("")
	for key, path := range tracker.ActionToURLPath {
		if path == r.URL.Path {
			action = key
		}
	}
	if action == Action("") {
		actionErr := errors.New("failure parsing action from url path")
		tracker.Logger.LogPkgError(actionErr)
		return nil, actionErr
	}

	return &MailMetadata{
		Timestamp:  time.Now(),
		UserAgent:  r.Header.Get("User-Agent"),
		UserIP:     r.Header.Get("X-Forwarded-For"),
		Action:     action,
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
	metadata, err := tracker.ExtractMetadata(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	connectorErr := tracker.ExternalConnector.NotifyExternal(metadata)
	if connectorErr != nil {
		tracker.Logger.LogEndpointError(connectorErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// write tracking pixel
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "image/png")
	w.Write(tracker.Pixel)
}
