#!/bin/bash

go build -o ./tmp/casbind ../cmd/main.go

./tmp/casbind -node-id node0 -raft-addr localhost:4002 -node-no-verify _1 &
sleep 2
./tmp/casbind   -node-id node1 -raft-addr localhost:4004 -join http://localhost:4002 -node-no-verify _2 &
sleep 2
./tmp/casbind   -node-id node2 -raft-addr localhost:4006 -join http://localhost:4002 -node-no-verify _3 &
wait
