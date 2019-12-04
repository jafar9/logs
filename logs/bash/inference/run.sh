#~/bin/bash
echo
read -p $'Action\e[2m(create or delete)\e[0m:' Action
while [ $Action != "create" -a $Action != "delete" ]; do
	echo -e "\e[31mGive Proper Action either create or delete\e[0m"
	read -p $'Action\e[2m(create or delete)\e[0m:' Action
done
echo
read -p 'Unique Name: ' Name
echo
read -p $'Dkube Username \e[2m(Onboarded username in dkube)\e[0m:' Username
echo
if [ $Action == "create" ] ; then
	read -p $'Tf-Serving-URL \e[2m(This URL is displayed in Dkube UI when model is deployed for serving)\e[0m: ' Tfurl
	echo
	read -p $'Tag\e[2m(mnist or catsdogs)\e[0m:' Tag
	echo
	read -p $'Public Ip\e[2m(optional)\e[0m:' IP
fi

Image="ocdr/dkube-d3inf:alpha3"

MASTERNODE=`kubectl get nodes -l node-role.kubernetes.io/master="" -o json | jq '.items[] | .status .addresses[] | select(.type=="InternalIP") | .address'`

MASTERNODE=`echo $MASTERNODE | tr -d '["]'`


echo $PORT
echo $MASTERNODE


if [ $Action == "create" ]; then 
    MODEL=`echo $Tfurl | awk -F/ '{print $NF}'`
    ./argo submit start-inference-test-wf.yaml -p container=$Image -p model=$MODEL -p tf-serving-url=$Tfurl --name $Name --namespace $Username

    PORT=""
    while [ ${#PORT} -eq 0 ]; do
        sleep 1
        PORT=`kubectl  get svc $Name -n $Username -o json | jq -j '.spec.ports[0] |.nodePort'`
    done

    if [ $Tag == "catsdogs" ]; then
	if [ -z "$IP" ];then	
        	echo -e "\e[31mAvailable\e[0m @ \e[34m\e[4mhttp://$MASTERNODE:$PORT/catsanddogs\e[0m"
	else
		echo -e "\e[31mAvailable\e[0m @ \e[34m\e[4mhttp://$IP:$PORT/catsanddogs\e[0m"
	fi
    else
	if [ -z "$IP" ];then
        	echo -e "\e[31mAvailable\e[0m @ \e[34m\e[4mhttp://$MASTERNODE:$PORT/digits\e[0m"
	else
		echo -e "\e[31mAvailable\e[0m @ \e[34m\e[4mhttp://$IP:$PORT/digits\e[0m"
	fi
    fi
else
    ./argo delete $Name -n $Username
fi

