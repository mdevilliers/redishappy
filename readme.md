Redis Happy
-----------

Automated Redis Failover using HaProxy and Sentinel

[![Build Status](https://travis-ci.org/mdevilliers/redishappy.svg?branch=master)](https://travis-ci.org/mdevilliers/redishappy)


```
go test -v ./...
```

```
go install
```

```
redishappy
```


Testing
-------

```
go test -cover -test.coverprofile=redishappy-test-coverage.out

go tool cover -html=redishappy-test-coverage.out

```

PreCheckin
----------

```
gofmt -l -s -w .
```

Testing with Docker
-------------------

https://github.com/mdevilliers/docker-rediscluster

Will start up a master/slave, 3 sentinel redis cluster for testing.

TODO
----

1. Need a circuit breaker if the a sentinel is unreachable