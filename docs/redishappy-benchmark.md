## RedisHappy Benchmarking

It is to benchmark the RedisHappy performance vis-a-vis Redis performance, to get a better picture of we stand. Since all Redis calls are routed through, HAProxy in case of the RedisHappy, there is an additional hop introduced. The setup for this benchmarking run includes only one sentinel, as we are testing performance rather than HA. In fact we could just have a HAProxy configured, instead of the entire RedisHappy setup, but to keep it _almost_ realistic we have we have the following setup.

    +-PLACEMENT GROUP-----------------------------------+
    |    +---------------+                              |
    |    |  REDIS SLAVE  |                              |
    |    +---------------+                              |
    |           ^                                       |
    |           |                                       |
    |    +----------------+        +---------------+    |
    |    |  REDIS MASTER  | <------|  REDIS-HAPPY  |    |
    |    +----------------+        +---------------+    |
    |           ^                                       |
    |           |                                       |
    |    +------------------+                           |
    |    |  REDIS SENTINEL  |                           |
    |    |  BENCHMARK TOOL  |                           |
    |    +------------------+                           |
    +---------------------------------------------------+
All servers, m4.xlarge running inside the same AWS EC2 Placement Group(for additional network throughput), on the same subnet. They are running Amazon Linux 2016.09 as the base operating system. Following are the package versions.

- Redis : 3.2.4
- Haproxy: 1.6.3
- Go: 1.7.1(Required for the RedisHappy to run)

### Redis Server and Sentinel System Conf

When benchmarking it is important to have a setup as close to the production as possible its essential configuration limits and kernel(or system) level parameters approriately.

#### Setting Limits

Edit the `/etc/security/limits.conf` add the following to lines before the `# End of file`.

    root            hard    nofile     65536
    root            soft    nofile      65536
    *               hard    nofile      65536
    *               soft    nofile      65536

This sets the limit no of opened files, for the specified user.

#### Sysctl Settings

Run the following commands reset the sysctl variables.

    $ sysctl -w net.core.somaxconn=1024
    $ sysctl -w net.core.netdev_max_backlog=3072
    $ sysctl -w net.ipv4.tcp_max_syn_backlog=2048
    $ sysctl -w net.ipv4.tcp_fin_timeout=30
    $ sysctl -w net.ipv4.tcp_keepalive_time=1024
    $ sysctl -w net.ipv4.tcp_max_orphans=131072
    $ sysctl -w net.ipv4.tcp_tw_reuse=1

Additional we have set the following as well specifically on the Redis(and Sentinel) servers.

    $ sysctl -w vm.overcommit_memory=1
    $ echo never > /sys/kernel/mm/transparent_hugepage/enabled

The final is required for the latest versions of Redis, otherwise you will end up getting the following warning while starting the Redis.

    # WARNING you have Transparent Huge Pages (THP) support enabled in your kernel. This will create latency and memory usage issues with Redis. To fix this issue run the command 'echo never > /sys/kernel/mm/transparent_hugepage/enabled' as root, and add it to your /etc/rc.local in order to retain the setting after a reboot. Redis must be restarted after THP is disabled.

### HAProxy System Settings

The limits as specified above are applicable to the HAProxy system as well. But the sysctl setting are slightly different, for efficient network throughput. HAProxy configuration is two-fold, one is the system setting and the other is the HAProxy configuration tweaking itself. The following elaborated on both.

#### System Settings

The sysctl variables that were tweaked are as following.

    sysctl -w net.core.somaxconn=60000
    sysctl -w net.ipv4.tcp_max_orphans=131072
    sysctl -w net.ipv4.tcp_max_syn_backlog=2048
    sysctl -w net.core.netdev_max_backlog=3072
    sysctl -w net.ipv4.tcp_tw_reuse=1
    sysctl -w net.ipv4.tcp_keepalive_time=1200
    sysctl -w net.ipv4.tcp_fin_timeout=30

    sysctl -w net.ipv4.ip_local_port_range="1024 65024"
    sysctl -w net.core.wmem_max=12582912
    sysctl -w net.core.rmem_max=12582912
    sysctl -w net.ipv4.tcp_rmem=20480 174760 25165824
    sysctl -w net.ipv4.tcp_wmem=20480 174760 25165824
    sysctl -w net.ipv4.tcp_mem=25165824 25165824 25165824

    sysctl -w net.ipv4.tcp_window_scaling=1
    sysctl -w net.ipv4.tcp_timestamps=1
    sysctl -w net.ipv4.route.flush=1
    sysctl -w net.ipv4.tcp_slow_start_after_idle=0

#### HAProxy Configuration Settings

The HAProxy configuration used for this benchmark is as given below.

    global
     nbproc 4
     cpu-map 1 0
     cpu-map 2 1
     cpu-map 3 2
     cpu-map 4 3
     log /dev/log local0
     log /dev/log local1 notice
     chroot /var/lib/haproxy
     user haproxy
     group haproxy
     daemon
     maxconn 50000
     tune.bufsize 1638400
     stats socket /tmp/haproxy mode 0600 level admin
     stats timeout 2m

    defaults
     mode tcp
     timeout client 60s
     timeout connect 1s
     timeout server 60s
     option tcpka

    {{range .Clusters}}
    ## start cluster {{ .Name }}
    frontend ft_{{ .Name }}
     bind *:{{ .ExternalPort }}
     bind-process all
     maxconn 50000
     default_backend bk_{{ .Name }}

    backend bk_{{ .Name }}
     server R_{{ .Name }}_1 {{ .Ip }}:{{ .Port }} maxconn 50000
    ## end cluster {{ .Name }}

    {{end}}

All specifications within double braces `{{}}` are variables that are populated from the config.json by RedisHappy. Note, `nbproc` which the number of haproxy processes that need to be started is `4`, and `cpu-map` directive will specify which process maps to which core. Also, the `bind-process` in the `frontend` section ensures is binds to all processes.

### Benchmark Run

We use the `redis-benchmark` tool to benchmark network performance. There a three benchmark runs conducted to each endpoint, Redis Direct and thru RedisHappy. The three benchmark runs included,

- No requests are pipelined
- 10 requests are pipelined
- 100 requests are pipelined

The command used is the `redis-benchmark -h <endpoint> -q -t set,get,incr,lpush,lpop,sadd,spop -c 100 -P <num>`

#### No Pipelining

__RedisHappy__

    SET: 37565.74 requests per second
    GET: 37565.74 requests per second
    INCR: 37579.86 requests per second
    LPUSH: 37565.74 requests per second
    LPOP: 37579.86 requests per second
    SADD: 37579.86 requests per second
    SPOP: 37579.86 requests per second

__Redis Direct__

    SET: 55555.56 requests per second
    GET: 75414.78 requests per second
    INCR: 54704.60 requests per second
    LPUSH: 55279.16 requests per second
    LPOP: 56529.11 requests per second
    SADD: 75414.78 requests per second
    SPOP: 75414.78 requests per second


__Comparison__

| Request Type | RedisDirect Throughput | RedisHappy Throughput | Percentage |
|--------------|-----------------------:|----------------------:|:-----------|
| SET | 55555.56 | 37565.74 | 68% |
| GET | 75414.78  | 37565.74 | 50% |
| INCR | 54704.60 | 37579.86 | 69% |
| LPUSH | 55279.16 | 37565.74 | 68% |
| LPOP | 56529.11 | 37579.86 | 66%|
| SADD | 75414.78 | 37579.86 | 50% |
| SPOP | 75414.78 | 37579.86 | 50% |

#### Pipeline 10 Requests

__RedisHappy__

    SET: 383141.75 requests per second
    GET: 384615.41 requests per second
    INCR: 383141.75 requests per second
    LPUSH: 383141.75 requests per second
    LPOP: 384615.41 requests per second
    SADD: 384615.41 requests per second
    SPOP: 383141.75 requests per second

__Redis Direct__

    SET: 534759.31 requests per second
    GET: 800000.00 requests per second
    INCR: 680272.12 requests per second
    LPUSH: 523560.22 requests per second
    LPOP: 584795.31 requests per second
    SADD: 793650.75 requests per second
    SPOP: 800000.00 requests per second

__Comparison__

| Request Type | RedisDirect Throughput | RedisHappy Throughput | Percentage |
|--------------|-----------------------:|----------------------:|:-----------|
| SET | 534759.31 | 383141.75 | 71% |
| GET | 800000.00  | 384615.41 | 48% |
| INCR | 680272.12 | 383141.75 | 56% |
| LPUSH | 523560.22 | 383141.75 | 73% |
| LPOP | 584795.31 | 384615.41 | 65% |
| SADD | 793650.75 | 384615.41 | 48% |
| SPOP | 800000.00 | 383141.75 | 47% |


#### Pipeline 100 Requests

__RedisHappy__

    SET: 558659.19 requests per second
    GET: 1333333.25 requests per second
    INCR: 704225.31 requests per second
    LPUSH: 526315.81 requests per second
    LPOP: 613496.94 requests per second
    SADD: 1149425.38 requests per second
    SPOP: 1408450.62 requests per second

__Redis Direct__

    SET: 555555.56 requests per second
    GET: 1333333.25 requests per second
    INCR: 709219.88 requests per second
    LPUSH: 520833.34 requests per second
    LPOP: 613496.94 requests per second
    SADD: 1149425.38 requests per second
    SPOP: 1388889.00 requests per second

__Comparison__

| Request Type | RedisDirect Throughput | RedisHappy Throughput | Percentage |
|--------------|-----------------------:|----------------------:|:-----------|
| SET | 555555.56 | 558659.19 | 100% |
| GET | 1333333.25 | 1333333.25 | 100% |
| INCR | 709219.88 | 704225.31 | 99% |
| LPUSH | 520833.34 | 526315.81 | 101% |
| LPOP | 613496.94 | 613496.94 | 100% |
| SADD | 1149425.38 | 1149425.38 | 100% |
| SPOP | 1388889.00 | 1408450.62 | 101% |

### Inference from Benchmark

Though this is very basic benchmark, it does provide a good overview of where HAProxy falls and where it promises much. So the observation points from the this initial study are.

- HAProxy on a single request flood falls short considerably on the thoughput front.
- HAProxy does seem to give consistent performance across all request types on pipeline size of 1
- HAProxy gains in performance as the pipeline size increases
- HAProxy at P=100 gives near redis performance
