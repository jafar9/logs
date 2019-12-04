#!/bin/bash
dkubepods=$(kubectl get po -n dkube -l 'app in (dkube-ext,dkube-d3api)' |grep -v 'Running\|Error')
for i in "${dkubepods[@]}"
do
    IFS=$'\n' array=($i)
    for j in "${array[@]}"
    do
        echo $j
    done
done

pods=$(kubectl get po  --all-namespaces |grep Error)
for i in "${pods[@]}"
do
    IFS=$'\n' array=($i)
    for j in "${array[@]}"
    do
        echo $i
    done
done
x=1
y=${#pods[@]}
if [ "$y" -gt "$x" ]; then 
    pods=$(kubectl get ns -l heritage=dkube)
    for i in "${pods[@]}"
    do
        IFS=$'\n' array=($i)
        for j in "${array[@]}"
        do
            IFS=$' ' array1=($j)
            if [ "$array1" == "NAME" ]; then
                 echo  "***********************************************************************"
             else
             echo $array1
                kubectl get pods -n $array1 | grep Error | awk 'NR>1 {print $1}' | xargs kubectl delete pod -n $array1
            fi
        done
    done
fi
kubectl get po -n dkube -l 'app in (dkube-ext,dkube-d3api)' | grep -v Running | awk 'NR>1 {print $1}' | xargs kubectl delete pod -n dkube

echo "All pods are in Running state"

