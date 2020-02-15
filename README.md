# delve-debug-sample

sample for debugging with delve

## how to use

Run main.go with delve debugger.
Make sure build binary with optimizations disabled.

```bash
$ go build -gcflags "-N -l" main.go
# -N    disable optimizations
# -l    disable inlini
$ dlv exec delve-debug-sample

(dlv) continue
2020/02/15 21:03:10 start server

# other terminal
$ curl localhost:12000/foo
```

### set conditional breakpoints

you can define conditional like below

```
(dlv) break main.go:59
(dlv) condition 1 k==10
```

### show goroutines

break at line below

```
log.Println("goroutines dispatched")
```

you can see main.fooHandler.func1, main.fooHandler.func2 goroutines.

### coredump

Unfortunately, `dlv core` is only supported with linux, windows now.

For testing `dlv core` on MacOS, use dockerimage

```bash
# build centos8 container
docker build -t centos8-gdb .

# build for linux
GOARCH=amd64 GOOS=linux go build -gcflags "-N -l" .

# run with security option to use ptrace
docker run --rm --cap-add=SYS_PTRACE --security-opt seccomp=unconfined \
  -it -v ${PWD}:/delve-debug-sample centos8-gdb /bin/bash
# in centos8 container
cd /delve-debug-sample
./delve-debug-sample &
gcore <PID>
exit

# on MacOS
dlv core delve-debug-sample core.<PID>
(dlv) goroutines
```

### profile

```bash
curl "localhost:12001/debug/pprof/trace?seconds=5" > app5s.pprof & \
for i in {1..5}; do curl "localhost:12000/foo"; done

go tool trace app5s.pprof
```

2020/02/15 : trace is not available because of issue: [cmd/trace: requires HTML imports, which doesn't work on any major browser anymore · Issue #34374 · golang/go](https://github.com/golang/go/issues/34374)
