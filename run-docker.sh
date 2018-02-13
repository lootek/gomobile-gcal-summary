#!/bin/bash

docker run \
	--interactive \
	--tty \
	--rm \
	--publish 8000:80 \
	--publish 8008:8333 \
	--security-opt=seccomp:unconfined \
	--cap-add=SYS_PTRACE \
	--volume "$PWD":/go/src/github.com/lootek/gomobile-gcal-summary \
	--workdir /go/src/github.com/lootek/gomobile-gcal-summary \
	gomobile-gcal-summary
