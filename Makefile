.PHONY: default install-tools check-update debug-env check imports fmt vet lint clean veryclean test deps serve deploy stage version gocloud

default: test

#get the path to this Makefile, its the last in this list
#MAKEFILE_LIST is the list of Makefiles that are executed.
TOP := $(dir $(lastword $(MAKEFILE_LIST)))
ROOT = $(realpath $(TOP))

APP_YAML = appengine/app.yaml
NODE_MODULES = $(ROOT)/node_modules/.bin
GOCLOUD = $(shell command -v gcloud 2> /dev/null)


# Prints out all the GO environment variables. Useful to see the state
# of what is going on with the GOPATH
debug-env:
	printenv | grep 'GO'

# Only update node_modules if the package.json has changed.
package-lock.json: package.json
	npm install
	touch $@

node_modules: package-lock.json

# We don't add this target as a dependency, because the `go get` are quite expensive to run.
install-tools: node_modules
	# TODO Do I need these:
	go get -u golang.org/x/lint/golint
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/mattn/goveralls

check-updates: install-tools
	$(NODE_MODULES)/ncu

deps: node_modules
	$(NODE_MODULES)/bower install

check: deps fmt vet lint

fmt:
	go fmt ./...

vet:
	go vet -v ./...

lint:
	golint -set_exit_status ./...

test: check
	go test ./...

coverage: check
	#goapp test -covermode=count -coverprofile=profile.cov lib/...
	#goveralls -coverprofile=profile.cov -service=travis-ci
	goveralls -service=travis-ci -debug

version: appengine/version.go

# TODO Move version into the app-engine directory
appengine/version.go: $(shell git ls-tree -r HEAD --name-only | grep -v /version.go$) .git/index
	# -ldflags "-X main.BuildTime `date '+%Y-%m-%d %T %Z'` -X main.Version `git describe --long --tags --dirty --always`"
	sed -i "" "s/\(Version[^\"]*\"\)[^\"]*/\1`git describe --long --tags --dirty --always`/" appengine/version.go
	sed -i "" "s/\(BuildTime[^\"]*\"\)[^\"]*/\1`date '+%Y-%m-%d %T %Z'`/" appengine/version.go

serve: version
	go run bramp.net/myip/appengine

gcloud:
ifndef GOCLOUD
	$(error "gcloud is not available. Please install the Google Cloud SDK https://cloud.google.com/sdk/docs")
endif

deploy: gcloud version check 
	gcloud app deploy --project myip-158305 --appyaml $(APP_YAML)

# Install but don't promote to the serving version
stage: gcloud version check 
	gcloud app deploy --project myip-158305 --appyaml $(APP_YAML) --no-promote

clean:
	rm -rf static/bower_components

veryclean: clean
	rm -rf node_modules

