FROM docker.elastic.co/beats/filebeat:7.0.0
COPY filebeat.yml /usr/share/filebeat
USER root
WORKDIR  /usr/share/filebeat
RUN chmod go-w /usr/share/filebeat/filebeat.yml
LABEL heritage="dkube"
CMD ["./filebeat","-e"]

