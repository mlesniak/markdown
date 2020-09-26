[![Automatic build, push and deploy](https://github.com/mlesniak/markdown/workflows/Build,%20Push%20and%20Deploy/badge.svg)](https://github.com/mlesniak/markdown/actions?query=workflow%3A%22Build%2C+Push+and+Deploy%22) 
[![Go Report Card](https://goreportcard.com/badge/github.com/mlesniak/markdown)](https://goreportcard.com/report/github.com/mlesniak/markdown)
[![Code of Conduct](https://img.shields.io/badge/%E2%9D%A4-code%20of%20conduct-orange.svg?style=flat)](CODE_OF_CONDUCT.md)


# Overview

This is a custom from-dropbox-to-markdown-to-html-to-in-memory-cache server for my personal site [mlesniak.com](https://mlesniak.com). As stated,  uses dropbox to retrieve files to display so I can edit content in real-time using any tools I like on multiple platforms. 

I try to use a [Zettelkasten](http://localhost:8080/202009010824-Zettelkasten.md)-based system (with tags and backlinks) to structure my thoughts and notes and the personal site is the publicly available part of it, i.e. it shows all pages tagged `#public` which are reachable from the root page. In combination with [The Archive](https://zettelkasten.de/the-archive/) on the desktop and [iaWriter](https://ia.net/de/writer) on mobile this allows for a very pleasent, quick and efficient note-taking and throught-processing process. 

Side remark: I should probably rewrite it to be simpler: Initially, the design was more complicated, dynamic generation of pages and without caching and the directory and package structure still shows it.   

## Usage for anyone besides me

I belive this software can be setup and started by any competent software developer, although quite a bit of configuration and code is tailored to my specific needs and some paths are (still) hard-coded. If you want to deploy it yourself and struggle, simply drop me a [mail](mailto:mail@mlesniak.com) or contact me on [Twitter](https://twitter.com/mlesniak). 

## Internal documentation

The following information is primary useful to me and shows different configuration options and things I don't want to forget. 

## Run on server

The following environment variables have to be set inside the container.

    # cat markdown.env
    TOKEN=<DROPBOX TOKEN>
    SECRET=<DROPBOX APP SECRET>
    LOGS_ENABLED=true

## Start logging daemon

Logging is submitted to [sematext](https://sematext.com) using their logagent. The agent collects all JSON-based output of
docker container which have `LOGS_ENABELED` set to `true`.

    docker run -d --name st-logagent --restart=always \
      -v /var/run/docker.sock:/var/run/docker.sock \
      -e LOGS_TOKEN="<TOKEN>" \
      -e LOGS_ENABLED_DEFAULT=false \
      -e REGION=EU \
      sematext/logagent:latest
      
## Display log on terminal

The following local command shows all json output from the server (but not the webserver logs):

    scripts/logs.sh|jq -r '(select(.message != null) | .time + "\t" + .message)'      
      
A automatic-login ssh config for `server` has to be pre-configured.        
      
## License

The source code is licensed under the [Apache license](LICENSE).      