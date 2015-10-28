package s3loader

import (
	"errors"
	"io"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/motain/s3downloader/cfg"
)

func TestNewDownloader(t *testing.T) {
	testCases := []struct {
		args                 *cfg.InArgs
		conf                 *cfg.Cfg
		hasError             bool
		downloaderInstanceOf interface{}
	}{
		{
			args:                 nil,
			conf:                 nil,
			hasError:             true,
			downloaderInstanceOf: nil,
		},
		{
			args:                 nil,
			conf:                 &cfg.Cfg{},
			hasError:             true,
			downloaderInstanceOf: nil,
		},
		{
			args:                 &cfg.InArgs{},
			conf:                 nil,
			hasError:             true,
			downloaderInstanceOf: nil,
		},
		{
			args:                 &cfg.InArgs{},
			conf:                 &cfg.Cfg{},
			hasError:             true,
			downloaderInstanceOf: nil,
		},
		{
			args:                 &cfg.InArgs{},
			conf:                 &cfg.Cfg{"test access key", "test secret key", "test region"},
			hasError:             false,
			downloaderInstanceOf: &Downloader{},
		},
	}

	errfmt := "Test case: %d. Expected %s: %v. Got: %v"

	for i, testCase := range testCases {
		obtainedDownloader, err := NewDownloader(testCase.args, testCase.conf)
		if (err != nil) != testCase.hasError {
			t.Errorf(errfmt, i+1, "has error", testCase.hasError, err != nil)
		}

		if testCase.downloaderInstanceOf == nil && obtainedDownloader != nil {
			t.Errorf(errfmt, i+1, "downloaderInstanceOf", testCase.downloaderInstanceOf, obtainedDownloader)
		}

		if testCase.downloaderInstanceOf == nil {
			continue
		}

		if testCase.downloaderInstanceOf == nil && reflect.TypeOf(obtainedDownloader) != reflect.TypeOf(testCase.downloaderInstanceOf) {
			t.Errorf(errfmt, i+1, "downloaderInstanceOf", reflect.TypeOf(testCase.downloaderInstanceOf), reflect.TypeOf(obtainedDownloader))

		}
	}
}

func TestDownloaderRunDownloadSuccess(t *testing.T) {
	input := &cfg.InArgs{
		Bucket:   "testbucket",
		Regexp:   ".*",
		LocalDir: "testfixtures",
	}

	mockDownloadManager := &mockDownloadManager{
		downloadFunc: func(w io.WriterAt, params *s3.GetObjectInput) (int64, error) {
			return 5, nil
		},
	}

	mockPageLister := &mockPageLister{
		listObjectsPagesFunc: func(params *s3.ListObjectsInput, pageIterator func(*s3.ListObjectsOutput, bool) bool) error {
			now := time.Now().UTC()
			s3output := &s3.ListObjectsOutput{
				Contents: []*s3.Object{
					{Key: aws.String("testkey"), LastModified: &now},
				},
			}

			pageIterator(s3output, true)
			return nil
		},
	}

	mockFCreator := &mockFileCreator{
		createFunc: func(fname string) (*os.File, error) { return nil, nil },
	}

	d, _ := NewDownloader(input, &cfg.Cfg{"test access key", "test secret key", "test region"})
	d.downloadManager = mockDownloadManager
	d.pageLister = mockPageLister
	d.fileCreator = mockFCreator

	err := d.Run()
	if err != nil {
		t.Errorf("Expected nil error. Got: %s", err)
	}
}

func TestDownloaderRunDownloadError(t *testing.T) {
	input := &cfg.InArgs{
		Bucket:   "testbucket",
		Regexp:   ".*",
		LocalDir: "testfixtures",
	}

	expectedErr := errors.New("test error")
	mockPageLister := &mockPageLister{
		listObjectsPagesFunc: func(params *s3.ListObjectsInput, pageIterator func(*s3.ListObjectsOutput, bool) bool) error {
			return expectedErr
		},
	}

	d, _ := NewDownloader(input, &cfg.Cfg{"test access key", "test secret key", "test region"})
	d.pageLister = mockPageLister

	obtainedErr := d.Run()
	if obtainedErr != expectedErr {
		t.Errorf("Expected err: %s. Got: %s", expectedErr, obtainedErr)
	}
}
