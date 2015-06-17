sha       = $(CI_BUILD_SHA)
number    = $(CI_BUILD_NUMBER)
short_sha = $(shell git log -1 --pretty=format:%h)
version   = 0.0.$(number).git$(short_sha)
host      = ssh docker-bot@vexor-bot-eu0.cloudapp.net

deps:
	go get github.com/constabulary/gb/...
	gb vendor update -all

build:
	gb build

run:
	gb build
	./bin/server

test:
	gb test

ci.image:
	git archive $(sha) | $(host) build vx-app-websocket $(version) latest

ci.push:
	$(host) push vx-app-websocket $(version) latest
