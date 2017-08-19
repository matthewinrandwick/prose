#!/usr/bin/python

import sys

for l in open('/usr/share/dict/american-english').readlines():
  if l > 'z' or l < 'a':
    continue
  sys.stdout.write('%s\t1\n' % l.strip())
