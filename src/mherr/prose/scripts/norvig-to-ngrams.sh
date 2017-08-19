#!/bin/bash

GET http://norvig.com/ngrams/count_1w.txt | head -10000 | sort
