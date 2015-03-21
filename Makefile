sha       = $(CI_BUILD_SHA)
number    = $(CI_BUILD_NUMBER)
short_sha = $(shell git log -1 --pretty=format:%h)
version   = 0.0.$(number).git$(short_sha)
host      = ssh docker-bot@vexor-bot-eu0.cloudapp.net

build:
	gom build

run:
	gom build
	./vx-app-websocket

test:
	gom test

gom:
	go get github.com/mattn/gom
	gom install

ci.image:
	git archive $(sha) | $(host) build vx-app-websocket $(version) latest

ci.push:
	$(host) push vx-app-websocket $(version) latest
