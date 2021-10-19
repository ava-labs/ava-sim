#!/bin/bash

MAIN_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )"; cd .. && pwd )

source "$MAIN_PATH"/scripts/constants.sh

# Create genesis
timestamp_genesis_path="${build_dir}/timestampvm/genesis.txt"
touch $timestamp_genesis_path
echo "fP1vxkpyLWnH9dD6BQA" > $timestamp_genesis_path

source "$MAIN_PATH"/scripts/run.sh $timestampvm_path $timestamp_genesis_path
