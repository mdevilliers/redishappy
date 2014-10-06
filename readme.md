Redis Happy
-----------

Automated Redis Failover using HaProxy and Sentinel

[![Build Status](https://travis-ci.org/mdevilliers/redishappy.svg?branch=master)](https://travis-ci.org/mdevilliers/redishappy)
[![Build Status](https://drone.io/github.com/mdevilliers/redishappy/status.png)](https://drone.io/github.com/mdevilliers/redishappy/latest)
[![Coverage Status](https://coveralls.io/repos/mdevilliers/redishappy/badge.png)](https://coveralls.io/r/mdevilliers/redishappy)


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
