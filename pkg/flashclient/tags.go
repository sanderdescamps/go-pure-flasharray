package flashclient

// Tag represents a tag that can be applied to Pure Storage resources.
// Tags support optional namespacing and can be configured for inclusion
// in resource copying operations.
type Tag struct {
	// Copyable specifies whether the tag is included when copying the parent resource.
	// If true, the tag is included in resource copying.
	// If false, the tag is not included.
	// Defaults to true if not specified.
	Copyable *bool `json:"copyable,omitempty"`

	// Key is the identifier of the tag.
	// Supports up to 64 Unicode characters.
	Key *string `json:"key,omitempty"`

	// Namespace is the optional category of the tag.
	// Omitting the namespace defaults to "default".
	// The "pure*" namespaces are reserved for plugins and integration partners.
	// It is recommended that customers avoid using reserved namespaces.
	Namespace *string `json:"namespace,omitempty"`

	// Resource is a reference to the resource that this tag is applied to.
	Resource *FixedReference `json:"resource,omitempty"`

	// Value is the content of the tag.
	// Supports up to 256 Unicode characters.
	Value *string `json:"value,omitempty"`
}
