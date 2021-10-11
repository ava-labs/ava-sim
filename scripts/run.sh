#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

coreth_version='v0.7.0-rc.14'
curr="${PWD}"
evm_path="${PWD}/system-plugins/evm"


if [ ! -d "system-plugins" ]
then
  echo "Building Coreth @ ${coreth_version} ..."
  go get "github.com/ava-labs/coreth@$coreth_version";
  coreth_path="$GOPATH/pkg/mod/github.com/ava-labs/coreth@$coreth_version"
  cd "$coreth_path"
  go build -ldflags "-X github.com/ava-labs/coreth/plugin/evm.Version=$coreth_version" -o "$evm_path" "plugin/"*.go
  cd "$curr"
  go mod tidy
fi

if [ $# -eq 0 ]; then
  go run main.go
elif [ $# -eq 2 ]; then
  go run main.go $1 $2
else
  echo 'invalid number of arguments (expected no args or [vm-path] [vm-genesis]'
  exit 1
fi
