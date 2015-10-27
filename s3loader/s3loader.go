package s3loader

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/motain/s3downloader/cfg"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	ErrInvalidArgs = errors.New("invalid  arguments")
)

type (
	Downloader struct {
		*s3manager.Downloader
		args         *cfg.InArgs
		regexp       *regexp.Regexp
		pageLister   PageLister
		pageIterator PageIterator
		wg           sync.WaitGroup
		workers      chan int
	}

	PageLister interface {
		ListObjectsPages(params *s3.ListObjectsInput, pageIterator func(*s3.ListObjectsOutput, bool) bool) error
	}

	PageIterator interface {
		Iterate(*s3.ListObjectsOutput, bool) bool
	}

	PageIteratorFunc func(*s3.ListObjectsOutput, bool) bool

	// PageListerFunc func(params *s3.ListObjectsInput, pageIterator PageIteratorFunc) error
)

func (f PageIteratorFunc) Iterate(page *s3.ListObjectsOutput, more bool) bool {
	return f(page, more)
}

func NewDownloader(agrs *cfg.InArgs, conf *cfg.Cfg) (*Downloader, error) {
	if agrs == nil || conf == nil {
		return nil, ErrInvalidArgs
	}
	creds := credentials.NewStaticCredentials(conf.AWSAccessKeyID, conf.AWSSecretKey, "")
	if _, err := creds.Get(); err != nil {
		return nil, err
	}

	client := s3.New(&aws.Config{Credentials: creds, Region: aws.String(conf.Region)})
	manager := NewS3DownloadManager(client)

	d := &Downloader{
		Downloader: manager,
		args:       agrs,
		pageLister: client,
		regexp:     regexp.MustCompile(agrs.Regexp),
		workers:    make(chan int, 50),
	}

	d.pageIterator = d.pickPageIterator()
	return d, nil
}

func (d *Downloader) Run() error {
	params := &s3.ListObjectsInput{Bucket: &d.args.Bucket, Prefix: &d.args.Prefix}
	if err := d.pageLister.ListObjectsPages(params, d.pageIterator.Iterate); err != nil {
		return err
	}
	return nil
}

func NewS3DownloadManager(client *s3.S3) *s3manager.Downloader {
	return s3manager.NewDownloader(&s3manager.DownloadOptions{
		PartSize:    s3manager.DefaultDownloadPartSize,
		Concurrency: s3manager.DefaultDownloadConcurrency,
		S3:          client,
	})
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
	d.logInfo(fmt.Sprintf("> Found: s3://%s/%s", d.args.Bucket, *obj.Key))
}

func (d *Downloader) onItemDownload(obj *s3.Object) {
	if err := d.downloadToFile(*obj.Key, obj.LastModified); err != nil {
		d.logErr(err)
	}
}

func (d *Downloader) downloadToFile(key string, lastModified *time.Time) error {
	file := generateDownloadFilename(key, d.args.LocalDir, d.args.PrependName, lastModified)
	if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		return err
	}

	fd, err := os.Create(file)
	if err != nil {
		return err
	}

	defer fd.Close()
	d.logInfo(fmt.Sprintf("> Downloading s3://%s/%s to %s...\n", d.args.Bucket, key, file))

	params := &s3.GetObjectInput{Bucket: &d.args.Bucket, Key: &key}
	_, err = d.Download(fd, params)

	if err != nil {
		return err
	}

	return nil
}

func (d *Downloader) logInfo(info string) {
	fmt.Println(info)
}

func (d *Downloader) logErr(err error) {
	fmt.Println(err)
}

func generateDownloadFilename(key, dir string, prependName bool, lastModified *time.Time) string {
	keyName := filepath.Base(key)

	if prependName {
		keyName = fmt.Sprintf("%s_%s", lastModified.Format(time.RFC3339), keyName)
	}

	return filepath.Join(dir, keyName)
}
