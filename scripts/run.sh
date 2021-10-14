#!/bin/sh

./scripts/prepare-system-plugins.sh

if [ $# -eq 0 ]; then
  go run main.go
elif [ $# -eq 2 ]; then
  go run main.go $1 $2
else
  echo 'invalid number of arguments (expected no args or [vm-path] [vm-genesis]'
  exit 1
fi
