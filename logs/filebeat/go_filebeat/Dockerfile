FROM ubuntu:18.04 
RUN apt-get update && apt-get install -y apt-utils
RUN apt-get install -y curl
RUN curl -L -O https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-6.4.3-amd64.deb
RUN dpkg -i filebeat-6.4.3-amd64.deb
RUN curl -L https://github.com/etcd-io/etcd/releases/download/v3.3.10/etcd-v3.3.10-linux-amd64.tar.gz | tar zxv -C /usr/src
RUN cp /usr/src/etcd-v3.3.10-linux-amd64/etcdctl /usr/local/bin/
COPY filebeat /
COPY insert.sh /
COPY delete.sh /
WORKDIR /
CMD ["./filebeat"]
