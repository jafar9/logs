FROM docker.elastic.co/beats/filebeat:6.5.0
COPY filebeat.yml /usr/share/filebeat
USER root
WORKDIR  /usr/share/filebeat
RUN chmod go-w /usr/share/filebeat/filebeat.yml
CMD ["./filebeat","-e"]
