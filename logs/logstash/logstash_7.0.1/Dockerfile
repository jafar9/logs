FROM docker.elastic.co/logstash/logstash:7.0.1
COPY logstash.yml /usr/share/logstash/config
COPY pipelines.yml /usr/share/logstash/config
COPY logstash-sample.conf /usr/share/logstash/config
USER root
LABEL heritage="dkube"
CMD ["logstash -f /config-dir/logstash.conf"]

