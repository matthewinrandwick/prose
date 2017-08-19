#!/bin/bash

set -o errexit

if [[ $1 == "" ]]; then
  echo "Usage: $(basename $0) [oanc-dir]

  Converts the unzipped OANC corpus into ngrams.
  " >&2
  exit 1
fi
corpus="$1"
path=$(dirname $0)
cd "$path/.."
files=$(find "$corpus" -name "*.txt")

for n in $(seq 2 5); do
  go run cmd/ngrams-from-text/main.go --filter $n $files > ngrams.$n.txt
done
