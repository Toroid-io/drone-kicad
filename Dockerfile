FROM toroid/kicad-base
ADD drone-kicad /bin/
COPY kicad-ci-scripts /bin/ci-scripts
COPY PcbDraw /bin/PcbDraw
ENTRYPOINT /bin/drone-kicad
