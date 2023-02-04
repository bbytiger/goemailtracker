package emailtracker

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func (tracker *EmailTracker) WrapTrackingLink() { // TODO
}

func (tracker *EmailTracker) AppendPixelToHTML(
	rawHTML string,
	metadata *MailMetadata,
) (string, error) {
	// parse input
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return rawHTML, err
	}
	if doc.Parent != nil || doc.PrevSibling != nil || doc.NextSibling != nil {
		// if parents OR siblings, create an outer div
		node := &html.Node{
			Parent:      nil,
			FirstChild:  nil,
			LastChild:   nil,
			PrevSibling: nil,
			NextSibling: nil,
			Type:        html.ElementNode,
			DataAtom:    atom.Div,
			Data:        "div",
			Attr:        []html.Attribute{},
		}
		node.AppendChild(doc)
		doc = node
	}

	// append pixel
	srcURL, err := tracker.GetURLFromPII(metadata.SenderInfo)
	if err != nil {
		return rawHTML, err
	}
	doc.AppendChild(&html.Node{
		Parent:      nil,
		FirstChild:  nil,
		LastChild:   nil,
		PrevSibling: nil,
		NextSibling: nil,
		Type:        html.ElementNode,
		DataAtom:    atom.Img,
		Data:        "img",
		Attr: []html.Attribute{
			{
				Key: "src",
				Val: srcURL.String(),
			},
		},
	})

	// write to buffer and return as string
	var b bytes.Buffer
	err = html.Render(&b, doc)
	if err != nil {
		return rawHTML, err
	}
	return b.String(), nil
}

func (tracker *EmailTracker) ServeAppendHandler(w http.ResponseWriter, r *http.Request) {
	// handle logging
	tracker.Logger.LogRequest(r)
	defer tracker.Logger.LogResponse(w)

	// only POST and context-type: application/json allowed
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("content-type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// extract metadata and append
	metadata, err := tracker.ExtractMetadata(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// check if pixel already appended
	query, queryErr := tracker.ExternalConnector.QueryTrackerStatus(metadata.SenderInfo.EmailId)
	if queryErr != nil && query == Untracked {
		// append pixel if emailId field does not exist
		modifiedHTML, err := tracker.AppendPixelToHTML(
			string(metadata.SenderInfo.HTML),
			metadata,
		)
		if err != nil {
			tracker.Logger.LogEndpointError(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// assign new email id and modified HTML
		metadata.SenderInfo.HTML = modifiedHTML
		metadata.SenderInfo.EmailId = uuid.New().String()
	}

	// invoke external connector
	connectorErr := tracker.ExternalConnector.NotifyExternal(metadata)
	if connectorErr != nil {
		tracker.Logger.LogEndpointError(connectorErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// write tracking pixel
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(metadata.SenderInfo.HTML))
}
