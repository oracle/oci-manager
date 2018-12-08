#!/bin/bash -x

files=$(find pkg -name *.go | awk '{print "echo", $1, "\`egrep \"Copyright .* Oracle\" " $1 "\`"}' | sh | egrep -v "Copyright .* Oracle")

for file in $files; do
  cat hack/custom-boilerplate.go.txt  > $file.tmp
  cat $file >> $file.tmp
  mv $file.tmp $file
done
