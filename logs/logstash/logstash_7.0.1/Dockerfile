FROM docker.elastic.co/logstash/logstash:7.0.1
COPY ./logs/logstash/logstash_7.0.1/logstash.yml /usr/share/logstash/config
COPY ./logs/logstash/logstash_7.0.1/pipelines.yml /usr/share/logstash/config
COPY ./logs/logstash/logstash_7.0.1/logstash-sample.conf /usr/share/logstash/config
USER root
LABEL heritage="dkube"
CMD ["logstash -f /config-dir/logstash.conf"]

