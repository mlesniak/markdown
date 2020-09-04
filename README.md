[![Automatic build, push and deploy](https://github.com/mlesniak/markdown/workflows/Build,%20Push%20and%20Deploy/badge.svg)](https://github.com/mlesniak/markdown/actions?query=workflow%3A%22Build%2C+Push+and+Deploy%22) 
[![Go Report Card](https://goreportcard.com/badge/github.com/mlesniak/markdown)](https://goreportcard.com/report/github.com/mlesniak/markdown)

# Overview

This is a custom hacked-together markdown-to-html server for [mlesniak.com](https://mlesniak.com) which uses dropbox to retrieve files to display, hence we can edit content in realtime using any tools we like. See [here](https://mlesniak.com/202009011431%20Missing%20Features) for implemented and planned features. 

## Build and run locally

    docker build -t server .
    docker run --name markdown --rm -it -p 8080:8080 -v $(pwd)/data:/data server

## Push to server

    docker tag server 116.203.24.33:5000/markdown:latest
    docker push 116.203.24.33:5000/markdown:latest

## Run on server

    # cat markdown.env
    TOKEN=<DROPBOX TOKEN>
    LOGS_ENABLED=true

    docker pull 116.203.24.33:5000/markdown:latest
    docker run -d --name markdown --env-file markdown.env -it -p 8088:8080 -v $(pwd)/data:/data 116.203.24.33:5000/markdown:latest
    
## Start logging daemon

    docker run -d --name st-logagent --restart=always \
      -v /var/run/docker.sock:/var/run/docker.sock \
      -e LOGS_TOKEN="<TOKEN>" \
      -e LOGS_ENABLED_DEFAULT=false \
      -e REGION=EU \
      sematext/logagent:latest
      