package emailtracker

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func (tracker *EmailTracker) WrapTrackingLink() {

}

func (tracker *EmailTracker) AppendTrackingPixelURL(
	rawHTML string,
	metadata *MailMetadata,
) (string, error) {

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
	doc.AppendChild(&html.Node{
		Parent:      nil,
		FirstChild:  nil,
		LastChild:   nil,
		PrevSibling: nil,
		NextSibling: nil,
		Type:        html.ElementNode,
		DataAtom:    atom.Img,
		Data:        "img",
		Attr:        []html.Attribute{},
	})

	// write to buffer and return as string
	var b bytes.Buffer
	err = html.Render(&b, doc)
	if err != nil {
		return rawHTML, err
	}
	return b.String(), nil
}
