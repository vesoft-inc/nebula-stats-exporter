FROM alpine:3.18.2

ARG TARGETPLATFORM
ARG TARGETARCH

COPY bin/${TARGETPLATFORM}/nebula-stats-exporter /
RUN chmod 0755 /nebula-stats-exporter

# The exporter has no writable runtime state; run it as the protected non-root
# identity.
USER 65532:65532

ENTRYPOINT  [ "/nebula-stats-exporter" ]
