#~/bin/bash
Action=$1

BROWN='\033[0;33m'
NC='\033[0m'
RED='\033[0;31m'


help() {
  printf "\n${RED}Usage:${NC} ${BROWN}./run.sh create name=<Name> user=<Username> program=<digits | catsdogs> model_serving_url=<URL> [--image=<container Image>] [--access_ip=<public Ip>]\n\n" 
  printf "OPTIONS\n"
  printf "  --image\n"
  printf "\t Container image to use creating inference\n"
  printf "  --access_ip\n"
  printf "\t To access directly through the web\n\n${NC}"

}

if [ $Action != "create" -a $Action != "delete" ]; then
	echo
	echo -e "\e[91m Invalid Operation (Note: either create or delete)\e[0m"
	echo
	exit 128
fi
for var in "$@"
do
   IFS== read var1 var2 <<<$var
   if [ $var1 == "name" ];then
	   Name=$var2
   elif [ $var1 == "user" ]; then
	   Username=$var2
   elif [ $var1 == "program" ]; then
	 Tag=$var2
   elif [ $var1 == "model_serving_url" ]; then
	Tfurl=$var2
   elif [ $var1 == "image" ]; then
        Image=$var2
   elif [ $var1 == "access_ip" ]; then
        IP=$var2 
   fi	
done

if [ $Action == "delete" ]; then
	if [[ ( -z "$Name" ) || ( -z "$Username") ]]; then
		echo -e "\nUsage: \e[33m./run.sh delete name=<unique name> user=<Username> \e[0m\n"
		exit 128
	fi
	./argo delete $Name -n $Username
	exit 128
fi

if [[ ( -z "$Name" ) ||  ( -z "$Username" ) || ( -z "$Tfurl" ) || ( -z "$Tag" ) ]]; then
	help
	exit 128
fi

if [ -z "$Image" ]; then
	Image="ocdr/dkube-d3inf:alpha3"
fi	
MASTERNODE=`kubectl get nodes -l node-role.kubernetes.io/master="" -o json | jq '.items[] | .status .addresses[] | select(.type=="InternalIP") | .address'`

MASTERNODE=`echo $MASTERNODE | tr -d '["]'`

echo $PORT
echo $MASTERNODE

MODEL=`echo $Tfurl | awk -F/ '{print $NF}'`
./argo submit start-inference-test-wf.yaml -p container=$Image -p model=$MODEL -p tf-serving-url=$Tfurl --name $Name --namespace $Username
PORT=""
while [ ${#PORT} -eq 0 ]; do
     sleep 1
     PORT=`kubectl  get svc $Name -n $Username -o json | jq -j '.spec.ports[0] |.nodePort'`
done

if [ $Tag == "catsdogs" ]; then
	if [ -z "$IP" ];then
        	echo -e "Available @ http://$MASTERNODE:$PORT/catsanddogs"
		echo
		echo -e "\e[93mAccess through web:\e[0m \e[34mhttp://\e[91m<public_ip>\e[0m\e[34m:$PORT/catsanddogs\e[0m "
	else
		echo
		echo -e "\e[93mAvailable\e[0m @ \e[34m\e[4mhttp://$IP:$PORT/catsanddogs\e[0m"
	fi
else
	if [ -z "$IP" ];then
       		echo -e "Available @ http://$MASTERNODE:$PORT/digits"
		echo
		echo -e "\e[93mAccess through web:\e[0m \e[34mhttp://\e[91m<public_ip>\e[0m\e[34m:$PORT/digits\e[0m "
	else
		echo
		echo -e "\e[93mAvailable\e[0m @ \e[34m\e[4mhttp://$IP:$PORT/digits\e[0m"

    	fi
fi
echo
