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

image:
	docker build -t quay.io/dmexe/vx-app-websocket .

push:
	docker push quay.io/dmexe/vx-app-websocket
