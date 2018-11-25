#!/bin/bash

#Get short commit hash from this repo
HASH=$(git rev-parse --short HEAD)
BODY=$(printf '{"tag_name": %s", "target_commitish": "master", "name": "%s", "draft": false, "prerelease": false}' ${HASH} ${HASH})

#Create new release in release repo
curl --data ${BODY} https://api.github.com/repos/GodlikePenguin/agogos-host-release/releases?access_token=$GITHUB_API_TOKEN