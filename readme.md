Redis Happy
-----------

Automated Redis Failover using HaProxy and Sentinel

[![Build Status](https://travis-ci.org/mdevilliers/redishappy.svg?branch=master)](https://travis-ci.org/mdevilliers/redishappy)


```
go get
```

```
go install
```

```
redishappy
```

To test the json end point -

```
curl -X POST -H "Content-Type: application/json" \
-d '{"method":"HelloService.Say","params":[{"Who":"Test"}], "id":"1"}' \
http://localhost:8085/rpc
```

```
go test ./... -cover
```
