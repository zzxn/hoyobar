#!/bin/sh

gofmt -l -w -s . && \
goimports -l -w .

