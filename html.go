package emailtracker

import (
	"bytes"
	"io"
	"net/http"
	"strings"

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

func (tracker *EmailTracker) ServeAppendPixelHandler(w http.ResponseWriter, r *http.Request) {
	// handle logging
	tracker.Logger.LogRequest(r)
	defer tracker.Logger.LogResponse(w)

	// only POST and context-type: text/html allowed
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("Content-Type") != "text/html" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// parse html body
	rawHTML, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// extract metadata and append
	metadata, err := tracker.ExtractMetadata(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	modifiedHTML, err := tracker.AppendPixelToHTML(string(rawHTML), metadata)
	if err != nil {
		tracker.Logger.LogEndpointError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// write tracking pixel
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(modifiedHTML))
}
