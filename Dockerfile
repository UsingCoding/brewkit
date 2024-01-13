# syntax=docker/dockerfile:1.4
FROM docker:20.10

COPY bin/brewkit /usr/local/bin/

COPY --from=gostore "README.md" "README.md"

ENTRYPOINT ["brewkit"]