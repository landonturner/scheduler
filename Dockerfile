# build client
FROM node:10 as client
WORKDIR /usr/src/app
COPY client/package*.json ./
RUN npm install
# install deps
COPY client/ ./
RUN npm run build

# build binaries
FROM golang:latest as builder
# install vendor packages first to avoid
# rebuilding them each time there is a code change
COPY ./vendor /go/src/github.com/landonturner/scheduler/vendor
WORKDIR /go/src/github.com/landonturner/scheduler/vendor
RUN go build -v ./...
# only copy in the code we need to avoid unnecessary rebuilds
COPY ./server.go /go/src/github.com/landonturner/scheduler/server.go
COPY ./internal /go/src/github.com/landonturner/scheduler/internal
WORKDIR /go/src/github.com/landonturner/scheduler
RUN go build -v

# build final image
FROM alpine:latest
RUN apk add --update bash ca-certificates sqlite
# See http://stackoverflow.com/questions/34729748/installed-go-binary-not-found-in-path-on-alpine-linux-docker
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

# removing apk cache
RUN rm -rf /var/cache/apk/*

COPY --from=builder /go/src/github.com/landonturner/scheduler/scheduler /opt/
COPY --from=client /usr/src/app/dist /opt/client/dist/
WORKDIR /opt
EXPOSE 1337
CMD ["/opt/scheduler"]
