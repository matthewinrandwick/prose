#!/bin/bash

GET https://gist.githubusercontent.com/jamiew/1112488/raw/7ca9b1669e1c24b27c66174762cb04e14cf05aa7/google_twunter_lol | \
  grep : | \
  cut -f1 -d ':' | \
  tr -d '"* ' |
  awk '{print "\\b" $1 "\\b"}'
