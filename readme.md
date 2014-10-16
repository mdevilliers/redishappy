Redis Happy
-----------

Automated Redis Failover using HaProxy and Sentinel

[![Build Status](https://travis-ci.org/mdevilliers/redishappy.svg?branch=master)](https://travis-ci.org/mdevilliers/redishappy)
[![Build Status](https://drone.io/github.com/mdevilliers/redishappy/status.png)](https://drone.io/github.com/mdevilliers/redishappy/latest)
[![Coverage Status](https://coveralls.io/repos/mdevilliers/redishappy/badge.png)](https://coveralls.io/r/mdevilliers/redishappy)

Api
---
GET /api/pingpong - healthcheck
GET /api/configuration - start up configurations
GET /api/sentinels - sentinels being monitored with cluster information
GET /api/topology - masters of the clusters and host/ip addresses exposed

PreCheckin
----------

```
build/ci_script.sh

```

Testing with Docker
-------------------

https://github.com/mdevilliers/docker-rediscluster

Will start up a master/slave, 3 sentinel redis cluster for testing.

Logging
-------

By default -

Trace - stdout
Info - stdout
Warning - syslog, file, stdout
Error - syslog, file, stdout

The log path is configurable.


Proposed configuration

  "Logging" :{

  	// values are none, all, errorsonly
  	"Level" : "none",
  	// values are none, stdout, file, syslog
  	"Trace" :"none",
	"Info" :"stdout",
	"Warning" : "file",
	"Error" : "syslog,file"
  }