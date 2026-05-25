/*
Package xml provides a protocol-agnostic XML document tree for spec-driven code generators.

[Context]
Builds a mutable element tree from encoding/xml tokens without schema-specific interpretation.
Consumers (Vulkan registry, Wayland protocol XML, xcb-proto) walk the tree to construct their own IR.
*/
package xml
