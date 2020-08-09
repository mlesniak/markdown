#!/bin/sh

docker build -t server .
docker tag server 116.203.24.33:5000/markdown:latest
docker push 116.203.24.33:5000/markdown:latest

cat <<EOF
docker run -d --name markdown -it -p 8088:8080 -v $(pwd)/data:/data 116.203.24.33:5000/markdown:latest
EOF