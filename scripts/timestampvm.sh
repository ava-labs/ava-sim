#!/bin/sh

sh ./scripts/prepare-system-plugins.sh

curr="${PWD}"

# Download timestampvm
timestampvm_version='v1.1.0'
timestampvm_path="${PWD}/timestampvm/timestampvm"
timestamp_genesis_path="${PWD}/timestampvm/genesis.txt"

if [ ! -d "timestampvm" ]
then
  mkdir timestampvm

  # Create genesis
  touch $timestamp_genesis_path
  echo "fP1vxkpyLWnH9dD6BQA" > $timestamp_genesis_path

  # Build
  echo "Building Timestampvm @ ${timestampvm_version} ..."
  go get "github.com/ava-labs/timestampvm@$timestampvm_version";
  tvm_path="$GOPATH/pkg/mod/github.com/ava-labs/timestampvm@$timestampvm_version"
  cd "$tvm_path"
  go build -o "$timestampvm_path" "main/"*.go
  cd "$curr"
  go mod tidy
fi

go run main.go $timestampvm_path $timestamp_genesis_path
