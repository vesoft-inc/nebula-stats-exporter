FROM alpine:3.18.2

ARG TARGETPLATFORM
ARG TARGETARCH

COPY bin/${TARGETPLATFORM}/nebula-stats-exporter /
ENTRYPOINT  [ "/nebula-stats-exporter" ]
