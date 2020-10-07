package document

// Document is a generic abstraction for reading and writing changelogs.
type Document interface {
	Parse() error
	Write() error
}
