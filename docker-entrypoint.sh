#!/bin/bash

# User wants to override options, so merge with defaults.
if [ "${1:0:1}" = '-' ]; then
        echo $@
        set -- /root/casbin_mesh -raft-address 0.0.0.0:4002 $@ /casbin_mesh/data/data
fi

exec "$@"