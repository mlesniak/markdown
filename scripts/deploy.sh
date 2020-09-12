#!/bin/sh

docker build -t server .
docker tag server ghcr.io/mlesniak/markdown:latest
docker push ghcr.io/mlesniak/markdown:latest

ssh server docker pull ghcr.io/mlesniak/markdown:latest
ssh server docker rm --force markdown
ssh server docker run -d --name markdown --env-file markdown.env -v /root/download:/data/download -it -p 8088:8080 ghcr.io/mlesniak/markdown:latest

#cat <<EOF
#docker pull 116.203.24.33:5000/markdown:latest
#docker rm --force markdown
#docker run -d --name markdown --env-file markdown.env -it -p 8088:8080 -v $(pwd)/data:/data 116.203.24.33:5000/markdown:latest
#EOF