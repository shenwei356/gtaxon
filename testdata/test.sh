#!/bin/sh

for f in gi_*; do echo local $f; time gtaxon cli local -t gi_taxid_prot -f gi_small > /dev/null; done

killall gtaxon && gtaxon server &

for f in gi_*; do echo remote $f; time gtaxon cli remote -t gi_taxid_prot -f gi_small > /dev/null; done
