#!/bin/bash

if [ "$1" = "" ]; then
	echo "Version is empty"
	exit 0
fi

go build .

docker build -t "akhileshsingh85/appd-mock-amd64:$1" --platform linux/amd64 .
docker build -t "akhileshsingh85/appd-mock:$1" .