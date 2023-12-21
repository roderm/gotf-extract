package server

import (
	"archive/tar"
	"io"

	"github.com/xi2/xz"
)

type extractor interface {
	Extract(io.Writer, io.Reader) (int64, error)
	Filename() string
}

type XZExtractor struct {
	filename string
}

func (x *XZExtractor) Extract(w io.Writer, r io.Reader) (int64, error) {
	xzr, err := xz.NewReader(r, 0)
	if err != nil {
		return 0, err
	}
	return io.Copy(w, xzr)
}
func (x *XZExtractor) Filename() string {
	return x.filename
}

type TarExtractor struct {
	filename string
}

func (x *TarExtractor) Extract(w io.Writer, r io.Reader) (int64, error) {
	tr := tar.NewReader(r)
	return io.Copy(w, tr)
}
func (x *TarExtractor) Filename() string {
	return x.filename
}
