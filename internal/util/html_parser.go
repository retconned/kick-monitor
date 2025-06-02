package util

import (
	"errors"
	"strings"

	"golang.org/x/net/html"
)

// ExtractJSONFromHTML attempts to find and extract a JSON string from HTML content.
func ExtractJSONFromHTML(htmlString string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		return "", err
	}

	var jsonString string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			// Simple check: if the text looks like a JSON object start or array start
			if strings.HasPrefix(text, "{") || strings.HasPrefix(text, "[") {
				jsonString = text
				return // Stop searching once potential JSON is found
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
			if jsonString != "" { // Stop if JSON found in a child
				return
			}
		}
	}

	f(doc)

	if jsonString == "" {
		return "", errors.New("JSON not found in HTML")
	}

	return jsonString, nil
}
