FROM toroid/kicad-base
ADD drone-kicad /bin/
COPY ci-scripts /bin/ci-scripts
ENTRYPOINT /bin/drone-kicad
