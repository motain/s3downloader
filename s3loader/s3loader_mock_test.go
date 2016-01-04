package s3loader

import (
	"io"
	"os"

	"github.com/aws/aws-sdk-go/service/s3"
)

type mockFileCreator struct {
	createFunc func(fname string) (*os.File, error)
}

func (fc mockFileCreator) Create(fname string) (*os.File, error) {
	return fc.createFunc(fname)
}

type mockDownloadManager struct {
	downloadFunc func(io.WriterAt, *s3.GetObjectInput) (int64, error)
}

func (d *mockDownloadManager) Download(w io.WriterAt, params *s3.GetObjectInput) (int64, error) {
	return d.downloadFunc(w, params)
}

type mockPageLister struct {
	listObjectsPagesFunc func(params *s3.ListObjectsInput, pageIterator func(*s3.ListObjectsOutput, bool) bool) error
}

func (l *mockPageLister) ListObjectsPages(params *s3.ListObjectsInput, pageIterator func(*s3.ListObjectsOutput, bool) bool) error {
	return l.listObjectsPagesFunc(params, pageIterator)
}
