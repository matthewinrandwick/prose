# prose

Prose is a console-based autopredicting text editor.

Author: Matthew Herrmann, 2017

You might want to use it to write stories, articles or long work documents. It
uses ngrams to complete likely phrases, similar to how autocomplete on phones
works. Since latency is of the utmost importance to make autocomplete
comfortable and useful, I decided to write an editor from scratch rather than
trying to integrate into an existing editor such as vim or emacs.

![example use](demo.gif)

## Dependencies

Prose requires a recent version of Go to compile. It runs on
Linux.

## Compiling

There is no need to place the code under your $GOPATH, the
Makefile manages this automatically.

On Ubuntu/Debian:

   sudo apt install golang-go
   git clone https://github.com/matthewinrandwick/prose.git
   cd prose/src/mherr/prose
   make

The make process also downloads some files from remote sources
which are not included as part of the source code. (I hope to
get permission from those sources to include the files directly
at some point.)

## Use

Start prose with an (existing) filename to edit:

   touch file.txt
   bin/prose file.txt

The prose console window will be shown. By default, the editor
is in autocomplete mode, where it will attempt to autocomplete words using
semicolon (;) and the number keys.

The following commands are supported:

 * Control-S - Saves the current file.
 * Control-D - Exits.
 * Control-A - Enables autocomplete.
 * Control-O - Disables autocomplete.

## Known issues

This is still a work in progress. Many things still do not work as expected, including:

 * Splitting/joining paragraphs is not right yet.
 * Autocomplete sometimes suggests phrases with incomplete word fragments.
 * The OANC corpus is much too small to get good recommendations. It probably needs to be replaced with the British corpus, along with translation between UK and US English spelling.
 * I'd like to set this up to learn ngrams from maildirs.
