#!/usr/bin/env bash
source ./base.sh
#filter='.auth.scope.project.name = \"$username\"| .auth.identity.password.user.name=\"$username\"| .auth.identity.password.user.password=\"$password\" '
body=$(jq  ".auth.scope.project.name = \"$username\"| .auth.identity.password.user.name=\"$username\"| .auth.identity.password.user.password=\"$password\"" ./test_data/token.json)
# command="curl -v  ${host}${token_endpoint} -d '$body' -v 2>&1 | grep X-Subject-Token | cut -d ':' -f2 | xargs"
# echo $command
# token=$(eval $command)
# echo ${token}"  888"
cmd="curl ${host}${token_endpoint} -d '$body'  -i -s | grep X-Subject-Token | cut -d\":\" -f2 | xargs"
token=$(eval $cmd)
# echo $cmd
# echo $token
echo $token > ./temp
token=$(sed s/"\^M"//g <<< `cat -v ./temp`)
