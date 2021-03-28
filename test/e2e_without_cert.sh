#!/bin/bash

export TMP_DATA=`cert_tmp`

go build -o ./tmp/casbind ../cmd/main.go

./tmp/casbind -node-id node0 -http-addr localhost:4001 -raft-addr localhost:4002 -node-no-verify ${TMP_DATA}_1 &
sleep 1
./tmp/casbind   -node-id node1 -http-addr localhost:4003 -raft-addr localhost:4004 -join http://localhost:4001 -node-no-verify ${TMP_DATA}_2 &
sleep 1
./tmp/casbind   -node-id node2 -http-addr localhost:4005 -raft-addr localhost:4006 -join http://localhost:4001 -node-no-verify ${TMP_DATA}_3 &
sleep 1

wait
