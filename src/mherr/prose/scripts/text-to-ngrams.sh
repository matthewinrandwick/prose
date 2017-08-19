#!/bin/bash

set -o errexit

if [[ $1 == "" ]]; then
  echo "Usage: $(basename $0) [filename]

  Converts a text file into a database of ngrams.
  " >&2
  exit 1
fi
filename="$1"
path=$(dirname $0)
cd "$path/.."

for n in $(seq 2 5); do
  go run cmd/ngrams-from-text/main.go --filter $n $filename > ngrams.$n.txt
done
