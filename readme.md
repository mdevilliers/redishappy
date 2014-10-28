Redis Happy
-----------

[![Build Status](https://travis-ci.org/mdevilliers/redishappy.svg?branch=master)](https://travis-ci.org/mdevilliers/redishappy)

[![Coverage Status](https://coveralls.io/repos/mdevilliers/redishappy/badge.png)](https://coveralls.io/r/mdevilliers/redishappy)

One method of providing a highly available Redis service is to deploy using [Redis Sentinel](http://redis.io/topics/sentinel).

Redis Sentinel monitors your Redis cluster and on detecting failure promotes a slave to become the new master. RedisHappy provides a daemon to monitor for this promotion and to tell the outside world that this has happened.

Currently we support [HAProxy](http://www.haproxy.org/) and [Consul](https://www.consul.io/).

Api
---

RedisHappy provides a readonly api at http://localhost:8000

GET /api/pingpong - healthcheck - will reply "pong" if running

GET /api/configuration - displays the start up configurations

GET /api/sentinels - displays the sentinels being currently monitored

GET /api/topology - displays the current view of the Redis clusters, their master and their host/ip addresses


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

 ## Copyright and license

Code and documentation copyright 2014 Mark deVilliers. Code released under the Apache 2.0 license.
Docs released under Creative commons.
