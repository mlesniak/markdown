FROM golang:alpine
RUN apk add --no-cache git
WORKDIR /markdown
ADD . /markdown
RUN go get .
RUN go build .

FROM alpine:latest
WORKDIR /data
COPY --from=0 /markdown/markdown /markdown/server
ENTRYPOINT ["/markdown/server"]