#!/bin/bash
echo "4 Node Setting"
curl -XPOST localhost:8000/node/scale/4/2001

echo "Request Start ...."
wrk  -t 1 -c 1 -d 1s http://localhost:8000 -s test.lua

echo "...."
curl -XGET localhost:8000/count
