FROM docker.elastic.co/logstash/logstash:6.5.0
COPY logstash.yml /usr/share/logstash/config
COPY pipelines.yml /usr/share/logstash/config
COPY logstash-sample.conf /usr/share/logstash/config
WORKDIR /usr/share/logstash
CMD ["logstash", "-f", "config/logstash-sample.conf"]
