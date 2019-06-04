rm -rf rest/*
swagger generate client --with-flatten=full -f $1  -A rest -t rest/
