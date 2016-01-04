package main

import (
	"flag"

	"github.com/motain/s3downloader/cfg"
	"github.com/motain/s3downloader/s3loader"
)

var inArgs = cfg.InArgs{Regexp: ".*", LocalDir: "downloads-s3"}

func main() {
	if err := start(); err != nil {
		panic(err)
	}
}

func parseFlags() {
	flag.StringVar(&inArgs.Bucket, "bucket", inArgs.Bucket, "Download bucket")
	flag.StringVar(&inArgs.Prefix, "prefix", inArgs.Prefix, "Bucket download path")
	flag.StringVar(&inArgs.LocalDir, "dir", inArgs.LocalDir, "Target local dir")
	flag.StringVar(&inArgs.Regexp, "regexp", inArgs.Regexp, "Item name regular expression")
	flag.BoolVar(&inArgs.DryRun, "dry-run", inArgs.DryRun, "Find only flag - no download")
	flag.BoolVar(&inArgs.PrependName, "p", inArgs.PrependName, "Prepend downloaded file name with lastmodified timestamp")
	flag.Parse()
}

func start() error {
	parseFlags()

	if err := inArgs.Validate(); err != nil {
		return err
	}

	conf, err := cfg.GetCfg()
	if err != nil {
		return err
	}

	d, err := s3loader.NewDownloader(&inArgs, conf)
	if err != nil {
		return err
	}

	if err := d.Run(); err != nil {
		return err
	}

	return nil
}
