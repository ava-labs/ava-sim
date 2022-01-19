#!/bin/bash

MAIN_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )"; cd .. && pwd )

source "$MAIN_PATH"/scripts/constants.sh

# Create genesis
timestamp_genesis_path="${build_dir}/timestampvm/genesis.txt"
vm_id="tGas3T58KzdjLHhBDMnH2TvrddhqTji5iZAMZ3RXs2NLpSnhH"
touch $timestamp_genesis_path
echo "fP1vxkpyLWnH9dD6BQA" > $timestamp_genesis_path

source "$MAIN_PATH"/scripts/run.sh $timestampvm_path $timestamp_genesis_path $vm_id
