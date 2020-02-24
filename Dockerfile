FROM busybox

COPY nebula-exporter /
RUN chmod +x /nebula-exporter
ENTRYPOINT  [ "/nebula-exporter" ]
