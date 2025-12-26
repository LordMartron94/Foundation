package link

import (
	"fmt"
	"regexp"
	"strings"
)

// LinkComponents represents the parsed components of a link.
type LinkComponents struct {
	Full     string // Original match
	UUID     string // Optional UUID
	Title    string // Link title/text
	Location string // Optional location (#...)
	Override string // Optional override text (|...)
}

// LinkParseBracketedLink parses a bracketed link in the format [[content]]
// and extracts its components (UUID, title, location, override text).
//
// Supported formats:
//   - [[title]]
//   - [[uuid:title]]
//   - [[title|override]]
//   - [[title#location]]
//   - [[uuid:title#location]]
//   - [[title#location|override]]
//   - [[uuid:title#location|override]]
//
// Returns an error if the link format is invalid.
func LinkParseBracketedLink(match string) (LinkComponents, error) {
	// Remove brackets
	if !strings.HasPrefix(match, "[[") || !strings.HasSuffix(match, "]]") {
		return LinkComponents{}, fmt.Errorf("invalid bracketed link format: %q", match)
	}

	content := match[2 : len(match)-2]
	if content == "" {
		return LinkComponents{}, fmt.Errorf("empty link content")
	}

	result := LinkComponents{
		Full: match,
	}

	// Extract override text (|...) first
	if pipeIdx := strings.Index(content, "|"); pipeIdx >= 0 {
		result.Override = content[pipeIdx+1:]
		content = content[:pipeIdx]
	}

	// Extract location (#...)
	if hashIdx := strings.Index(content, "#"); hashIdx >= 0 {
		result.Location = content[hashIdx+1:]
		content = content[:hashIdx]
	}

	// Check if it's UUID format: uuid:title
	if colonIdx := strings.Index(content, ":"); colonIdx >= 0 {
		uuid := content[:colonIdx]
		title := content[colonIdx+1:]
		// Validate UUID format (basic check - should be 36 chars with hyphens)
		if LinkValidateUUIDFormat(uuid) {
			result.UUID = uuid
			result.Title = title
		} else {
			// Not a valid UUID, treat as regular title
			result.Title = content
		}
	} else {
		result.Title = content
	}

	return result, nil
}

// LinkExtractAllBracketedLinks extracts all bracketed links from content
// matching the provided UUID pattern. The uuidPattern should be a regex pattern
// for matching UUIDs (e.g., `[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`).
//
// Returns a slice of LinkComponents for all found links.
func LinkExtractAllBracketedLinks(content, uuidPattern string) []LinkComponents {
	allLinks := make([]LinkComponents, 0)

	// Pattern for UUID links: [[uuid:title]]
	uuidLinkPattern := regexp.MustCompile(fmt.Sprintf(`\[\[(%s):([^\]]+)\]\]`, uuidPattern))
	uuidMatches := uuidLinkPattern.FindAllStringSubmatch(content, -1)

	for _, match := range uuidMatches {
		if len(match) >= 3 {
			components, err := LinkParseBracketedLink(match[0])
			if err == nil {
				allLinks = append(allLinks, components)
			}
		}
	}

	// Pattern for legacy links: [[title]]
	legacyLinkPattern := regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	legacyMatches := legacyLinkPattern.FindAllStringSubmatch(content, -1)

	seenUUIDs := make(map[string]bool)
	for _, link := range allLinks {
		if link.UUID != "" {
			seenUUIDs[link.UUID] = true
		}
	}

	for _, match := range legacyMatches {
		if len(match) < 2 {
			continue
		}

		// Skip if already processed as UUID link
		content := match[1]
		if strings.Contains(content, ":") {
			parts := strings.SplitN(content, ":", 2)
			if len(parts) == 2 && LinkValidateUUIDFormat(parts[0]) {
				// Already processed as UUID link
				continue
			}
		}

		components, err := LinkParseBracketedLink(match[0])
		if err == nil {
			// Only add if UUID not already seen (for legacy links that might resolve to same UUID)
			if components.UUID == "" || !seenUUIDs[components.UUID] {
				allLinks = append(allLinks, components)
				if components.UUID != "" {
					seenUUIDs[components.UUID] = true
				}
			}
		}
	}

	return allLinks
}

// LinkValidateUUIDFormat performs basic UUID format validation.
// Checks that the UUID is 36 characters long and contains exactly 4 hyphens.
func LinkValidateUUIDFormat(uuid string) bool {
	return len(uuid) == 36 && strings.Count(uuid, "-") == 4
}


