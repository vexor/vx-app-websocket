FROM golang:1.3.3

ADD . /app/vx-app-websocket
WORKDIR /app/vx-app-websocket

RUN make gom && make

CMD ["/app/vx-app-websocket/vx-app-websocket"]
EXPOSE 3003
