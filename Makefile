# Helpful Makefile for dealing with vendoring Go adapted from from:
# http://www.compoundtheory.com/configuring-your-gopath-with-go-and-google-app-engine/
#

.PHONY: install-tools debug-env check imports fmt vet lint clean test deps serve deploy


#Fixes a bug in OSX Make with exporting PATH environment variables
#See: http://stackoverflow.com/questions/11745634/setting-path-variable-from-inside-makefile-does-not-work-for-make-3-81
export SHELL := $(shell echo $$SHELL)

#get the path to this Makefile, its the last in this list
#MAKEFILE_LIST is the list of Makefiles that are executed.
TOP := $(dir $(lastword $(MAKEFILE_LIST)))
ROOT = $(realpath $(TOP))

GOPATH=$(ROOT)
export GOPATH

# Prints out all the GO environment variables. Useful to see the state
# of what is going on with the GOPATH
debug-env:
	printenv | grep 'GO'

install-tools:
	go get github.com/golang/lint/golint
	go get golang.org/x/tools/cmd/goimports
	rm -rf $(ROOT)/vendor/src/golang.org/lint/golint
	rm -rf $(ROOT)/vendor/src/github.com/golang/x/tools/cmd/goimports

check: fmt vet lint test

fmt:
	goapp fmt main lib/...

vet:
	go vet main lib/...

lint:
	golint main
	golint lib/...

test:
	goapp test lib/...
