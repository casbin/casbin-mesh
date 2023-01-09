 # Read Consistency
 
The Enforce function in the struct handles enforcing policies in casbin-mesh. It does this by checking the read consistency level passed in the request and determining how to execute the enforcement accordingly. The read consistency levels are QUERY_REQUEST_LEVEL_NONE, QUERY_REQUEST_LEVEL_WEAK, and QUERY_REQUEST_LEVEL_STRONG.

## None

In casbin-mesh, the QUERY_REQUEST_LEVEL_NONE read consistency level indicates that the Enforce request will be handled by the local node without regard to whether the node is the leader or not, and without checking for staleness. This offers the fastest query response, but results may be stale if the node has fallen behind the leader in terms of updates to its underlying database, or if the node is no longer part of the cluster and has stopped receiving updates.

You can use the QUERY_REQUEST_LEVEL_NONE read consistency level by setting the level field in the EnforceRequest message to QUERY_REQUEST_LEVEL_NONE when making the Enforce RPC call.

```bash

result, err := s.Core.Enforce(ctx, request.GetNamespace(), int32(request.Payload.GetLevel()), request.Payload.GetFreshness(), params...)

```
Here, the level field in the EnforceRequest message determines the read consistency level that will be used for the Enforce request. In this case, if level is set to QUERY_REQUEST_LEVEL_NONE, the Enforce request will be handled by the local node without regard to its role (leader or follower) or staleness.

Explicitly selecting this read consistency level can be done by setting the level field in the EnforceRequest message of the gRPC request to QUERY_REQUEST_LEVEL_NONE. This can be seen in the following example:
```bash
request := &command.EnforceRequest{
  Namespace: "my-namespace",
  Payload: &command.EnforcePayload{
    Level: command.EnforcePayload_QUERY_REQUEST_LEVEL_NONE,
    // other fields...
  },
}
response, err := client.Enforce(context.Background(), request)

```
## Weak

The QUERY_REQUEST_LEVEL_WEAK read consistency level is used to ensure that read requests are served by the Leader node in the Raft cluster. This is done to ensure that the returned results are not stale, meaning that they are up-to-date and reflect the current state of the database.

When a read request is made with the QUERY_REQUEST_LEVEL_WEAK read consistency level, the casbin-mesh node serving the request will check if it is the Leader node in the Raft cluster. If it is not the Leader, the request will return an error indicating that it is not the Leader and cannot serve the request. This ensures that the returned results are not stale and reflect the current state of the database.

Explicitly selecting this read consistency level can be done by setting the level field in the EnforceRequest message of the gRPC request to QUERY_REQUEST_LEVEL_WEAK. This can be seen in the following example:
```bash
request := &command.EnforceRequest{
  Namespace: "my-namespace",
  Payload: &command.EnforcePayload{
    Level: command.EnforcePayload_QUERY_REQUEST_LEVEL_WEAK,
    // other fields...
  },
}
response, err := client.Enforce(context.Background(), request)

```

## Strong

The QUERY_REQUEST_LEVEL_STRONG is the highest level of read consistency. When this level is used, the request will be sent to the leader node to execute. The leader node will then replicate the request to the other nodes in the cluster and wait for a quorum of nodes to acknowledge the request before returning the result to the client. This ensures that the result returned to the client reflects the most up-to-date state of the database.

However, using QUERY_REQUEST_LEVEL_STRONG can result in slower performance because the request has to be replicated to multiple nodes and a quorum of nodes must acknowledge the request before it can be completed. This level is useful in scenarios where the most up-to-date result is required and it is acceptable to trade off some performance for stronger consistency.

It is also worth noting that QUERY_REQUEST_LEVEL_STRONG is the default level when the read consistency is not explicitly specified by the client.
