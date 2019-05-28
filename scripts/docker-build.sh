#!/usr/bin/env sh

GOPKG=github.com/s-urbaniak/prometheus-adapter

docker run --rm -it \
	--user="$(id -u):$(id -g)" \
	--volume "${PWD}:/go/src/${GOPKG}" \
	-w "/go/src/${GOPKG}" \
	quay.io/surbania/code-generator:latest \
	\
	$@
