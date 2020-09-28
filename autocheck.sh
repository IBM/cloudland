checkcommit(){
  cd /opt/cloudland/
  while :
  do
    git fetch
    HEADHASH=$(git rev-parse HEAD)
    UPSTREAMHASH=$(git rev-parse master@{upstream})

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
      sudo git clone https://github.com/IBM/cloudland.git
      sudo mv ./netconf.yml.bak ./cloudland/deploy/netconf.yml
      cd /opt/cloudland/deploy/
      ./allinone.sh
    else
      echo "Code up to date"
    fi
    sleep 1h
  done
}

checkcommit
