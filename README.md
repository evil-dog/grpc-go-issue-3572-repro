Code to Reproduce grpc-go issue #3572
=================================
BACKGROUND
----------
This is a slightly modified version of the [greeter helloworld example](https://github.com/grpc/grpc-go/tree/master/examples)
from the [grpc-go library](https://github.com/grpc/grpc-go). It has been
modified to reproduce grpc-go issue
[#3572](https://github.com/grpc/grpc-go/issues/3572).

PREREQUISITS
------------
- Docker
- Docker-compose
- code from this repository

Execute It
----------
First we'll run it to highlight the issue.
```
$ cd <path to repo>/examples

$ 01-build-containers.sh
...

$ 02-run-containers.sh
...
```
From the log output you can see that the client was not able to connect to the
server, however it was able to resolve the DNS. In the original issue the
chirpstack-application-server was not able to resolve the DNS either.

In this next step we'll connect the container to the internet. Doing this allows
the client to connect to the server and execute RPCs.
```
$ 03-connect-client-to-internet.sh
...
```
NOTE: If you want to reset the containers execute the `98-stop-containers.sh`
script and then run the `02-run-containers.sh` script.

Run Original Client
-------------------
Now we'll run the original example client in the same environment.
```
$ 98-stop-containers.sh
...

$ 04-run-orig-client.sh
...
```
With this case the client is able to connect to the server without an internet
connection. The primary difference is that the original client is using the
simplified grpc.Dial() method while the modified code is using the
grpc.DialContext() method.

Cleanup Containers
------------------
When you are all done you can execute the following to clean up the containers
from your system.
```
$ 98-stop-containers.sh
...

$ 99-cleanup-containers.sh
...
```
