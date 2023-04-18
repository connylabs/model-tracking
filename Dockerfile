FROM gcr.io/distroless/static-debian11
COPY bin/linux/amd64/model-tracking /usr/local/bin/
COPY db/migrations /db/migrations
ENTRYPOINT ["/usr/local/bin/model-tracking"]
