checkcommit(){
  cd /opt/cloudland/
  while :
  do
    git fetch
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
      cd /opt/cloudland/deploy/
      ./allinone.sh
      cd ..
      exec ./autocheck.sh
    else
      echo "Code up to date"
    fi
    sleep 5m
  done
}

checkcommit
