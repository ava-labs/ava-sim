#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

coreth_version='v0.7.0-rc.14'
evm_path="${PWD}/system-plugins/evm"

if [ ! -d "system-plugins" ]
then
  echo "Building Coreth @ ${coreth_version} ..."
  go get "github.com/ava-labs/coreth@$coreth_version"
  go build -ldflags "-X github.com/ava-labs/coreth/plugin/evm.Version=$coreth_version" -o "$evm_path" "plugin/*.go"
  go mod tidy
fi

# Config Dir, VM Location, Genesis Location
go run main.go "$1" "$2" "$3"
