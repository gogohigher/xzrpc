package compressor

const (
	NOCompressor = iota
	Gzip
	Snappy
)

var Compressors = map[byte]Compressor{
	Gzip:   &GzipCompressor{},
	Snappy: &SnappyCompressor{},
}

type Compressor interface {
	Zip(data []byte) ([]byte, error)
	UnZip(data []byte) ([]byte, error)
}
