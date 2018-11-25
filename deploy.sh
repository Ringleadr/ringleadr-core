#!/bin/bash

HASH=$(git rev-parse --short HEAD)

go get github.com/aktau/github-release

github-release release \
  --user GodlikePenguin \
  --repo agogos-host-release \
  --tag ${HASH} \
  --name ${HASH} \
  --description "Release for tag ${HASH}"

github-release upload \
  --user GodlikePenguin \
  --repo agogos-host-release \
  --tag ${HASH} \
  --name "agogos-host-darwin" \
  --file build/agogos-host-darwin

github-release upload \
  --user GodlikePenguin \
  --repo agogos-host-release \
  --tag ${HASH} \
  --name "agogos-host-linux" \
  --file build/agogos-host-linux