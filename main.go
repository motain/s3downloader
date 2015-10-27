package main

import (
	"flag"
	"github.com/motain/s3downloader/cfg"
	"github.com/motain/s3downloader/s3loader"
)

var inArgs = cfg.InArgs{Regexp: ".*"}

func parseFlags() {
	flag.StringVar(&inArgs.Bucket, "bucket", inArgs.Bucket, "Download bucket")
	flag.StringVar(&inArgs.Prefix, "prefix", inArgs.Prefix, "Bucket download path")
	flag.StringVar(&inArgs.LocalDir, "dir", inArgs.LocalDir, "Target local dir")
	flag.StringVar(&inArgs.Regexp, "regexp", inArgs.Regexp, "Item name regular expression")
	flag.BoolVar(&inArgs.DryRun, "dry-run", inArgs.DryRun, "Find only flag - no download")
	flag.Parse()
}

func main() {
	parseFlags()
	if err := s3loader.Download(&inArgs); err != nil {
		panic(err)
	}
}
