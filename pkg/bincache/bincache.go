package bincache

type BinCache interface {
	BinPath(version string) (string, error)
	CacheDir() string
	Clean() error
}
