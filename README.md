# Overview

NOTE: This code is hacked together in one big function, not yet refactored and probably full of bugs. I'll change this after I've implemented most of the features I want...

This is a custom hacked-together markdown-to-html server for [mlesniak.com](https://mlesniak.com). The code is hacky and ugly and needs a lot of refactoring. But it works...

TODO Write documentation with dropbox screenshots, etc.

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
      