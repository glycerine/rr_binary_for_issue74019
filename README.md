issue 74019 files
=======================

For https://github.com/golang/go/issues/74019

# rr binary

rr binary and .so dependencies are here https://github.com/glycerine/rr_binary_for_issue74019/raw/refs/heads/master/rr-debuger.tar.gz


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

With the recorded rr trace 
https://github.com/glycerine/rr_binary_for_issue74019/tree/master/traces/rpc045/rpc.test-24

and gemini 2.5 pro (preview)'s help, I tracked this bug
down to a race between the go runtime initialization
for a new thread and the tsan thread sanitizer.

The comment here
https://github.com/golang/go/blob/go1.24.3/src/runtime/mheap.go#L1449

does not seem to account for the hazards of running
under the race detector.

The full debugging session with rr and gemini is
saved here as a single file. It demonstrates
how to use a recorded rr trace along with
hardware watchpoints to debug.

https://github.com/glycerine/rr_binary_for_issue74019/raw/refs/heads/master/issue74019_gemini_root_cause_rr_debug_session.mhtml

You'll probably have to download it and then open
the mhtml file with a browser, as the .mhtml file
is the saved output of the gemini session 
by Chrome's File->Save Page As -> Webpage, single file.
I don't know how to share gemini sessions otherwise.
