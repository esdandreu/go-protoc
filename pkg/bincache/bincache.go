package bincache

type BinCache interface {
	Protoc(version string) (string, error)
}
