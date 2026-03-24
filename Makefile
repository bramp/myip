.PHONY: all analyze upgrade fix test-ci default check-update debug-env check imports fmt vet lint clean veryclean test deps test-ui serve deploy stage version gocloud

default: all

all: format analyze test test-ui

# get the path to this Makefile
TOP := $(dir $(lastword $(MAKEFILE_LIST)))
ROOT = $(realpath $(TOP))

APP_YAML = appengine/app.yaml
GOCLOUD = $(shell command -v gcloud 2> /dev/null)
PLAYWRIGHT = npx playwright

# Detect OS for portable sed -i
ifeq ($(shell uname), Darwin)
	SED_I := sed -i ""
else
	SED_I := sed -i
endif

# Prints out all the GO environment variables. Useful to see the state
# of what is going on with the GOPATH
debug-env:
	printenv | grep 'GO'

# Only update yarn if the package.json has changed.
yarn.lock: package.json
	yarn
	touch $@

deps: static/bower_components

static/bower_components: yarn.lock
	yarn install

check-updates:
	yarn ncu
	go mod tidy
	go get -u all

upgrade:
	go mod tidy
	go get -u ./...
	go mod tidy

analyze: vet lint

check: deps fmt vet lint

format: fmt

fmt:
	go fmt ./...

vet:
	go vet -v ./...

lint:
	staticcheck ./...

test: check
	go test ./...

test-ui: deps
	$(PLAYWRIGHT) test

test-ci: test test-ui

fix:
	go fmt ./...
	go fix ./...

coverage: check
	#goapp test -covermode=count -coverprofile=profile.cov lib/...
	#goveralls -coverprofile=profile.cov -service=travis-ci
	goveralls -service=travis-ci -debug

version: appengine/version.go

# TODO Move version into the app-engine directory
appengine/version.go: $(shell git ls-tree -r HEAD --name-only | grep -v /version.go$) .git/index
	# -ldflags "-X main.BuildTime `date '+%Y-%m-%d %T %Z'` -X main.Version `git describe --long --tags --dirty --always`"
	$(SED_I) "s/\(Version[^\"]*\"\)[^\"]*/\1`git describe --long --tags --dirty --always`/" appengine/version.go
	$(SED_I) "s/\(BuildTime[^\"]*\"\)[^\"]*/\1`date '+%Y-%m-%d %T %Z'`/" appengine/version.go

serve: version deps
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
