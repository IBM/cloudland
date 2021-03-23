checkpr(){
  sudo echo "PENDING" > /opt/test_status
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
  if [ $? -ne 0 ]
  then
    sudo echo "FAILED" > ../web/clui/public/test_status
    sudo echo "FAILED" > /opt/test_status
    exit 1
  fi
  cd /opt/cloudland/tests/
  sudo echo "export endpoint=https://localhost" > testrc
  sudo bash /opt/cloudland/tests/test3.sh
  if [ $? -eq 0 ]
  then
    sudo echo "DONE" > ../web/clui/public/test_status
    sudo echo "DONE" > /opt/test_status
  else
    sudo echo "FAILED" > ../web/clui/public/test_status
    sudo echo "FAILED" > /opt/test_status
    exit 1
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

pend(){
  i=0
  while :
  do
    status=$(ssh -i ~/.ssh/skey cland@$1 'cat /opt/test_status')
    if [ "$status" == "PENDING" ]
    then
      echo $status
    elif [ "$status" == "DONE" ]||[ "$status" == "FAILED" ]
    then
      echo "RUNNING"
      return 0
    elif [ $i -gt 180 ]
    then
      echo "TIMEOUT"
      exit 1
    fi
    let i+=1
    sleep 10
  done
}

if [ ! -n "$1" ]||[ "$1" == "pull_request" ]
then
  checkpr $2 $3
elif [ "$1" == "test" ]
then
  checktest $2
elif [ "$1" == "queue" ]
then
  pend $2
fi
