package deps

import "io"

// Versioned package
type Versioned interface {
	Name() string
	String() string
	SetVersion(ver string)
}

// VersionedPackages could be a manifest (pom, requirements.txt) of
// strictly versioned packages
type VersionedPackages interface {
	Load(r io.Reader) error
	Write(w io.Writer) error
	GetArtifacts() map[string]Versioned
	SetVersion(pkg, ver string) error
}
