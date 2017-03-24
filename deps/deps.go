package deps

import "io"

type Versioned interface {
	String() string
	SetVersion(ver string)
}

type VersionedPackages interface {
	Load(r io.Reader) error
	Write(w io.Writer) error
}
