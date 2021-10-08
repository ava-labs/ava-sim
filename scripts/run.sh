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

# Config Dir, VM Location, Genesis Location
config_dir=''
vm_path=''
vm_genesis=''
for flag in "$@"; do
  case "${flag}" in
          --config-dir=*)
            config_dir="${flag#*=}"
                  ;;
          --vm-path=*)
            vm_path="${flag#*=}"
            shift
                   ;;
          --vm-genesis=*)
            vm_genesis="${flag#*=}"
            shift
                   ;;
          *) echo "Invalid option: -$flag" ;;
  esac
done

go run main.go "--config-dir=${config_dir}" "--vm-path=${vm_path}" "--vm-genesis=${vm_genesis}"
