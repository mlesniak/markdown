FROM golang:alpine
RUN apk add --no-cache git
WORKDIR /markdown
ADD . /markdown
RUN go mod download
RUN go build -o markdown cmd/server/main.go

FROM alpine:latest
WORKDIR /data
COPY --from=0 /markdown/markdown /markdown/server
ENTRYPOINT ["/markdown/server"]