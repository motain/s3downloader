package main

import (
	"flag"

	"github.com/motain/s3downloader/s3loader"
)

var (
	bucket   string
	prefix   string
	localDir string
	regexp   string = ".*"
	dryRun   bool
)

func parseFlags() {
	flag.StringVar(&bucket, "bucket", bucket, "Download bucket")
	flag.StringVar(&prefix, "prefix", prefix, "Bucket download path")
	flag.StringVar(&localDir, "dir", localDir, "Target local dir")
	flag.StringVar(&regexp, "regexp", regexp, "Item name regular expression")

	flag.BoolVar(&dryRun, "dry-run", dryRun, "Find only flag - no download")
	flag.Parse()
}

func main() {
	parseFlags()
	if err := s3loader.Download(bucket, prefix, localDir, regexp, dryRun); err != nil {
		panic(err)
	}
}
