#!/bin/sh

docker build -t server .
docker tag server 116.203.24.33:5000/markdown:latest
docker push 116.203.24.33:5000/markdown:latest
ssh server docker pull 116.203.24.33:5000/markdown:latest
ssh server docker rm --force markdown
ssh server docker run -d --name markdown --env-file markdown.env -it -p 8088:8080 -v /root/data:/data 116.203.24.33:5000/markdown:latest

#cat <<EOF
#docker pull 116.203.24.33:5000/markdown:latest
#docker rm --force markdown
#docker run -d --name markdown --env-file markdown.env -it -p 8088:8080 -v $(pwd)/data:/data 116.203.24.33:5000/markdown:latest
#EOF