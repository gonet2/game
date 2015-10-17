#!/bin/sh
docker rm -f game1
docker build --no-cache --rm=true -t game .
docker run --rm=true --name game1 -h game_dev -it -P -e SERVICE_ID=game1 game
