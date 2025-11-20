package bincache

type BinCache interface {
	Path() string
	BinPath() (string, error)
	Clean() error
}
