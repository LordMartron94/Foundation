package link

import "testing"

func TestLinkValidateUUIDFormat(t *testing.T) {
	tests := []struct {
		uuid     string
		expected bool
	}{
		{"123e4567-e89b-12d3-a456-426614174000", true},
		{"123e4567-e89b-12d3-a456", false},
		{"123e4567e89b12d3a456426614174000", false},
		{"", false},
		{"123e4567-e89b-12d3-a456-42661417400", false},
		{"123e4567-e89b-12d3-a456-4266141740000", false},
	}

	for _, tt := range tests {
		result := LinkValidateUUIDFormat(tt.uuid)
		if result != tt.expected {
			t.Errorf("LinkValidateUUIDFormat(%q) = %v, expected %v",
				tt.uuid, result, tt.expected)
		}
	}
}

func TestLinkParseBracketedLink(t *testing.T) {
	tests := []struct {
		match     string
		expected  LinkComponents
		shouldErr bool
	}{
		{
			"[[title]]",
			LinkComponents{Full: "[[title]]", Title: "title"},
			false,
		},
		{
			"[[uuid:title]]",
			LinkComponents{Full: "[[uuid:title]]", UUID: "uuid", Title: "title"},
			false,
		},
		{
			"[[title|override]]",
			LinkComponents{Full: "[[title|override]]", Title: "title", Override: "override"},
			false,
		},
		{
			"[[title#location]]",
			LinkComponents{Full: "[[title#location]]", Title: "title", Location: "location"},
			false,
		},
		{
			"[[uuid:title#location|override]]",
			LinkComponents{
				Full:     "[[uuid:title#location|override]]",
				UUID:     "uuid",
				Title:    "title",
				Location: "location",
				Override: "override",
			},
			false,
		},
		{
			"[title]",
			LinkComponents{},
			true,
		},
		{
			"[[",
			LinkComponents{},
			true,
		},
	}

	for _, tt := range tests {
		result, err := LinkParseBracketedLink(tt.match)
		if tt.shouldErr {
			if err == nil {
				t.Errorf("LinkParseBracketedLink(%q) expected error, got nil", tt.match)
			}
		} else {
			if err != nil {
				t.Errorf("LinkParseBracketedLink(%q) unexpected error: %v", tt.match, err)
			}
			if result.Full != tt.expected.Full ||
				result.UUID != tt.expected.UUID ||
				result.Title != tt.expected.Title ||
				result.Location != tt.expected.Location ||
				result.Override != tt.expected.Override {
				t.Errorf("LinkParseBracketedLink(%q) = %+v, expected %+v",
					tt.match, result, tt.expected)
			}
		}
	}
}

func TestLinkExtractAllBracketedLinks(t *testing.T) {
	uuidPattern := `[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`
	content := "This has [[title1]] and [[123e4567-e89b-12d3-a456-426614174000:title2]] links."
	links := LinkExtractAllBracketedLinks(content, uuidPattern)

	if len(links) < 2 {
		t.Errorf("LinkExtractAllBracketedLinks expected at least 2 links, got %d", len(links))
	}
}


