S3downloader
------------
[![Build Status](https://travis-ci.org/motain/s3downloader.svg?branch=master)](https://travis-ci.org/motain/s3downloader)

Download data from s3 to a local dir

Usage
-----

- Enter your s3 credentials in config file

`cp config.json.dist config.json`

- For available options run `./s3downloader -h`

#### Example 1

```
// download all data from s3://mybucket - data will be stored in downloads-s3 dir next to binary  
s3downloader -bucket=mybucket

// list s3://mybucket contents  
s3downloader -bucket=mybucket -dry-run
```

#### Example 2

```
// download all files from s3://mybucket/docs/backup to /localdir  
s3downloader -bucket=mybucket -prefix=docs/backup -dir=/localdir

// download all files from s3://mybucket/docs/backup to /localdir, prepend filenames with lastmodified timestamp (for repeated filename)  
s3downloader -bucket=mybucket -prefix=docs/backup -dir=/localdir -p
```

#### Example 3

```
// download only files with "txt" extention from s3://mybucket  
s3downloader -bucket=mybucket -regexp=^*\\.txt$
```

License
-------
See the [LICENSE](LICENSE.txt) file for license rights and limitations (MIT).