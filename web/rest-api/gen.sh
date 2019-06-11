rm -rf rest/*
swagger generate client --with-flatten=full -f swagger.yaml  -A rest -t rest/
rm -rf rest/client
