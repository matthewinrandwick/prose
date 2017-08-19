#!/bin/bash

set -o errexit

path=$(dirname $0)
cd "$path/.."
files=$(find sources/OANC-GrAF -name "*.txt")

for n in $(seq 2 5); do
  go run cmd/ngrams-from-text/main.go --filter $n $files > ngrams.$n.txt
done
