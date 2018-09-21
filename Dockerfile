ARG BASE_TAG=latest
FROM toroid/kicad-base:${BASE_TAG}
ADD drone-kicad /bin/
COPY kicad-ci-scripts /bin/ci-scripts
COPY PcbDraw /bin/PcbDraw
ENTRYPOINT /bin/drone-kicad
