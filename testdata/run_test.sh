#!/bin/sh

rm *.result > /dev/null 2>&1

for f in gi_*.gz; do echo local $f; time gtaxon cli local -t gi_taxid_prot -f $f > $f.local.result; done


killall gtaxon > /dev/null 2>&1
gtaxon server > /dev/null 2>&1 &

for f in gi_*.gz; do echo remote $f; time gtaxon cli remote -t gi_taxid_prot -f $f > $f.remote.result; done

killall gtaxon > /dev/null 2>&1
