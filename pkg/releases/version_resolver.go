package releases

type VersionResolver interface {
	// ResolveVersion returns the version string for a given tag. As a special
	// case, if the tag is "latest", the latest version should be returned.
	// Otherwise, it sanitizes the input tag into a valid version string
	// without the 'v' prefix.
	ResolveVersion(tag string) (string, error)
}
