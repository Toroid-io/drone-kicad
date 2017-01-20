FROM toroid/kicad
ADD drone-kicad /bin/
ADD gen_fp_lib_table.sh /bin/
ENTRYPOINT /bin/drone-kicad
