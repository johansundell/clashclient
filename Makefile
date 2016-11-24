GHACCOUNT := johansundell
NAME := clashclient
VERSION := v1.0

include common.mk

deps:
	go get github.com/c4milo/github-release
	go get github.com/mitchellh/gox
	go get github.com/johansundell/cocapi

