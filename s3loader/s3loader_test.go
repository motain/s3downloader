package s3loader

import (
	"reflect"
	"testing"

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
