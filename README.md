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

    docker pull 116.203.24.33:5000/markdown:latest
    docker run -d --name markdown --env-file markdown.env -it -p 8088:8080 -v $(pwd)/data:/data 116.203.24.33:5000/markdown:latest
    
    # Get chatid using
    http https://api.telegram.org/bot<TOKEN>/getUpdates

## Submit content

    ./publish.sh

## Things todo

[ ] Add Sonarqube support
