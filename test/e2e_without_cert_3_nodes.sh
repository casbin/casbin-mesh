#!/bin/bash

go build -o ./tmp/casmesh ../cmd/app/*

./tmp/casmesh -node-id node0 -raft-address localhost:4002 -raft-advertise-address localhost:4002 -enable-basic -endpoint-no-verify _1 &
sleep 3
./tmp/casmesh -node-id node1 -raft-address localhost:4004 -raft-advertise-address localhost:4004 -enable-basic -join http://localhost:4002 -endpoint-no-verify _2 &
sleep 3
./tmp/casmesh -node-id node2 -raft-address localhost:4006 -raft-advertise-address localhost:4006 -enable-basic -join http://localhost:4002 -endpoint-no-verify _3 &
wait
