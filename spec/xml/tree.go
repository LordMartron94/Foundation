package xml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

/*
Attr is one attribute on an XML element.
*/
type Attr struct {
	Name  string
	Value string
}

/*
Node is an element node in a parsed XML document tree.
*/
type Node struct {
	Name     string
	Attrs    []Attr
	Children []*Node
	Text     string
}

/*
TreeParse builds a document tree from raw XML using encoding/xml token decoding.

[Parameters]
content is the full document bytes. No schema-specific interpretation is applied.

[Returns]
The root element node and nil error on success.
*/
func TreeParse(content []byte) (*Node, error) {
	decoder := xml.NewDecoder(bytes.NewReader(content))

	var stack []*Node
	var root *Node

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decode XML: %w", err)
		}

		switch element := token.(type) {
		case xml.StartElement:
			node := &Node{
				Name:  element.Name.Local,
				Attrs: attrsFromXML(element.Attr),
			}
			if len(stack) == 0 {
				root = node
			} else {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
			}
			stack = append(stack, node)

		case xml.EndElement:
			if len(stack) == 0 {
				return nil, fmt.Errorf("decode XML: unexpected end element </%s>", element.Name.Local)
			}
			stack = stack[:len(stack)-1]

		case xml.CharData:
			if len(stack) == 0 {
				continue
			}
			text := strings.TrimSpace(string(element))
			if text == "" {
				continue
			}
			current := stack[len(stack)-1]
			if current.Text == "" {
				current.Text = text
			} else {
				current.Text += " " + text
			}
		}
	}

	if root == nil {
		return nil, fmt.Errorf("decode XML: document has no root element")
	}

	return root, nil
}

/*
Attr returns the value of attrName on node, or an empty string when absent.
*/
func (node *Node) Attr(attrName string) string {
	if node == nil {
		return ""
	}
	for _, attr := range node.Attrs {
		if attr.Name == attrName {
			return attr.Value
		}
	}
	return ""
}

/*
Child returns the first direct child element with the given name, or nil.
*/
func (node *Node) Child(name string) *Node {
	if node == nil {
		return nil
	}
	for _, child := range node.Children {
		if child.Name == name {
			return child
		}
	}
	return nil
}

/*
NodesNamed returns all direct child elements with the given name.
*/
func (node *Node) NodesNamed(name string) []*Node {
	if node == nil {
		return nil
	}
	var matched []*Node
	for _, child := range node.Children {
		if child.Name == name {
			matched = append(matched, child)
		}
	}
	return matched
}

/*
Walk visits node and every descendant in depth-first order.
*/
func (node *Node) Walk(fn func(*Node)) {
	if node == nil || fn == nil {
		return
	}
	fn(node)
	for _, child := range node.Children {
		child.Walk(fn)
	}
}

/*
DirectComments returns documentation from the comment attribute and direct comment child elements.
*/
func (node *Node) DirectComments() string {
	if node == nil {
		return ""
	}

	var parts []string
	if comment := strings.TrimSpace(node.Attr("comment")); comment != "" {
		parts = append(parts, comment)
	}

	for _, child := range node.Children {
		if child.Name != "comment" {
			continue
		}
		if text := strings.TrimSpace(child.Text); text != "" {
			parts = append(parts, text)
		}
	}

	return joinNonEmpty(parts...)
}

func attrsFromXML(attrs []xml.Attr) []Attr {
	if len(attrs) == 0 {
		return nil
	}
	out := make([]Attr, len(attrs))
	for i, attr := range attrs {
		out[i] = Attr{
			Name:  attr.Name.Local,
			Value: attr.Value,
		}
	}
	return out
}

func joinNonEmpty(parts ...string) string {
	nonEmpty := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			nonEmpty = append(nonEmpty, part)
		}
	}
	return strings.Join(nonEmpty, "\n\n")
}
