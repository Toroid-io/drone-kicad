#!/bin/bash

TABLE_DIR=/root/.config/kicad
TABLE=$TABLE_DIR/fp-lib-table

mkdir -p $TABLE_DIR
touch $TABLE

ls /usr/share/kicad/footprints/ -1 | awk -F . 'BEGIN{print "(fp-lib-table"} {print "  (lib (name "$1")(type KiCad)(uri \"$(KISYSMOD)/"$1"."$2"\")(options \"\")(descr \"\"))"}END{print ")"}' > $TABLE
