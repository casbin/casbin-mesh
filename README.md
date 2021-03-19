# Casbin-Mesh Draft

> A Copy of [Casbin-Mesh Draft](https://docs.google.com/document/d/1-DrUu24sfRP3wvX1F61t1_oyjouC5Nu-aaFd5hBAfiU/edit#)
## Overview
Casbin-Mesh （以下简称：应用）意在为用户提供一种可以开箱即用的支持多租户的权限控制应用。即用户可以简单的通过 Helm , docker-compose 等工具快速部署一个高性能的可伸缩的 Casbin 应用，并无需关心使用哪一种 Adapter 和持久化存储。
## Feature
本应用将使用 Raft[0] 一致性算法来同步多个节点中的内存数据，并使用嵌入式数据库（Embedded Database）对 Raft Log 进行持久化存储，支持从本地持久化存储的 Log 或 Snapshots 当中恢复服务的内存数据。应用从架构上可分为两部分，Server（即响应用户发起的请求），Client（即用户使用 Client libraries 向服务器发起请求。例如，发起 Enforcer 请求，进行 RBAC Management 操作等）。
## Server
在本草稿中，我们考虑使用 Hashicorp/Raft[1] 的 Raft 实现和使用 bboltDB[2] 实现持久化存储。我们将不再赘述一些 Raft 的基本特性（例如 Tolerating failure, Snapshots (Log compaction) 等），我们将试着着重阐述在应用层的一些特性。
## Transport
在较大请求数的场景下，外部接口使用基于非 JSON 序列化的（及非 HTTP 1.X 的）协议可以获得更好的性能，并可以使用长链接来减少上下文（例如，减少数据库访问权限的检查。将在 Security Section 中具体阐述）的切换。从接口的通用性角度考虑，我们将在外部接口中使用 gRPC ，并同时提供 HTTP API。
## Backup and Restore
备份和恢复主要分为两种类型：Snapshot 和 Log，区别主要在于 Log 可支持有限的回滚（例如，在未来可能支持的事务操作中，通过 Log 备份将一个生产环境中的数据备份到另一个生产环境中，应用仍然可以对现有 Log 进行撤销提交操作来进行数据回滚）。
## Read consistency (Read preference)

让我们想象以下的情况，我们有一个 Casbin 集群，从 follower (secondary) 节点查询数据。当出现下列情况时，我们可能会读取到一些旧的数据（stale read）。
节点依然是集群的一部分，但是在内存数据更新上已落后 Leader。
节点不再是集群的一部分，不再应用（ Apply）来自 Leader 的 Logs。

解决上述问题可以通过：
强制从 Leader 读取数据（可能）可以保证读的一致性。（例如，得到请求时候，从本地 State 查询当前是否为 Leader 节点，若不是返回一个重定向或者返回一个错误。）
（在对于读一致性要求相对不高的情况下）检查当前节点的当前时间于最后收到来自 Leader 心跳包的间隔，若该间隔高于给定阈值则返回一个读取失败（过期）的错误。

让我们再想象一种情况，我们有一个 Casbin 集群，尽管我们已经要求从 Leader 中查询数据，在某些极端情况下，查询可能依然会返回一些旧的数据（stale read）。
Old Leader 的通信被中断，同时剩下的节点中选出了 New Leader ，但 Old Leader 仍然在任期中且收到（并响应）了查询请求（Split-brain）。

![](https://aphyr.com/data/posts/316/etcd-raft-multiprimary.jpg)

在 Image-0[3]  图中 New Leader 已将数值从 1 变更到 4，同时用户向 Old Leader 发起了一个强制 Leader 节点读的请求，依然读到旧数据（stale）。

解决上述问题可以通过：
在 Raft 集群中询问其他（多数的）节点当前节点是否依然是 Leader 节点。

在这里我们引入 rqlite[4] 对读一致性的查询 Level [5] 的定义。

Level 定义如下：
- None: 从本地内存响应查询
- Weak: 从 Leader 读取数据，从本地内存检查是否为 Leader
- Strong: 从 Leader 读取数据，在集群中询问（多数的）节点当前节点是否为 Leader
Read-only （non-voting） nodes
如上文所提到的，权限控制是一个读多写少的模型。我们通过支持添加 Read-only 节点来增加集群的读性能，同时不会增加写入延迟（即订阅来自 Leader 的 Logs 并可以接受一定的写延迟）。Read-only 节点不参与 Raft 共识系统，也不参与投票（其节点数量不计算在投票的总节点数中）。
在 Read-only 节点中读取数据，通过设置查询 Level = None ，并设置数据陈旧情况的阈值。在查询中通过比较给定阈值和当前时间于最后收到来自 Leader 心跳包的时间差，来决定当前查询是否有效（即时间差大于阈值说明数据过于陈旧）。
  
## Security
上文提到的，这是一个支持多租户的应用。这也意味着我们需要对用户的读写请求进行权限验证（当然用户也可以选择关闭请求的权限认证）。同时安全机制分为以下两种：通过 SSL/TLS 通道发送认证信息，亦或是使用 SCRAM[6] 双向认证机制。

### Security State

```json
state
{
  pwd:string,
  roles:map[target:string]role:string,
  mechanisms: "[<SCRAM-SHA-1|SCRAM-SHA-256|NONE>]"
}
```

# Client
## Core
// TODO
## Unit test / E2E test
### Basic
### Advantage
### Related Information
- Raft[0]:  Raft: In Search of an Understandable Consensus Algorithm
- Hashicorp/Raft[1]: https://github.com/hashicorp/raft
- bboltDB[2]: https://github.com/etcd-io/bbolt
- Image-0[3] : Kyle Kingsbury's article on linearizability and stale reads in Raft systems
- rqlite[4] :https://github.com/rqlite/rqlite
- Level [5]: https://github.com/rqlite/rqlite/blob/master/DOC/CONSISTENCY.md
- SCRAM[6]:https://tools.ietf.org/html/rfc5802

