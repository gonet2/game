#!/bin/sh
curl -L http://172.17.42.1:2379/v2/keys/backends/names -XPUT --data-urlencode value@names.txt
curl http://172.17.42.1:2379/v2/keys/seqs/snowflake-uuid -XPUT -d value="0"
curl http://172.17.42.1:2379/v2/keys/seqs/userid -XPUT -d value="0"

