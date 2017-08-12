#!/bin/bash
NODES=$1
echo "$NODES Node Setting"
curl -XPOST localhost:8000/node/scale/$NODES/0

echo "Request Start ...."
FAILNODE=""
for fnode in {4,8,12,16,20,24,28,32}
do
	echo "Testing.... $fnode"
	wrk  -t 8 -c 16 -d 10s http://localhost:8000 -s test.lua
	curl -XGET localhost:8000/count
	curl -XPUT localhost:8000/node/fail/0/$fnode
	curl -XDELETE localhost:8000/count
done
