checkpr(){
  cd /opt/cloudland/
  sudo git fetch
  BRANCHNAME=$1
  PRSLUG=$2

  echo "Deploying new environment"
  sudo systemctl stop hypercube
  sudo systemctl stop cloudland
  sudo systemctl stop cloudlet
  sudo systemctl stop scid
  cd /opt/
  sudo rm -rf ./cloudland/
  sudo rm -rf ./libvirt-console-proxy/
  sudo rm -rf ./sci/
  sudo git clone --branch=$BRANCHNAME https://github.com/$PRSLUG.git
  sudo echo "PENDING" > ./cloudland/web/clui/public/test_status
  cd /opt/cloudland/deploy/
  ./allinone.sh
  if [ $? -eq 0 ]
  then
    sudo sed -i "s/PENDING/DONE/g" ../web/clui/public/test_status
  else
    sudo sed -i "s/PENDING/FAILED/g" ../web/clui/public/test_status
  fi
}

checktest(){
  echo "checktest here from $1"
  i=0
  while :
  do
    status=$(curl -k https://$1/test_status)
    echo $status
    let i+=1
    if [ "$status" == "DONE" ]
    then
      return 0
    elif [ $i -gt 10 ]||[ "$status" == "FAILED" ]
    then
      return 1
    fi
    sleep 2
  done
}

if [ ! -n "$1" ]||[ "$1" == "pull_request" ]
then
  checkpr $2 $3
elif [ "$1" == "test" ]
then
  checktest $2
fi
