Kubernetes Port-Forward
=======================

A Kubernetes controller for automatic local port-forward. Pod initialization and deletion are automatically forwarded and managed by this.

Install
=======

```
$ export PATH=$PATH:$GOPATH/bin
$ make install
```

Usage
=====

Pod setup
---------

Imagine you have PostgreSQL installed via HELM, it's already running when you start up the port forwarder.

```
$ forwarder
I0527 16:39:52.297931   87418 controller.go:64] Starting Pod controller
I0527 16:39:52.409962   87418 port.go:82] Starting pod postgis-postgresql-5cd4b8c798-ggbhf
W0527 16:39:52.409995   87418 port.go:60] Listening postgis-postgresql-5cd4b8c798-ggbhf gorouting start on 5432
```

You can test it:

```
$ psql -h localhost -U postgres
psql (10.3, server 10.1)
Type "help" for help.

postgres=#
```

So far, so good, but now, you need to start up a Redis, and after a helm install --name redis stable/redis, you can see the logs inside the controller:

```
I0527 16:39:59.759055   87418 port.go:82] Starting pod redis-redis-b987dbdf6-x44p5
W0527 16:40:15.143969   87418 port.go:60] Listening redis-redis-b987dbdf6-x44p5 gorouting start on 6379
```

Test the Redis connection with:

```
$ nc localhost 6379
SADD set "hello"
:1
SMEMBERS set
*1
$5
hello
```

Pod teardown
------------

For this scenario we are going to kill the Redis Pod with helm delete --purge redis, the pod is removed from the internal list, closing the goroutine responsible to keep it running and consequently killing the socket opened for listen.

```
W0527 16:50:51.210209   87418 port.go:101] default/redis-redis-b987dbdf6-x44p5 pod deleted.
W0527 16:50:51.210337   87418 port.go:63] So long and thanks for all the fish

$ netstat -an | grep 6379
$
```
