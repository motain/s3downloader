#!/usr/bin/env bash

USAGE='
	This script downloads file from amazon s3 and upload them to ftp_receiver server.

	USAGE:
		./ftp_uploader.sh "FTP_RECEIVER_IP" "prefix" "competition name"

	EXAMPLE:
		./ftp_uploader.sh "127.0.0.1" "backup/standings/scoreradar/stats/" "Mexico.CopaMX"

'

WATCHING_FOLDER=/sharedstorage/middleware/pub/push/scoreradar/stats
ILIGAHOME=/home/iliga
DOWNLOAD_DIR=./tmp/s3feeds

COMMAND="sudo mkdir -p $ILIGAHOME/tmp_feeds && \
		sudo mv ~/tmp_feeds/*.xml $ILIGAHOME/tmp_feeds && \
		sudo chown iliga:iliga $ILIGAHOME/tmp_feeds/*.xml && \
		sudo mv $ILIGAHOME/tmp_feeds/*.xml $WATCHING_FOLDER && \
		sleep 1 && \
		echo 'Number of files in watching folder after 1 second:' && \
		sudo ls -lA $WATCHING_FOLDER | wc -l && \
		sleep 5 && \
		echo 'Number of files in the watching folder after 5 seconds:' && \
		sudo ls -lA $WATCHING_FOLDER | wc -l
		"

if [[ -z "$1" ]] || [[ -z "$2" ]] || [[ -z "$3" ]]; then
  printf "$USAGE"
else
	FTP_RECEIVER_IP=$1
	echo "1) Downloading files from s3" && \
	rm -f $DOWNLOAD_DIR/*.xml && \
	go run main.go -bucket=motain-feeds -dir=$DOWNLOAD_DIR -prefix=$2 -regexp=$3 && \
	read -p "Do you want to upload these files to the ftp_receiver server? (y/n): " -r
	if [[ "$REPLY" = "y" ]]
	then
		echo "2) Uploading files ssh $FTP_RECEIVER_IP" && \
		rsync -v -r $DOWNLOAD_DIR/*.xml $FTP_RECEIVER_IP:~/tmp_feeds && \
		echo "3) Moving files to $WATCHING_FOLDER"  && \
		ssh $FTP_RECEIVER_IP $COMMAND
	fi
	echo "4) Finish"

fi
