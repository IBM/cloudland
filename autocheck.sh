checkcommit(){
  cd /opt/cloudland/
  while :
  do
    sudo git fetch
    HEADHASH=$(git rev-parse HEAD)
    UPSTREAMHASH=$(git rev-parse @{upstream})
    BRANCHNAME=$(git rev-parse --abbrev-ref HEAD)
    REPOURL=$(git config --get remote.origin.url)

    if [ "$HEADHASH" != "$UPSTREAMHASH" ]
    then
      echo "Deploying new environment"
      sudo systemctl stop hypercube
      sudo systemctl stop cloudland
      sudo systemctl stop cloudlet
      sudo systemctl stop scid
      cd /opt/
      sudo mv ./cloudland/deploy/netconf.yml ./netconf.yml.bak
      sudo rm -rf ./cloudland/
      sudo rm -rf ./libvirt-console-proxy/
      sudo rm -rf ./sci/
      sudo git clone --branch=$BRANCHNAME $REPOURL
      sudo mv ./netconf.yml.bak ./cloudland/deploy/netconf.yml
      sudo cp /home/centos/server.crt ./cloudland/web/clui/public/server.crt
      sudo cp /home/centos/server.key ./cloudland/web/clui/public/server.key
      cd /opt/cloudland/deploy/
      ./allinone.sh
      cd ..
      sudo exec ./autocheck.sh
    else
      echo "Code up to date"
    fi
    sleep 5m
  done
}

checktest(){
  echo "checktest here from $1"
  i=0
  while :
  do
    status=$(curl https://$1/test)
    echo $status
    let i+=1
    if [ "$status" == "DONE" ]
    then
      return 0
    elif [ i -gt 100 ]||[ "$status" == "FAILED" ]
    then
      return 1
    sleep 2
  done
}

if [ ! -n "$1" ]||[ "$1" == "commit" ]
then
  checkcommit
elif [ "$1" == "test" ]
then
  checktest $2
fi
