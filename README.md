issue 74019 files
=======================

For https://github.com/golang/go/issues/74019

# rr binary

rr binary and .so dependencies (https://github.com/glycerine/rr_binary_for_issue74019/raw/refs/heads/master/rr-debuger.tar.gz), for 


# original trace

The original program and trace are here
too at https://github.com/glycerine/rr_binary_for_issue74019/raw/refs/heads/master/bk.rr.go.crash.unexpected.signal.packed.tar.gz


# async pre-emption or not

comparing runs with and with GODEBUG=asyncpreemptoff=1


The runs.without_preempt.txt logs 12 rr chaos runs with
this command line. Packed traces for each are in traces/rr.no_preempt.
~~~
GODEBUG=asyncpreemptoff=1 rr record -h ./rpc.test -test.v -test.run 016; rr pack
~~~

The runs.preempt.txt file logs 10 rr chaos runs with
this command line. Packed traces for each are in traces/rr.preempt.
~~~
rr record -h ./rpc.test -test.v -test.run 016 ; rr pack
~~~

# analysis

