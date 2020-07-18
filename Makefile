# Helpful Makefile for dealing with vendoring Go adapted from from:
# http://www.compoundtheory.com/configuring-your-gopath-with-go-and-google-app-engine/
#

.PHONY: default install-tools check-update debug-env check imports fmt vet lint clean veryclean test deps serve deploy version

default: test

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

APP_YAML = appengine/app.yaml

NODE_MODULES := $(ROOT)/node_modules/.bin

# Prints out all the GO environment variables. Useful to see the state
# of what is going on with the GOPATH
debug-env:
	printenv | grep 'GO'

# Only update node_modules if the package.json has changed.
node_modules: package.json
	npm install
	touch $@

install-tools: node_modules
	go get -u github.com/golang/lint/golint
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/mattn/goveralls

	# We don't need the source for the tools
	#rm -rf $(ROOT)/vendor/src/golang.org/lint/golint
	#rm -rf $(ROOT)/vendor/src/github.com/golang/x/tools/cmd/goimports

check-updates: install-tools
	$(NODE_MODULES)/ncu -m npm
	cd appengine; $(NODE_MODULES)/ncu -m bower
	# TODO Write goapp get -u script

deps: node_modules
	cd appengine; $(NODE_MODULES)/bower install

	for pkg in github.com/miekg/dns github.com/domainr/whois                           \
	    github.com/gorilla/handlers github.com/gorilla/mux github.com/gorilla/context  \
		github.com/golang/protobuf/proto google.golang.org/appengine/socket            \
		github.com/ua-parser/uap-go/uaparser github.com/kylelemons/godebug/pretty;     \
	do \
		if [ ! -d $(ROOT)/vendor/src/$$pkg ]; then \
		    echo Getting $$pkg; \
			goapp get -u $$pkg; \
		fi \
	done

check: deps fmt vet lint

fmt:
	goapp fmt appengine lib/...

vet:
	# Due to https://github.com/golang/go/issues/17571 go vet doesn't support vendored directories.
	# Disabled until that is fixed (in Go 1.9 most likely)
	# go tool vet -v src

lint:
	golint -set_exit_status ...

test: check
	# Testing standard go
	go test lib/...

	# Testing appengine go
	goapp test lib/...

coverage: check
	#goapp test -covermode=count -coverprofile=profile.cov lib/...
	#goveralls -coverprofile=profile.cov -service=travis-ci
	goveralls -service=travis-ci -debug

version: lib/myip/version.go

lib/myip/version.go: $(shell find . -type f ! -name "version.go")
	# -ldflags "-X main.BuildTime `date '+%Y-%m-%d %T %Z'` -X main.Version `git rev-parse HEAD`"
	sed -i "" "s/\(Version[^\"]*\"\)[^\"]*/\1`git rev-parse HEAD`/" lib/myip/version.go
	sed -i "" "s/\(BuildTime[^\"]*\"\)[^\"]*/\1`date '+%Y-%m-%d %T %Z'`/" lib/myip/version.go

serve:
	goapp serve $(APP_YAML)

deploy: check
	# TODO get the version number from the git-hash
	#@read -p "What is your Project ID?: " projectID; \
	#goapp deploy -application $$projectID $(APP_YAML)
	goapp deploy -application myip-158305 -version v3 $(APP_YAML)

clean:
	rm -rf vendor/src
	rm -rf appengine/static/bower_components

veryclean: clean
	rm -rf node_modules

