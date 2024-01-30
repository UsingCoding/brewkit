# syntax=docker/dockerfile:1.4
FROM docker:20.10

COPY bin/brewkit /usr/local/bin/

ENTRYPOINT ["brewkit"]