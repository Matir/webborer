#!/usr/bin/env bash

set -e
touch coverage.txt

for dir in $(go list ./... | grep -v '/vendor'); do
  go test -race -coverprofile=profile.out -covermode=atomic $dir
  if [ -f profile.out ]; then
    cat profile.out >> coverage.txt
    rm profile.out
  fi
done
