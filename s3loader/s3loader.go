package s3loader

import (
	"fmt"
	"io"
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

// func Download(bucket, prefix, localDir, regPattern string, dryRun bool) error {
func Download(inArgs *cfg.InArgs) error {
	conf, err := cfg.GetCfg()
	if err != nil {
		return err
	}

	creds := credentials.NewStaticCredentials(conf.AWSAccessKeyID, conf.AWSSecretKey, "")
	if _, err := creds.Get(); err != nil {
		return err
	}

	client := s3.New(&aws.Config{Credentials: creds, Region: aws.String(conf.Region)})

	manager := NewS3DownloadManager(client)
	d := NewDownloader(inArgs.Bucket, inArgs.LocalDir, inArgs.Regexp, manager, inArgs.DryRun)

	params := &s3.ListObjectsInput{Bucket: &inArgs.Bucket, Prefix: &inArgs.Prefix}
	if err := client.ListObjectsPages(params, d.eachPage); err != nil {
		return err
	}

	return nil
}

type (
	DownloadManager interface {
		Download(io.WriterAt, *s3.GetObjectInput) (int64, error)
	}

	S3DownloadManager struct {
		*s3manager.Downloader
	}

	Downloader struct {
		dryRun          bool
		downloadManager DownloadManager
		bucket, dir     string
		regexp          *regexp.Regexp
		w               io.Writer
		wg              sync.WaitGroup
		workers         chan int
	}
)

func NewDownloader(
	bucket,
	localDir,
	regPattern string,
	manager *S3DownloadManager,
	dryRun bool,
) *Downloader {
	return &Downloader{
		bucket:          bucket,
		dir:             localDir,
		downloadManager: manager,
		regexp:          regexp.MustCompile(regPattern),
		dryRun:          dryRun,
		workers:         make(chan int, 50),
	}
}

func NewS3DownloadManager(client *s3.S3) *S3DownloadManager {
	downloader := s3manager.NewDownloader(&s3manager.DownloadOptions{
		PartSize:    s3manager.DefaultDownloadPartSize,
		Concurrency: s3manager.DefaultDownloadConcurrency,
		S3:          client,
	})

	return &S3DownloadManager{downloader}
}

func (d *Downloader) eachPage(page *s3.ListObjectsOutput, more bool) bool {
	for _, obj := range page.Contents {
		if !d.regexp.MatchString(*obj.Key) {
			continue
		}

		if d.dryRun {
			d.logInfo(fmt.Sprintf("> Found: s3://%s/%s", d.bucket, *obj.Key))
			continue
		}

		d.workers <- 1
		go func(key string, lastModified *time.Time) {
			d.wg.Add(1)
			if err := d.downloadToFile(key, lastModified); err != nil {
				d.logErr(err)
			}
			<-d.workers
			d.wg.Done()
		}(*obj.Key, obj.LastModified)
	}

	d.wg.Wait()

	return true
}

func (d *Downloader) downloadToFile(key string, lastModified *time.Time) error {
	file := filepath.Join(d.dir, fmt.Sprintf("%s_%s", lastModified.Format(time.RFC3339), filepath.Base(key)))
	if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		return err
	}

	fd, err := os.Create(file)
	if err != nil {
		return err
	}

	defer fd.Close()
	d.logInfo(fmt.Sprintf("> Downloading s3://%s/%s to %s...\n", d.bucket, key, file))

	params := &s3.GetObjectInput{Bucket: &d.bucket, Key: &key}
	_, err = d.downloadManager.Download(fd, params)
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
