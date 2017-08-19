.SUFFIXES:

root = $(shell readlink -f ../../..)

all: bin/ngrams.1.all.txt bin/ngrams.1.txt bin/prose bin/ngrams.2.txt bin/bnc-to-text bin/ngrams-from-text

bin/bad.txt:
	mkdir -p bin && scripts/bad-words.sh > bin/bad.txt

bin/ngrams.1.all.txt: bin/bad.txt
	mkdir -p bin && scripts/dict-to-ngrams.py | \
		grep -v -f bin/bad.txt \
	 	> bin/ngrams.1.all.txt

bin/ngrams.1.txt: bin/bad.txt
	mkdir -p bin && scripts/norvig-to-ngrams.sh | \
		grep -v -f bin/bad.txt \
		> bin/ngrams.1.txt

bin/prose: $(wildcard *.go) $(wildcard */*.go) $(wildcard */*/*.go) bin/ngrams.1.txt bin/ngrams.2.txt
	export GOPATH="$(root)" && \
		go test ./... && \
		mkdir -p bin && \
		go build -o bin/prose cmd/prose/main.go

bin/bnc-to-text: $(wildcard *.go) $(wildcard */*.go) $(wildcard */*/*.go)
	export GOPATH="$(root)" && \
		mkdir -p bin && \
		go build -o bin/bnc-to-text cmd/bnc-to-text/bnc-to-text.go

bin/ngrams-from-text: $(wildcard *.go) $(wildcard */*.go) $(wildcard */*/*.go)
	export GOPATH="$(root)" && \
		mkdir -p bin && \
		go build -o bin/ngrams-from-text cmd/ngrams-from-text/main.go

bin/ngrams.2.txt:
	cp -f db-oanc/ngrams.*.txt.gz bin/ && gunzip bin/*.gz

.PHONY: clean

clean:
	rm -f bin/* && rmdir bin
