package location

import (
	"fmt"
	"foundation/extensions"
	"net/url"
	"strings"
)

/*
Location is a generic identifier of resources.

It takes a lot of inspiration from the URI according to spec: https://www.rfc-editor.org/rfc/rfc3986

The main difference is that the Location struct has a separate Coordinates map
so that the main identifier does not get polluted with specific details such as line/column.

While most attributes can be directly accessed on the location through their accessor methods,
the coordinates map needs to be interacted with using the freeform functions `LocationCoordinateGet` and `LocationCoordinateGetAs`.
This prevents issues with accidental modification of the underlying map (without extra allocations by copying), and manual type conversions.
*/
type Location struct {
	scheme      string
	authority   string
	path        string
	query       string
	fragment    string
	coordinates map[string]any
}

/*
Scheme returns the location's scheme.
*/
func (l Location) Scheme() string {
	return l.scheme
}

/*
Authority returns the location's authority.
*/
func (l Location) Authority() string {
	return l.authority
}

/*
Path returns the location's path.
*/
func (l Location) Path() string {
	return l.path
}

/*
Query returns the location's query.
*/
func (l Location) Query() string {
	return l.query
}

/*
Fragment returns the location's fragment.
*/
func (l Location) Fragment() string {
	return l.fragment
}

/*
LocationCoordinateGet returns the raw value associated with the given coordinate key.

If the value does not exist, it returns an error.
*/
func LocationCoordinateGet(location Location, coordinateKey string) (value any, error error) {
	val, exist := location.coordinates[coordinateKey]
	if !exist {
		return nil, fmt.Errorf("coordinate map does not have key '%s'", coordinateKey)
	}

	return val, nil
}

/*
LocationCoordinateGetAs is a convenience wrapper around `LocationCoordinateGet` which also converts the type.

If the provided value type is different from the stored value, an error is returned.
*/
func LocationCoordinateGetAs[TValue any](location Location, coordinateKey string) (TValue, error) {
	var zero TValue

	raw, err := LocationCoordinateGet(location, coordinateKey)
	if err != nil {
		return zero, err
	}

	casted, ok := raw.(TValue)
	if !ok {
		return zero, fmt.Errorf("location key '%s': expected type %T, got %T", coordinateKey, zero, raw)
	}

	return casted, nil
}

/*
LocationCreate constructs a location with the provided data.
It strictly validates the scheme against RFC 3986 rules.
*/
func LocationCreate(scheme, authority, path, query, fragment string, coordinates map[string]any) Location {
	if !isValidURIScheme(scheme) {
		panic(fmt.Errorf("invalid URI scheme '%s': must start with a letter and contain only letters, digits, '+', '-', or '.'", scheme))
	}

	if coordinates == nil {
		coordinates = make(map[string]any)
	}

	return Location{
		scheme:      scheme,
		authority:   authority,
		path:        path,
		query:       query,
		fragment:    fragment,
		coordinates: coordinates,
	}
}

/*
LocationCoordinateValueFormatter returns a string representation of the value associated with the given key.
*/
type LocationCoordinateValueFormatter func(coordinateKey string, coordinateValue any) string

/*
LocationSerializeURI serializes the location strictly according to RFC 3986 Section 5.3.
It ignores diagnostic coordinates to ensure the output is a perfectly valid network identifier.
*/
func LocationSerializeURI(location Location) string {
	builder := &strings.Builder{}

	builder.WriteString(location.scheme)
	builder.WriteRune(':')

	if location.authority != "" {
		builder.WriteString("//")
		builder.WriteString(location.authority)
	}

	builder.WriteString(location.path)

	if location.query != "" {
		builder.WriteRune('?')
		builder.WriteString(location.query)
	}

	if location.fragment != "" {
		builder.WriteRune('#')
		builder.WriteString(location.fragment)
	}

	return builder.String()
}

/*
LocationDebug returns a human-readable diagnostic representation of the location,
combining the strict URI with formatted spatial coordinates.
*/
func LocationDebug(location Location, coordinateValueFormatter LocationCoordinateValueFormatter) string {
	uriBase := LocationSerializeURI(location)

	if len(location.coordinates) == 0 {
		return uriBase
	}

	builder := &strings.Builder{}
	builder.WriteString(uriBase)

	builder.WriteString(" @ [")

	comparator := extensions.OrderedCMPGenerator[string](false) // ascending
	sorted := extensions.MapSortFunc(location.coordinates, func(pairA, pairB extensions.KeyValuePair[string, any]) int {
		return comparator(pairA.Key, pairB.Key)
	})

	amount := len(sorted)
	for i := 0; i < amount; i++ {
		item := sorted[i]

		if i > 0 {
			builder.WriteString(", ")
		}

		encodedKey := url.QueryEscape(item.Key)
		formattedValue := coordinateValueFormatter(item.Key, item.Value)
		encodedValue := url.QueryEscape(formattedValue)

		builder.WriteString(encodedKey)
		builder.WriteRune('=')
		builder.WriteString(encodedValue)
	}

	builder.WriteRune(']')
	return builder.String()
}

// ------------------------------------------------------------ PRIVATE HELPERS

/*
isValidURIScheme validates a scheme against RFC 3986 Section 3.1.
scheme = ALPHA *( ALPHA / DIGIT / "+" / "-" / "." )
*/
func isValidURIScheme(s string) bool {
	if s == "" {
		return false
	}

	// Must start with an ALPHA
	first := s[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')) {
		return false
	}

	for i := 1; i < len(s); i++ {
		c := s[i]
		if !((c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '+' || c == '-' || c == '.') {
			return false
		}
	}
	return true
}
