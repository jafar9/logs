#!/bin/bash
while IFS='' read -r line || [[ -n "$line" ]]; do
         etcdctl --endpoints=http://dkube-etcd-server.dkube:2379 --no-sync rm "/dkube/exporter/"$line  ;
         etcdctl --endpoints=http://dkube-etcd-server.dkube:2379 --no-sync rm "/meta/users/{{workflow.parameters.username}}/devices/"$line ;
        etcdctl --endpoints=http://dkube-etcd-server.dkube:2379 --no-sync rm "/meta/jobs/{{workflow.parameters.jobuuid}}/devices/"$line;
done < "/tmp/shared_dir/gpus";
