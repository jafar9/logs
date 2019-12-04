#!/bin/bash
while true ; do
   if [  -f /tmp/shared_dir/gpus ]; then
           while IFS='' read -r line || [[ -n "$line" ]]; do
                etcdctl --endpoints=http://dkube-etcd-server.dkube:2379 --no-sync set "/dkube/exporter/"$line " {\"jobid\":\"$JOBID\","" \"username\":\"$USERNAME\" } " ;
                etcdctl --endpoints=http://dkube-etcd-server.dkube:2379 --no-sync set "/meta/users/{{workflow.parameters.username}}/devices/"$line "" ;
                etcdctl --endpoints=http://dkube-etcd-server.dkube:2379 --no-sync set "/meta/jobs/{{workflow.parameters.jobuuid}}/devices/"$line "" ;
           done < "/tmp/shared_dir/gpus";
            break ;
   fi;
   sleep 0.1 ;
done
