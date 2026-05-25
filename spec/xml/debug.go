package xml

import (
	"fmt"
	"strings"
)

const (
	defaultDebugMaxChildrenListed uint64 = 40
	defaultDebugMaxTextRunes             = 120
)

/*
DebugOptions controls how much of a large document is expanded in debug output.
*/
type DebugOptions struct {
	Title             string
	MaxChildrenListed uint64
	MaxTextRunes      int
}

/*
DebugDefaultOptions returns limits suited to large registry XML debug dumps.
*/
func DebugDefaultOptions(title string) DebugOptions {
	return DebugOptions{
		Title:             title,
		MaxChildrenListed: defaultDebugMaxChildrenListed,
		MaxTextRunes:      defaultDebugMaxTextRunes,
	}
}

/*
TreeDebugRender formats a parsed tree and a short structural summary for human inspection.

[Parameters]
sourcePath labels the XML in the dump header. outputPath labels the debug file path in the header.
root is the parsed tree from TreeParse.

[Returns]
A multi-line text dump. Does not mutate root.
*/
func TreeDebugRender(sourcePath string, outputPath string, root *Node, options DebugOptions) string {
	var builder strings.Builder

	title := strings.TrimSpace(options.Title)
	if title == "" {
		title = "XML tree (debug)"
	}

	builder.WriteString(title + "\n")
	builder.WriteString(fmt.Sprintf("Source: %s\n", sourcePath))
	builder.WriteString(fmt.Sprintf("Debug output: %s\n", outputPath))
	builder.WriteString("\n")
	builder.WriteString("=== Document summary (depth-1 sections) ===\n")
	treeSummaryWrite(&builder, root, options.MaxTextRunes)
	builder.WriteString("\n")
	builder.WriteString("=== Element tree ===\n")
	treeNodeWrite(&builder, root, 0, options)
	builder.WriteString("\n")

	return builder.String()
}

func treeSummaryWrite(builder *strings.Builder, root *Node, maxTextRunes int) {
	if root == nil {
		builder.WriteString("(empty document)\n")
		return
	}

	builder.WriteString(fmt.Sprintf("Root: <%s> (%d top-level children)\n", root.Name, len(root.Children)))
	for _, child := range root.Children {
		builder.WriteString(fmt.Sprintf("  <%s>", child.Name))
		if len(child.Attrs) > 0 {
			builder.WriteString(" ")
			builder.WriteString(attrsInline(child.Attrs))
		}
		builder.WriteString(fmt.Sprintf(" → %d children", len(child.Children)))
		if text := textTruncate(child.Text, maxTextRunes); text != "" {
			builder.WriteString(fmt.Sprintf(" text=%q", text))
		}
		builder.WriteByte('\n')
	}
}

func treeNodeWrite(builder *strings.Builder, node *Node, depth int, options DebugOptions) {
	if node == nil {
		return
	}

	indent := strings.Repeat("  ", depth)
	builder.WriteString(indent)
	builder.WriteString(fmt.Sprintf("<%s>", node.Name))

	if len(node.Attrs) > 0 {
		builder.WriteString(" ")
		builder.WriteString(attrsInline(node.Attrs))
	}

	childCount := len(node.Children)
	if childCount > 0 {
		builder.WriteString(fmt.Sprintf(" children=%d", childCount))
	}

	if text := textTruncate(node.Text, options.MaxTextRunes); text != "" {
		builder.WriteString(fmt.Sprintf(" text=%q", text))
	}
	builder.WriteByte('\n')

	listed := childCount
	if options.MaxChildrenListed > 0 && uint64(childCount) > options.MaxChildrenListed {
		listed = int(options.MaxChildrenListed)
	}

	for i := 0; i < listed; i++ {
		treeNodeWrite(builder, node.Children[i], depth+1, options)
	}

	if listed < childCount {
		builder.WriteString(indent)
		builder.WriteString("  ")
		builder.WriteString(fmt.Sprintf("... %d more <%s> children not listed\n", childCount-listed, node.Name))
	}
}

func attrsInline(attrs []Attr) string {
	parts := make([]string, len(attrs))
	for i, attr := range attrs {
		parts[i] = fmt.Sprintf("%s=%q", attr.Name, attr.Value)
	}
	return strings.Join(parts, " ")
}

func textTruncate(text string, maxRunes int) string {
	text = strings.TrimSpace(text)
	if text == "" || maxRunes <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	return string(runes[:maxRunes]) + "..."
}
