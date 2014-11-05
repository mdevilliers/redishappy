Redis Happy
-----------

[![Build Status](https://travis-ci.org/mdevilliers/redishappy.svg?branch=master)](https://travis-ci.org/mdevilliers/redishappy)
[![Coverage Status](https://coveralls.io/repos/mdevilliers/redishappy/badge.png)](https://coveralls.io/r/mdevilliers/redishappy)

One method of providing a highly available Redis service is to deploy using [Redis Sentinel](http://redis.io/topics/sentinel).

Redis Sentinel monitors your Redis cluster and on detecting failure, promotes a slave to become the new master. RedisHappy provides a daemon to monitor for this promotion and to tell the outside world that this has happened.

Currently we support [HAProxy](http://www.haproxy.org/) and [Consul](https://www.consul.io/).

FAQ
---

Q. Why - I thought in 2014 Redis clients should be Sentinel aware? They should connect to the correct Redis instance on failover.

A. Some do, some don't. Some it seems to be an eternal 'work in progress'. Rather than fixing all of the clients we needed to work correctly with Sentinel, RedisHappy was built upon the fact that all of the clients I have tested are great at connecting to a single address. 

Q. Why - This [article](http://blog.haproxy.com/2014/01/02/haproxy-advanced-redis-health-check/) suggests that HAProxy can healthcheck Redis instances quite fine by itself. 

A. Yes. It can do. But not reliably... I'll explain. 

Suppose we have this setup. R1 and R2 are redis instances, S1,S2,S3 are Sentinel instances, H1 and H2 are HAProxy instances. 

<pre>
	R1,R2
	S1, S2, S3
	H1, H2
</pre>

- Life is good - R1 and R2 are in a master slave configuration, H1 and H2 correctly identify R1 as the master

<pre>
	R1      R2
	M  ---- S
    ^
    |
    ---------
    |       |
	H1      H2
</pre>

- Disaster! - R1 dies or is partitioned but don't fear R2 is now the "master". Day saved! 

<pre>
	*       R2
			M
    		^
            |
    ---------
    |       |
	H1      H2
</pre>

- Disaster! - R1 comes back online and announces itself as a "master". Both R1 and R2 are now accepting writes, as HAProxy's healthcheck identifies both as online.

<pre>
	R1		R2
	M       M
    ^		^
    |       |
    ---------
    |       |       
	H1      H2
</pre>

- R1 is made the "slave" of R2. Everything is ok now, except for the writes that R1 accepted which are lost forever.

<pre>
	R1      R2
	S ----- M
    		^
            |
    ---------
    |       |
	H1      H2
</pre>

When a Redis instance is started and stopped it initially announces itself as a "master". It will some time later be made a "slave" but in the meantime accept writes which will be lost when it is correctly made a slave.

RedisHappy attempts to avoid this failure mode by only presenting the correct server to HAProxy or any other service once it is confirmed as a "master". We assume clients will either block or fail until the master is online again.

Defaults
--------

Installs to /opt/redishappy

Configuration to /var/redishappy

Logs to file in /var/redishappy/logs

Warnings, Errors got to syslog

Api
---

RedisHappy provides a readonly api on port 8000

GET /api/pingpong - healthcheck - will reply "pong" if running

GET /api/configuration - displays the start up configuration

GET /api/sentinels - displays the sentinels being currently monitored

GET /api/topology - displays the current view of the Redis clusters, their master and their host/ip addresses


Hacking
-------

Running the following script will gofmt, govet, rune the tests, build all of the executables.

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

Copyright and license
---------------------

Code and documentation copyright 2014 Mark deVilliers. Code released under the Apache 2.0 license.
