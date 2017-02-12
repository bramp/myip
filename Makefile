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

GOPATH := $(ROOT)/vendor:$(ROOT)
export GOPATH

#Add our vendored GOPATH bin directory to the path as well
PATH := $(ROOT)/vendor/bin:$(PATH)
export PATH

APP_YAML = src/main/app.yaml

# Prints out all the GO environment variables. Useful to see the state
# of what is going on with the GOPATH
debug-env:
	printenv | grep 'GO'

install-tools: node_modules
	go get github.com/golang/lint/golint
	go get golang.org/x/tools/cmd/goimports
	rm -rf $(ROOT)/vendor/src/golang.org/lint/golint
	rm -rf $(ROOT)/vendor/src/github.com/golang/x/tools/cmd/goimports

check-updates:
	ncu -m npm
	cd src/main; ncu -m bower
	# TODO Write goapp get -u script

# UA profile data
src/main/regexes.yaml:
	curl -o "$@" -z "$@" "https://raw.githubusercontent.com/ua-parser/uap-core/master/regexes.yaml"

# Only update node_modules if the package.json has changed.
node_modules: package.json
	npm install
	touch $@

deps: node_modules src/main/regexes.yaml
	goapp get github.com/miekg/dns
	goapp get github.com/gorilla/handlers
	goapp get github.com/gorilla/mux
	goapp get github.com/domainr/whois
	goapp get golang.org/x/net/context
	goapp get google.golang.org/appengine/socket
	goapp get github.com/ua-parser/uap-go/uaparser

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

serve:
	goapp serve $(APP_YAML)

deploy: #deps check
	#@read -p "What is your Project ID?: " projectID; \
	#goapp deploy -application $$projectID $(APP_YAML)
	goapp deploy -application myip-158305 -version v1 $(APP_YAML)

