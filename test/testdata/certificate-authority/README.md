# Certificate Authority

Generate the TLS certificates for testing:

```shell
echo '{"CN":"CA","key":{"algo":"rsa","size":2048}}' | cfssl gencert -initca - |cfssljson -bare ca -
echo '{"signing":{"default":{"expiry":"87600h","usages":["signing","key encipherment","server auth","client auth"]}}}' > ca-config.json
echo '{"CN":"raft-server","hosts":["localhost","127.0.0.1"],"key":{"algo":"rsa","size":2048}}' | cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json  - | cfssljson -bare raft-server
echo '{"CN":"raft-client","hosts":["localhost","127.0.0.1"],"key":{"algo":"rsa","size":2048}}' | cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json  - | cfssljson -bare raft-client
rm *.csr *.json | true
```
