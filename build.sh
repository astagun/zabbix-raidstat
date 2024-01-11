#!/bin/sh
# golang stretch links against older glibc
name=`basename $(pwd)`
docker run --rm -ti -v .:/$name golang:stretch bash -c "cd /${name}; go build -ldflags '-s -w' && chown `id -u`:`id -g` $name"
