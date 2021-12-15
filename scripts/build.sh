#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

MAIN_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )"; cd .. && pwd )

source "$MAIN_PATH"/scripts/constants.sh

# Download coreth
echo "Building Coreth @ ${coreth_version} ..."
go get "github.com/ava-labs/coreth@$coreth_version";
coreth_path="$GOPATH/pkg/mod/github.com/ava-labs/coreth@$coreth_version"
cd "$coreth_path"
go build -ldflags "-X github.com/ava-labs/coreth/plugin/evm.Version=$coreth_version" -o "$evm_path" "plugin/"*.go
cd "$MAIN_PATH"

# Build subnet-evm
echo "Building Subnet-EVM @ ${subnetevm_version} ..."
go get "github.com/ava-labs/subnet-evm@$subnetevm_version";
svm_path="$GOPATH/pkg/mod/github.com/ava-labs/subnet-evm@$subnetevm_version"
cd "$svm_path"
go build -ldflags "-X github.com/ava-labs/subnet-evm/plugin/evm.Version=$subnetevm_version" -o "$subnetevm_path" "plugin/"*.go
cd "$MAIN_PATH"

# Build timestampvm
echo "Building Timestampvm @ ${timestampvm_version} ..."
go get "github.com/ava-labs/timestampvm@$timestampvm_version";
tvm_path="$GOPATH/pkg/mod/github.com/ava-labs/timestampvm@$timestampvm_version"
cd "$tvm_path"
go build -o "$timestampvm_path" "main/"*.go
cd "$MAIN_PATH"

# Building coreth + using go get can mess with the go.mod file.
go mod tidy

# Exit build successfully if the binaries are created
if [[ -f "$evm_path" && -f "$subnetevm_path" && -f "$timestampvm_path" ]]; then
        echo "Build Successful"
        exit 0
else
        echo "Build failure" >&2
        exit 1
fi
