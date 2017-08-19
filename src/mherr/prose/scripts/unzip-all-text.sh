#!/bin/bash

for f in $(find -name "*.zip"); do
  pushd $(dirname "$f")
  unzip -n $(basename "$f") "*.txt"
  popd
done
