#!/bin/bash

#Get short commit hash from this repo
HASH=git rev-parse --short HEAD

#Create new release in release repo
curl --data '{"tag_name": "$HASH",
              "target_commitish": "master",
              "name": "$HASH",
              "draft": false,
              "prerelease": false}' \
    https://api.github.com/repos/GodlikePenguin/agogos-host-release/releases?access_token=$GITHUB_API_TOKEN