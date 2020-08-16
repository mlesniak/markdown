# Overview

This is a custom hacked-together markdown-to-html server for [mlesniak.com](https://mlesniak.com). The code is hacky and ugly and needs a lot of refactoring.

## Build and run locally

    docker build -t server .
    docker run --name markdown --rm -it -p 8080:8080 -v $(pwd)/data:/data server

## Push to server

    docker tag server 116.203.24.33:5000/markdown:latest
    docker push 116.203.24.33:5000/markdown:latest

## Run on server

    docker pull 116.203.24.33:5000/markdown:latest
    docker run -d --name markdown -it -p 8088:8080 -v $(pwd)/data:/data 116.203.24.33:5000/markdown:latest

## Submit content

    ./publish.sh

## Things todo

[ ] Add Sonarqube support
