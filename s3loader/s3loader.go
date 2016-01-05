// Package s3loader wraps up aws sdk s3manager functionality
package s3loader

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/motain/s3downloader/cfg"
)

var (
	ErrInvalidArgs = errors.New("invalid arguments")

	infoLog  = log.New(os.Stdout, "INFO: ", log.LstdFlags)
	errorLog = log.New(os.Stdout, "ERR: ", log.LstdFlags)
)

type (
	// DownloadManager describes logic for saving an s3 item to disc
	DownloadManager interface {
		Download(io.WriterAt, *s3.GetObjectInput, ...func(*s3manager.Downloader)) (int64, error)
	}

	// Downloader is a wrapper for DownloadManager
	// Downloader handles additional input parameter
	// and concurrency logic
	Downloader struct {
		args            *cfg.InArgs
		regexp          *regexp.Regexp
		downloadManager DownloadManager
		pageLister      PageLister
		pageIterator    PageIterator
		fileCreator     fileCreator
		wg              sync.WaitGroup
		workers         chan int
	}
)

type (
	// PageIterator describes logic for every s3 item
	PageIterator interface {
		Iterate(*s3.ListObjectsOutput, bool) bool
	}

	// PageIteratorFunc is a PageIterator wrapper
	PageIteratorFunc func(*s3.ListObjectsOutput, bool) bool

	// PageLister describes logic for handling s3 page items
	PageLister interface {
		ListObjectsPages(params *s3.ListObjectsInput, pageIterator func(*s3.ListObjectsOutput, bool) bool) error
	}
)

// Iterate calls f(page, more)
func (f PageIteratorFunc) Iterate(page *s3.ListObjectsOutput, more bool) bool {
	return f(page, more)
}

// NewDownloader inits and returns a Downloader pointer
func NewDownloader(agrs *cfg.InArgs, conf *cfg.Cfg) (*Downloader, error) {
	if agrs == nil || conf == nil {
		return nil, ErrInvalidArgs
	}
	creds := credentials.NewStaticCredentials(conf.AWSAccessKeyID, conf.AWSSecretKey, "")
	if _, err := creds.Get(); err != nil {
		return nil, err
	}

	awsConf := &aws.Config{Credentials: creds, Region: aws.String(conf.Region)}
	sess := session.New(awsConf)
	client := s3.New(sess, awsConf)

	manager := NewS3DownloadManager(sess)

	d := &Downloader{
		downloadManager: manager,
		args:            agrs,
		pageLister:      client,
		regexp:          regexp.MustCompile(agrs.Regexp),
		workers:         make(chan int, 50),
		fileCreator:     &fsAdapter{},
	}

	d.pageIterator = d.pickPageIterator()
	return d, nil
}

// Run starts a downloader - s3 file download or search
func (d *Downloader) Run() error {
	params := &s3.ListObjectsInput{Bucket: &d.args.Bucket, Prefix: &d.args.Prefix}
	if err := d.pageLister.ListObjectsPages(params, d.pageIterator.Iterate); err != nil {
		return err
	}
	return nil
}

// NewS3DownloadManager inits with defaults and returns
// a *s3manager.Downloader
func NewS3DownloadManager(c client.ConfigProvider) *s3manager.Downloader {
	return s3manager.NewDownloader(c)
}

func (d *Downloader) pickPageIterator() PageIterator {
	itemHandler := d.onItemDownload

	if d.args.DryRun {
		itemHandler = d.onItemSearch
	}

	return d.pageIteratorFunc(itemHandler)
}

func (d *Downloader) pageIteratorFunc(f func(*s3.Object)) PageIteratorFunc {
	return func(page *s3.ListObjectsOutput, more bool) bool {
		for _, obj := range page.Contents {
			if !d.regexp.MatchString(*obj.Key) {
				continue
			}

			d.workers <- 1
			d.wg.Add(1)

			go func(obj *s3.Object) {
				f(obj)

				<-d.workers
				d.wg.Done()
			}(obj)
		}

		d.wg.Wait()

		return true
	}
}

func (d *Downloader) onItemSearch(obj *s3.Object) {
	infoLog.Printf("Found: s3://%s/%s", d.args.Bucket, *obj.Key)
}

func (d *Downloader) onItemDownload(obj *s3.Object) {
	if *obj.Size == 0 {
		return
	}

	if err := d.downloadToFile(*obj.Key, obj.LastModified); err != nil {
		errorLog.Println(err)
	}
}

type (
	// fileCreator describes file and dir create logic
	fileCreator interface {
		Create(f string) (*os.File, error)
	}

	// fsAdapter implements fileCreator interface
	fsAdapter struct{}
)

// Create creates file with a given name and returns file descriptor
func (*fsAdapter) Create(fname string) (*os.File, error) {
	dir := filepath.Dir(fname)

	_, err := os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
	}

	if err != nil {
		return nil, err
	}

	return os.Create(fname)
}

// downloadToFile downloads s3 file to filesystem
func (d *Downloader) downloadToFile(key string, lastModified *time.Time) error {
	fname := generateDownloadFilename(key, d.args.LocalDir, d.args.PrependName, lastModified)

	fd, err := d.fileCreator.Create(fname)
	if err != nil {
		return err
	}

	defer fd.Close()
	infoLog.Printf("Downloading s3://%s/%s to %s...\n", d.args.Bucket, key, fname)

	params := &s3.GetObjectInput{Bucket: &d.args.Bucket, Key: &key}
	_, err = d.downloadManager.Download(fd, params)
	if err != nil {
		return err
	}

	return nil
}

// generateDownloadFilename returns file name of downloaded s3 item
func generateDownloadFilename(key, dir string, prependName bool, lastModified *time.Time) string {
	keyName := filepath.Base(key)

	if prependName {
		keyName = fmt.Sprintf("%s_%s", lastModified.Format(time.RFC3339), keyName)
	}

	return filepath.Join(dir, keyName)
}
