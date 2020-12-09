FROM centos:7.6.1810

COPY nebula-stats-exporter /
RUN chmod +x /nebula-stats-exporter
ENTRYPOINT  [ "/nebula-stats-exporter" ]
