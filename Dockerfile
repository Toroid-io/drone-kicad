FROM toroid/kicad-base:5.0.1
ADD drone-kicad /bin/
COPY kicad-ci-scripts /bin/ci-scripts
COPY PcbDraw /bin/PcbDraw
ENTRYPOINT /bin/drone-kicad
