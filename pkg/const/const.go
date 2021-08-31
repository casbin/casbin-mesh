/*
@Author: Weny Xu
@Date: 2021/08/31 23:38
*/

package _const

const RBACModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

var SystemRules = [][]string{{"read", "readWrite"}, {"write", "readWrite"}, {"manage", "admin"}, {"readWrite", "admin"}}
