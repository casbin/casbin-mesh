#!/bin/bash

go build -o ./tmp/casbind ../cmd/main.go

export HOST="my.host"
export IP="127.0.0.1"
openssl req -newkey rsa:4096 -nodes -keyout _.key -x509 -days 365 -out _.crt -addext 'subjectAltName = IP:${IP}' -subj '/C=US/ST=CA/L=SanFrancisco/O=MyCompany/OU=RND/CN=${HOST}/'


./tmp/casbind -node-id node0 -http-addr localhost:4001 -raft-addr localhost:4002 -node-cert _cert.pem -node-key _key.pem -node-no-verify -node-encrypt _1 &
sleep 5
./tmp/casbind -node-id node1 -http-addr localhost:4003 -raft-addr localhost:4004 -join http://localhost:4001 -node-cert _cert.pem -node-key _key.pem -node-no-verify -node-encrypt _2 &
sleep 5
./tmp/casbind -node-id node2 -http-addr localhost:4005 -raft-addr localhost:4006 -join http://localhost:4001 -node-cert _cert.pem -node-key _key.pem -node-no-verify -node-encrypt _3 &
sleep 5

wait
