FROM toroid/kicad
ADD drone-kicad /bin/
ENTRYPOINT /bin/drone-kicad
