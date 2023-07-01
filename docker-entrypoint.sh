#!/bin/bash

# User wants to override options, so merge with defaults.
if [ "${1:0:1}" = '-' ]; then
        echo $@
        set -- /root/casmesh --raft-address 0.0.0.0:4002 $@ /casmesh/data/data
fi

exec "$@"