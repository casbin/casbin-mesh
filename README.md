<p align="center"><img src="./casbin-mesh.png" width="370"></p>
<p align="center">
<b>A scalable authorization application built on Casbin</b>
</p>

# Casbin-Mesh

<p>
  <a href="https://goreportcard.com/report/github.com/casbin/casbin-mesh">
    <img src="https://goreportcard.com/badge/github.com/casbin/casbin-mesh">
  </a>
  <a href="https://godoc.org/github.com/casbin/casbin-mesh">
    <img src="https://godoc.org/github.com/casbin/casbin-mesh?status.svg" alt="GoDoc">
  </a>
    <img src="https://github.com/casbin/casbin-mesh/workflows/Go/badge.svg?branch=master"/>
</p>

Casbin-Mesh is a lightweight, distributed authorization application. Casbin-Mesh uses [Raft](https://raft.github.io) to gain consensus across all the nodes.

# TOC

- [Install](#install)
- [Quick Start](#quick-start)
- [Documentation](#documentation)
- [License](#license)

# Install

## Single Node

### Docker

You can easily start a single Casbin-Mesh node like:

```bash
$ docker pull ghcr.io/casbin/casbin-mesh:latest

$ docker run -it -p 4002:4002 --name=casbin_mesh_single ghcr.io/casbin/casbin-mesh:latest
```

### Binary

```bash
$ casmesh -node-id node0 ~/node1_data
```

## Cluster

- The first benefit of the cluster is that it can be fault-tolerant several nodes crash, which will not affect your business.

- For some special scenarios, you can read from the follower nodes which can increment the throughput of enforcing (reading) operations.

### Docker Compose

docker-compose.yml

```yml
version: "3"
services:
  node0:
    image: ghcr.io/casbin/casbin-mesh:latest
    command: >
      -node-id node0
      -raft-address 0.0.0.0:4002
      -raft-advertise-address node0:4002
      -endpoint-no-verify
    ports:
      - "4002:4002"
    volumes:
      - ./store/casbin/node1:/casmesh/data
  node1:
    image: ghcr.io/casbin/casbin-mesh:latest
    command: >
      -node-id node1
      -raft-address 0.0.0.0:4002
      -raft-advertise-address node1:4002
      -join http://node0:4002
      -endpoint-no-verify
    ports:
      - "4004:4002"
    volumes:
      - ./store/casbin/node2:/casmesh/data
    depends_on:
      - node0
  node2:
    image: ghcr.io/casbin/casbin-mesh:latest
    command: >
      -node-id node2
      -raft-address 0.0.0.0:4002
      -raft-advertise-address node2:4002
      -join http://node0:4002
      -endpoint-no-verify
    ports:
      - "4006:4002"
    volumes:
      - ./store/casbin/node3:/casmesh/data
    depends_on:
      - node0
```

```
$ docker-compose up
```

### Binary

```bash
$ casmesh -node-id -raft-address localhost:4002 -raft-advertise-address localhost:4002 node0 ~/node1_data

$ casmesh -node-id -raft-address localhost:4004 -raft-advertise-address localhost:4004 node1 -join http://localhost:4002  ~/node2_data

$ casmesh -node-id -raft-address localhost:4006 -raft-advertise-address localhost:4006 node2 -join http://localhost:4002  ~/node3_data
```

_Notes: In practice, you should deploy nodes on different machines._

# Quick Start

### Create namespaces

First, We need to create a new namespace, which can be done by performing an HTTP request on the `/create/namespace` on any Casbin-Mesh node.

```bash
$ curl --location --request GET 'http://localhost:4002/create/namespace' \
--header 'Content-Type: application/json' \
--data-raw '{
    "ns": "test"
}'
```

### Set an RBAC model for the test namespace

To setup an Casbin model for a specific namespace, executes following request on `/set/model` endpoint. See all supported [models](https://casbin.org/docs/supported-models).

```bash
$ curl --location --request GET 'http://localhost:4002/set/model' \
--header 'Content-Type: application/json' \
--data-raw '{
    "ns":"test",
    "text":"[request_definition]\nr = sub, obj, act\n\n[policy_definition]\np = sub, obj, act\n\n[role_definition]\ng = _, _\n\n[policy_effect]\ne = some(where (p.eft == allow))\n\n[matchers]\nm = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act"
}'
```

### List all namespaces

Now, let's list the namespaces which we created.

```bash
$ curl --location --request GET 'http://localhost:4002/list/namespaces'
```

The response:

```json
["test"]
```

### Add Policies

Let's add policies for the `test` namespace. See more of [Policies](https://casbin.org/docs/supported-models)

```bash
$ curl --location --request GET 'http://localhost:4002/add/policies' \
--header 'Content-Type: application/json' \
--data-raw '{
    "ns":"test",
    "sec":"p",
    "ptype":"p",
    "rules":[["alice","data1","read"],["bob","data2","write"]]
}'
```

We will receive the sets of effected rules from the response.

```json
{
  "effected_rules": [
    ["alice", "data1", "read"],
    ["bob", "data2", "write"]
  ]
}
```

### First enforce

Now, Let's figure out whether Alice can read data1.

```bash
$ curl --location --request GET 'http://localhost:4002/enforce' \
--header 'Content-Type: application/json' \
--data-raw '{
    "ns":"test",
    "params":["alice","data1","read"]
}'
```

The answer is yes:

```json
{
  "ok": true
}
```

# Documentation

## APIs Overview

### HTTP Endpoints

HTTP endpoints for managing the namespaces and policies in casbin-mesh:

- /create/namespace: to create a new namespace.
- /list/namespaces: to list all existing namespaces.
- /print/model: to print the model for a given namespace.
- /list/policies: to list all policies for a given namespace.
- /set/model: to set the model for a given namespace.
- /add/policies: to add policies to a given namespace.
- /remove/policies: to remove policies from a given namespace.
- /remove/filtered_policies: to remove policies matching a filter from a given namespace.
- /update/policies: to update policies in a given namespace.
- /clear/policy: to clear all policies from a given namespace.
- /enforce: to enforce a policy for a given namespace.
- /stats: to get statistics for a given namespace.

### gRPC Endpoints

gRPC endpoints for managing the namespaces and policies in casbin-mesh:

- Enforce: to enforce a policy for a given namespace.
- CreateNamespace: to create a new namespace.
- SetModelFromString: to set the model for a given namespace.
- AddPolicies: to add policies to a given namespace.
- RemovePolicies: to remove policies from a given namespace.
- RemoveFilteredPolicy: to remove policies matching a filter from a given namespace.
- UpdatePolicies: to update policies in a given namespace.
- ClearPolicy: to clear all policies from a given namespace.
- PrintModel: to print the model for a given namespace.
- ListPolicies: to list all policies for a given namespace.

### gRPC API Reference for the Command Service

The casbin-mesh service has the following methods:

- ShowStats: This method take a StatsRequest message and returns a StatsResponse message. It can be used to retrieve statistical information about the Casbin system.
- ListNamespaces: This method takes a ListNamespacesRequest message and returns a ListNamespacesResponse message. It can be used to list all the namespaces in the       Casbin system.
- PrintModel: This method takes a PrintModelRequest message and returns a PrintModelResponse message. It can be used to retrieve the model of a specific namespace in     the Casbin system.
- ListPolicies: This method takes a ListPoliciesRequest message and returns a ListPoliciesResponse message. It can be used to list all the policies in a specific         namespace in the Casbin system.
- Request: This method takes a Command message and returns a Response message. It can be used to send various commands to the Casbin system.
- Enforce: This method takes an EnforceRequest message and returns an EnforceResponse message. It can be used to check if a request is authorized or not.
- The StatsRequest, ListNamespacesRequest, PrintModelRequest, and ListPoliciesRequest messages are all empty. They only serve as a placeholder for metadata that can be   added to the request.
- The StatsResponse, PrintModelResponse, and ListPoliciesResponse messages all contain a payload field that can be used to pass back additional information.
- The ListNamespacesResponse message contains a list of namespaces in the namespace field.
- The Command message is used to send various commands to the Casbin system. It has a type field that specifies the type of command being sent, a namespace field that   specifies the namespace the command should be applied to, a payload field that contains additional data for the command, and a metadata field that can be used to       pass metadata along with the command.
- The EnforceRequest message is used to request authorization for a specific action. It has a namespace field that specifies the namespace to check the authorization     in, an enforce_payload field that contains the details of the request, and a metadata field that can be used to pass metadata along with the request.
- The EnforceResponse message is used to respond to an EnforceRequest. It has an ok field that specifies if the request is authorized or not.



All documents were located in [docs](https://mesh.casbin.org/docs/guide/getting-started) directory.

# License

This project is licensed under the [Apache 2.0 license](/LICENSE).
