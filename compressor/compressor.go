package compressor

const (
	NOCompressor = iota
	Gzip
)

var Compressors = map[byte]Compressor{
	Gzip: &GzipCompressor{},
}

type Compressor interface {
	Zip(data []byte) ([]byte, error)
	UnZip(data []byte) ([]byte, error)
}
