package validate

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var allowedHTMLTags = map[string]bool{
	"b": true, "strong": true, "i": true, "em": true, "u": true,
	"s": true, "mark": true, "small": true, "sub": true, "sup": true,
	"span": true, "br": true, "p": true, "ul": true, "ol": true,
	"li": true, "h1": true, "h2": true, "h3": true, "h4": true,
	"h5": true, "h6": true,
}

var allowedHTMLAttributes = map[string]bool{
	"style": true,
	"class": true,
}

// SafeHTML validates HTML from WYSIWYG editors. It permits only formatting tags
// and style/class attributes, rejects event handlers and common script-capable
// CSS forms.
func SafeHTML(value string) error {
	if value == "" {
		return nil
	}
	// Parse as a fragment in <body> context so the parser does not inject
	// implicit html/head/body wrapper elements around the input.
	nodes, err := html.ParseFragment(strings.NewReader(value), &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	})
	if err != nil {
		return fmt.Errorf("mengandung karakter yang tidak diperbolehkan")
	}

	var validateNode func(*html.Node) error
	validateNode = func(n *html.Node) error {
		if n.Type == html.ElementNode {
			tag := strings.ToLower(n.Data)
			if !allowedHTMLTags[tag] {
				return fmt.Errorf("contain not allowed tag <%s>", tag)
			}
			for _, attr := range n.Attr {
				name := strings.ToLower(attr.Key)
				if strings.HasPrefix(name, "on") {
					return fmt.Errorf("contain restricted event (%s)", name)
				}
				if !allowedHTMLAttributes[name] {
					return fmt.Errorf("containt not allowed attribute %s", name)
				}
				if name == "style" && styleIsUnsafe(attr.Val) {
					return fmt.Errorf("contain unsafe script")
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			if err := validateNode(child); err != nil {
				return err
			}
		}
		return nil
	}
	for _, n := range nodes {
		if err := validateNode(n); err != nil {
			return err
		}
	}
	return nil
}

func styleIsUnsafe(style string) bool {
	lower := strings.ToLower(style)
	for _, pattern := range []string{
		"expression(",
		"javascript:",
		"vbscript:",
		"data:",
		"-moz-binding",
	} {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}
