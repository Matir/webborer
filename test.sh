#!/usr/bin/env bash

set -e
mode="atomic"
echo "mode: ${mode}" > coverage.txt

for dir in $(go list ./... | grep -v '/vendor'); do
  go test -race -coverprofile=profile.out -covermode=${mode} $dir
  if [ -f profile.out ]; then
    grep -v '^mode: ' profile.out >> coverage.txt
    rm profile.out
  fi
done

if [ -t 1 ] ; then
  go tool cover -html=coverage.txt
fi
