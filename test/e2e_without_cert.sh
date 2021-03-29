#!/bin/bash

go build -o ./tmp/casbind ../cmd/main.go

./tmp/casbind -node-id node0 -http-addr localhost:4001 -raft-addr localhost:4002 -node-no-verify _1 &
sleep 2
./tmp/casbind   -node-id node1 -http-addr localhost:4003 -raft-addr localhost:4004 -join http://localhost:4001 -node-no-verify _2 &
sleep 2
./tmp/casbind   -node-id node2 -http-addr localhost:4005 -raft-addr localhost:4006 -join http://localhost:4001 -node-no-verify _3 &
sleep 2
./tmp/casbind   -node-id node3 -http-addr localhost:4007 -raft-addr localhost:4008 -join http://localhost:4001 -node-no-verify _4 &
sleep 2
./tmp/casbind   -node-id node4  -http-addr localhost:4009 -raft-addr localhost:4010 -join http://localhost:4001 -node-no-verify _5 &
sleep 2
wait
