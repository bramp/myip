#
# This is a Makefile specific to this vendored module.
#

#Include the core Makefile and all its targets
include ../../Makefile

#now let's change the GOPATH to have a `vendor` prefix.
GOPATH := $(GOPATH)/vendor:$(GOPATH)
export GOPATH

#Add our vendored GOPATH bin directory to the path as well
PATH := $(GOPATH)/vendor/bin:$(PATH)
export PATH

check-updates:
	ncu -m npm
	ncu -m bower

# UA profile data
regexes.yaml:
	curl -o "$@" -z "$@" "https://raw.githubusercontent.com/ua-parser/uap-core/master/regexes.yaml"

# Only update node_modules if the package.json has changed.
node_modules: package.json
	npm install
	touch $@

deps: node_modules regexes.yaml
	goapp get github.com/miekg/dns
	goapp get github.com/gorilla/handlers
	goapp get github.com/gorilla/mux
	goapp get github.com/domainr/whois
	goapp get golang.org/x/net/context
	goapp get google.golang.org/appengine/socket
	goapp get github.com/ua-parser/uap-go/uaparser


serve:
	goapp serve app.yaml

deploy: #deps check
	#@read -p "What is your Project ID?: " projectID; \
	#goapp deploy -application $$projectID app.yaml
	goapp deploy -application myip-158305 -version v1 app.yaml

