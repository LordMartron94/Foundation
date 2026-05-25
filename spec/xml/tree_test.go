package xml

import "testing"

const fixtureXML = `<root comment="root note">
  <child name="alpha">
    <comment>child doc</comment>
  </child>
  <child name="beta" />
</root>`

func TestTreeParseRoot(t *testing.T) {
	root, err := TreeParse([]byte(fixtureXML))
	if err != nil {
		t.Fatalf("TreeParse: %v", err)
	}
	if root.Name != "root" {
		t.Fatalf("root name = %q, want root", root.Name)
	}
	if root.Attr("comment") != "root note" {
		t.Fatalf("root comment attr = %q", root.Attr("comment"))
	}
	if len(root.Children) != 2 {
		t.Fatalf("root children = %d, want 2", len(root.Children))
	}
}

func TestNodeChildAndChildren(t *testing.T) {
	root, err := TreeParse([]byte(fixtureXML))
	if err != nil {
		t.Fatalf("TreeParse: %v", err)
	}
	if child := root.Child("child"); child == nil {
		t.Fatal("Child(child) = nil")
	}
	if child := root.Child("missing"); child != nil {
		t.Fatal("Child(missing) should be nil")
	}
	if children := root.NodesNamed("child"); len(children) != 2 {
		t.Fatalf("NodesNamed(child) = %d, want 2", len(children))
	}
}

func TestNodeWalk(t *testing.T) {
	root, err := TreeParse([]byte(fixtureXML))
	if err != nil {
		t.Fatalf("TreeParse: %v", err)
	}
	var names []string
	root.Walk(func(node *Node) {
		names = append(names, node.Name)
	})
	if len(names) != 4 {
		t.Fatalf("walk visited %d nodes, want 4: %v", len(names), names)
	}
}

func TestNodeDirectComments(t *testing.T) {
	root, err := TreeParse([]byte(fixtureXML))
	if err != nil {
		t.Fatalf("TreeParse: %v", err)
	}
	if doc := root.DirectComments(); doc != "root note" {
		t.Fatalf("root DirectComments = %q", doc)
	}
	child := root.NodesNamed("child")[0]
	if doc := child.DirectComments(); doc != "child doc" {
		t.Fatalf("child DirectComments = %q", doc)
	}
}

func TestTreeDebugRender(t *testing.T) {
	root, err := TreeParse([]byte(fixtureXML))
	if err != nil {
		t.Fatalf("TreeParse: %v", err)
	}
	rendered := TreeDebugRender("fixture.xml", "out.txt", root, DebugDefaultOptions("test"))
	if rendered == "" {
		t.Fatal("TreeDebugRender returned empty string")
	}
}
