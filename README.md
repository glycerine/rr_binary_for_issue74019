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

and gemini 2.5 pro (preview)'s help, I tracked 
one specifc instance of a runtime race bug
down to a race between the go runtime initialization
for a new thread and the tsan thread sanitizer.

So really, there are probably several race condition bugs
in the runtime that rr -h chaos mode will flush out.

Fortunately, each one gets recorded in its own trace.
Thus each can be tackled in turn.

I've made a bunch a progress on the rpc.test-24 trace,
and gotten pretty close to a conclusion, at least a
tentative one. 

The recording was of this test, under the race detector:

~~~
$ cd rpc25519 # at tag v1.22.62 https://github.com/glycerine/rpc25519
$ go test -race -c -o rpc.test
$ rr record -h ./rpc.test -test.v -test.run Test04[05] ## -h activates chaos mode: harsh thread scheduling aimed at finding concurrency bugs
~~~

which gives in some small fraction of executions (say
10% of the time), a crash quite similar to the following.

# To cut a long story short

The root cause seems to be that the comment here
https://github.com/golang/go/blob/go1.24.3/src/runtime/mheap.go#L1449

does not seem to account for the hazards of running
under the race detector (tsan), and we witness a race between
the thread startup process and a tsan thread.

The full debugging session with rr and gemini is
saved here as a single file. It demonstrates
how to use a recorded rr trace along with
hardware watchpoints to debug.

https://github.com/glycerine/rr_binary_for_issue74019/raw/refs/heads/master/issue74019_gemini_root_cause_rr_debug_session.mhtml

(Ignore my first two questions to gemini, the fun really
starts at the third prompt).

To view, you'll probably have to download it and then open
the mhtml file with a browser, as the .mhtml file
is the saved output of the gemini session 
by Chrome's File->Save Page As -> Webpage, single file.

(I don't know how to share gemini sessions otherwise;
if there are tricks, let me know).

Now back to the story...

# initial segfault stack traces

Here is the stack trace of the segfault:

~~~
-*- mode: compilation; default-directory: "~/rpc25519/" -*-
Compilation started at Sun Jun  8 18:10:18

rr record -h ./rpc.test -test.v -test.run Test04[05]
rr: Saving execution to trace directory `/mnt/oldrog/home/jaten/go/src/github.com/glycerine/rr_binary_for_issue74019/rr/rpc.test-24'.
faketime = false
=== RUN   Test040_remote_cancel_by_context

cli_test.go:948 [goID 10] 2025-06-08 23:10:18.343620423 +0000 UTC server Start() returned serverAddr = '127.0.0.1:34869'

cli.go:211 [goID 12] 2025-06-08 23:10:18.366232721 +0000 UTC connected to server 127.0.0.1:34869

cli_test.go:982 [goID 10] 2025-06-08 23:10:18.366636546 +0000 UTC cli_test 040: about to block on test040callStarted

cli_test.go:974 [goID 18] 2025-06-08 23:10:18.367014513 +0000 UTC client.Call() goro top about to call over net/rpc: MustBeCancelled.WillHangUntilCancel()
example.go:300 server-side: WillHangUntilCancel() is running
example.go:304 net.rpc API: our net.Conn has 
local = '127.0.0.1:34869'
remote = '127.0.0.1:51946'

cli_test.go:984 [goID 10] 2025-06-08 23:10:21.227595976 +0000 UTC cli_test 040: we got past test040callStarted

cli_test.go:988 [goID 10] 2025-06-08 23:10:21.227787850 +0000 UTC past cancelFunc()

cli_test.go:977 [goID 18] 2025-06-08 23:10:21.231561912 +0000 UTC client.Call() returned with cliErr = 'cancellation request sent'

cli_test.go:991 [goID 10] 2025-06-08 23:10:21.231755899 +0000 UTC past cliErrIsSet channel; cliErr40 = 'cancellation request sent'

cli_test.go:998 [goID 10] 2025-06-08 23:10:21.231934678 +0000 UTC about to verify that server side context was cancelled.
example.go:310 MustBeCancelled.WillHangUntilCancel(): ctx.Done() was closed!

cli_test.go:1000 [goID 10] 2025-06-08 23:10:21.237329200 +0000 UTC server side saw the cancellation request: confirmed.
example.go:324 server-side: MessageAPI_HangUntilCancel() is running
example.go:326: Message API: our net.Conn has local = '127.0.0.1:34869'; remote = '127.0.0.1:51946'

cli_test.go:1020 [goID 10] 2025-06-08 23:10:21.243213831 +0000 UTC cli_test 041: we got past test041callStarted

cli_test.go:1024 [goID 10] 2025-06-08 23:10:21.243403981 +0000 UTC past cancelFunc()

cli_test.go:1014 [goID 24] 2025-06-08 23:10:21.243850237 +0000 UTC client.Call() returned with cliErr = 'cancellation request sent'

cli_test.go:1027 [goID 10] 2025-06-08 23:10:21.244036540 +0000 UTC past cliErrIsSet channel; cliErr = 'cancellation request sent'

cli_test.go:1038 [goID 10] 2025-06-08 23:10:21.244213305 +0000 UTC about to verify that server side context was cancelled.
example.go: MustBeCancelled.MessageAPI_HangUntilCancel(): ctx.Done() was closed!

0 assertions thus far

--- PASS: Test040_remote_cancel_by_context (2.91s)
=== RUN   Test045_upload

cli.go:211 [goID 29] 2025-06-08 23:10:21.313925974 +0000 UTC connected to server 127.0.0.1:38633
SIGSEGV: segmentation violation
PC=0x45e574 m=3 sigcode=1 addr=0x61662

goroutine 0 gp=0xc200003180 m=3 mp=0xc20006b008 [idle]:
runtime.markrootSpans(0xc200055750, 0x0?)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mgcmark.go:370 +0x134 fp=0x7ba951eb3ce0 sp=0x7ba951eb3c58 pc=0x45e574
runtime.markroot(0xc200055750, 0x74, 0x1)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mgcmark.go:193 +0xef fp=0x7ba951eb3d88 sp=0x7ba951eb3ce0 pc=0x45de0f
runtime.gcDrain(0xc200055750, 0xb)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mgcmark.go:1186 +0x3d4 fp=0x7ba951eb3df0 sp=0x7ba951eb3d88 pc=0x460234
runtime.gcDrainMarkWorkerFractional(...)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mgcmark.go:1116
runtime.gcBgMarkWorker.func2()
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mgc.go:1517 +0x7c fp=0x7ba951eb3e40 sp=0x7ba951eb3df0 pc=0x45c51c
runtime.systemstack(0x98f000)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:514 +0x4a fp=0x7ba951eb3e50 sp=0x7ba951eb3e40 pc=0x4bb14a

goroutine 19 gp=0xc2003ac8c0 m=3 mp=0xc20006b008 [GC worker (active)]:
runtime.systemstack_switch()
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:479 +0x8 fp=0xc2003d2f38 sp=0xc2003d2f28 pc=0x4bb0e8
runtime.gcBgMarkWorker(0xc20003cee0)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mgc.go:1483 +0x1e7 fp=0xc2003d2fc8 sp=0xc2003d2f38 pc=0x45c1e7
runtime.gcBgMarkStartWorkers.gowrap1()
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mgc.go:1339 +0x25 fp=0xc2003d2fe0 sp=0xc2003d2fc8 pc=0x45bfc5
runtime.goexit({})
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc2003d2fe8 sp=0xc2003d2fe0 pc=0x4bd101
created by runtime.gcBgMarkStartWorkers in goroutine 12
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mgc.go:1339 +0x105

goroutine 1 gp=0xc200002380 m=nil [chan receive]:
runtime.gopark(0x18?, 0x1989a00?, 0x80?, 0x4?, 0xc200361598?)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:435 +0xce fp=0xc200361530 sp=0xc200361510 pc=0x4b4d6e
runtime.chanrecv(0xc20003d2d0, 0xc200361616, 0x1)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/chan.go:664 +0x3e5 fp=0xc2003615a8 sp=0xc200361530 pc=0x44aa45
runtime.chanrecv1(0x1987800?, 0x1122520?)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/chan.go:506 +0x12 fp=0xc2003615d0 sp=0xc2003615a8 pc=0x44a632
testing.(*T).Run(0xc2002ff180, {0x123680f, 0xe}, 0x1273f98)
	/mnt/oldrog/usr/local/go1.24.3/src/testing/testing.go:1859 +0x91e fp=0xc200361708 sp=0xc2003615d0 pc=0x60e37e
testing.runTests.func1(0xc2002ff180)
	/mnt/oldrog/usr/local/go1.24.3/src/testing/testing.go:2279 +0x86 fp=0xc200361758 sp=0xc200361708 pc=0x612886
testing.tRunner(0xc2002ff180, 0xc20011f960)
	/mnt/oldrog/usr/local/go1.24.3/src/testing/testing.go:1792 +0x226 fp=0xc200361828 sp=0xc200361758 pc=0x60c8c6
testing.runTests(0xc200330b28, {0x19825e0, 0x71, 0x71}, {0xffffffffffffffff?, 0x0?, 0x0?})
	/mnt/oldrog/usr/local/go1.24.3/src/testing/testing.go:2277 +0x96d fp=0xc200361990 sp=0xc200361828 pc=0x6126ed
testing.(*M).Run(0xc2000d6820)
	/mnt/oldrog/usr/local/go1.24.3/src/testing/testing.go:2142 +0xeeb fp=0xc200361d08 sp=0xc200361990 pc=0x60fccb
github.com/glycerine/rpc25519.TestMain.func1(0xc200361e50, 0xc2000d6820)
	/home/jaten/rpc25519/cli_test.go:71 +0x21e fp=0xc200361e40 sp=0xc200361d08 pc=0xf771be
github.com/glycerine/rpc25519.TestMain(0xc2000d6820)
	/home/jaten/rpc25519/cli_test.go:72 +0x35 fp=0xc200361e68 sp=0xc200361e40 pc=0xf76f75
main.main()
	_testmain.go:447 +0x172 fp=0xc200361f50 sp=0xc200361e68 pc=0x1015472
runtime.main()
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:283 +0x28b fp=0xc200361fe0 sp=0xc200361f50 pc=0x47d6ab
runtime.goexit({})
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc200361fe8 sp=0xc200361fe0 pc=0x4bd101

goroutine 2 gp=0xc200002e00 m=nil [force gc (idle)]:
runtime.gopark(0x1919610?, 0x1989a00?, 0x0?, 0x0?, 0x0?)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:435 +0xce fp=0xc2000647a8 sp=0xc200064788 pc=0x4b4d6e
runtime.goparkunlock(...)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:441
runtime.forcegchelper()
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:348 +0xb3 fp=0xc2000647e0 sp=0xc2000647a8 pc=0x47d9f3
runtime.goexit({})
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc2000647e8 sp=0xc2000647e0 pc=0x4bd101
created by runtime.init.7 in goroutine 1
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:336 +0x1a

goroutine 3 gp=0xc200003340 m=nil [GC sweep wait]:
runtime.gopark(0x1987201?, 0x0?, 0x0?, 0x0?, 0x0?)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:435 +0xce fp=0xc20007af80 sp=0xc20007af60 pc=0x4b4d6e
runtime.goparkunlock(...)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:441
runtime.bgsweep(0xc200034100)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mgcsweep.go:316 +0xdf fp=0xc20007afc8 sp=0xc20007af80 pc=0x4657ff
runtime.gcenable.gowrap1()
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mgc.go:204 +0x25 fp=0xc20007afe0 sp=0xc20007afc8 pc=0x459c85
runtime.goexit({})
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc20007afe8 sp=0xc20007afe0 pc=0x4bd101
created by runtime.gcenable in goroutine 1
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mgc.go:204 +0x66

goroutine 4 gp=0xc200003500 m=nil [runnable]:
runtime.gopark(0x10000?, 0x13c8bb8?, 0x0?, 0x0?, 0x0?)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:435 +0xce fp=0xc200065778 sp=0xc200065758 pc=0x4b4d6e
runtime.goparkunlock(...)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:441
runtime.(*scavengerState).park(0x1988240)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mgcscavenge.go:425 +0x49 fp=0xc2000657a8 sp=0xc200065778 pc=0x463269
runtime.bgscavenge(0xc200034100)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mgcscavenge.go:658 +0x59 fp=0xc2000657c8 sp=0xc2000657a8 pc=0x4637f9
runtime.gcenable.gowrap2()
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mgc.go:205 +0x25 fp=0xc2000657e0 sp=0xc2000657c8 pc=0x459c25
runtime.goexit({})
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc2000657e8 sp=0xc2000657e0 pc=0x4bd101
created by runtime.gcenable in goroutine 1
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mgc.go:205 +0xa5

goroutine 5 gp=0xc200003dc0 m=nil [finalizer wait]:
runtime.gopark(0x0?, 0x0?, 0x0?, 0x0?, 0x0?)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:435 +0xce fp=0xc200074e30 sp=0xc200074e10 pc=0x4b4d6e
runtime.runfinq()
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mfinal.go:196 +0x145 fp=0xc200074fe0 sp=0xc200074e30 pc=0x458c45
runtime.goexit({})
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc200074fe8 sp=0xc200074fe0 pc=0x4bd101
created by runtime.createfing in goroutine 1
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mfinal.go:166 +0x3d

goroutine 6 gp=0xc2002fe000 m=nil [runnable]:
runtime.gopark(0x19aa720?, 0xffffffff?, 0x69?, 0x4?, 0xc20007bf50?)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:435 +0xce fp=0xc20007bf18 sp=0xc20007bef8 pc=0x4b4d6e
runtime.chanrecv(0xc20003c1c0, 0x0, 0x1)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/chan.go:664 +0x3e5 fp=0xc20007bf90 sp=0xc20007bf18 pc=0x44aa45
runtime.chanrecv1(0x0?, 0x0?)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/chan.go:506 +0x12 fp=0xc20007bfb8 sp=0xc20007bf90 pc=0x44a632
runtime.unique_runtime_registerUniqueMapCleanup.func2(...)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mgc.go:1796
runtime.unique_runtime_registerUniqueMapCleanup.gowrap1()
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mgc.go:1799 +0x2f fp=0xc20007bfe0 sp=0xc20007bfb8 pc=0x45cdcf
runtime.goexit({})
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc20007bfe8 sp=0xc20007bfe0 pc=0x4bd101
created by unique.runtime_registerUniqueMapCleanup in goroutine 1
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mgc.go:1794 +0x79

goroutine 7 gp=0xc2002fe1c0 m=nil [select, locked to thread]:
runtime.gopark(0xc200079fa8?, 0x2?, 0x60?, 0x0?, 0xc200079f94?)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:435 +0xce fp=0xc200079de8 sp=0xc200079dc8 pc=0x4b4d6e
runtime.selectgo(0xc200079fa8, 0xc200079f90, 0x0?, 0x0, 0x2?, 0x1)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/select.go:351 +0x985 fp=0xc200079f58 sp=0xc200079de8 pc=0x491265
runtime.ensureSigM.func1()
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/signal_unix.go:1085 +0x192 fp=0xc200079fe0 sp=0xc200079f58 pc=0x4aee52
runtime.goexit({})
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc200079fe8 sp=0xc200079fe0 pc=0x4bd101
created by runtime.ensureSigM in goroutine 1
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/signal_unix.go:1068 +0xc8

goroutine 8 gp=0xc2002fe700 m=6 mp=0xc20009a808 [syscall]:
runtime.notetsleepg(0x19aafa0, 0xffffffffffffffff)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/lock_futex.go:123 +0x29 fp=0xc200064fa0 sp=0xc200064f78 pc=0x4506e9
os/signal.signal_recv()
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/sigqueue.go:152 +0x29 fp=0xc200064fc0 sp=0xc200064fa0 pc=0x4b72a9
os/signal.loop()
	/mnt/oldrog/usr/local/go1.24.3/src/os/signal/signal_unix.go:23 +0x1d fp=0xc200064fe0 sp=0xc200064fc0 pc=0x6c4e1d
runtime.goexit({})
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc200064fe8 sp=0xc200064fe0 pc=0x4bd101
created by os/signal.Notify.func1.1 in goroutine 1
	/mnt/oldrog/usr/local/go1.24.3/src/os/signal/signal.go:152 +0x47

goroutine 9 gp=0xc2002fe8c0 m=nil [chan receive]:
runtime.gopark(0x0?, 0x0?, 0x0?, 0x0?, 0x0?)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:435 +0xce fp=0xc200075e18 sp=0xc200075df8 pc=0x4b4d6e
runtime.chanrecv(0xc2001e9b90, 0xc200075f98, 0x1)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/chan.go:664 +0x3e5 fp=0xc200075e90 sp=0xc200075e18 pc=0x44aa45
runtime.chanrecv2(0x0?, 0x0?)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/chan.go:511 +0x12 fp=0xc200075eb8 sp=0xc200075e90 pc=0x44a652
github.com/glycerine/rpc25519.init.1.func1()
	/home/jaten/rpc25519/filterstacks.go:22 +0x28c fp=0xc200075fe0 sp=0xc200075eb8 pc=0xfee3cc
runtime.goexit({})
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc200075fe8 sp=0xc200075fe0 pc=0x4bd101
created by github.com/glycerine/rpc25519.init.1 in goroutine 1
	/home/jaten/rpc25519/filterstacks.go:21 +0xca

goroutine 27 gp=0xc2002ff500 m=nil [runnable]:
runtime.gopark(0xc20004b418?, 0x2?, 0x27?, 0x95?, 0xc20004b394?)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:435 +0xce fp=0xc20004b1e8 sp=0xc20004b1c8 pc=0x4b4d6e
runtime.selectgo(0xc20004b418, 0xc20004b390, 0xc200023840?, 0x0, 0xc20037fd10?, 0x1)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/select.go:351 +0x985 fp=0xc20004b358 sp=0xc20004b1e8 pc=0x491265
github.com/glycerine/rpc25519.(*Client).Start(0xc200004c08)
	/home/jaten/rpc25519/cli.go:1668 +0x4f2 fp=0xc20004b458 sp=0xc20004b358 pc=0xe8ff32
github.com/glycerine/rpc25519.Test045_upload.func1()
	/home/jaten/rpc25519/cli_test.go:1072 +0x44a fp=0xc20004b868 sp=0xc20004b458 pc=0xf7f4aa
github.com/glycerine/goconvey/convey.discover.parseAction.func1({0xc200010498?, 0xc20037e660?})
	/home/jaten/go/pkg/mod/github.com/glycerine/goconvey@v0.0.0-20190410193231-58a59202ab31/convey/discovery.go:80 +0x2f fp=0xc20004b888 sp=0xc20004b868 pc=0xe5f04f
github.com/glycerine/goconvey/convey.(*context).conveyInner(0xc2003c21e0, {0x125ca78, 0x32}, 0xc20036f130)
	/home/jaten/go/pkg/mod/github.com/glycerine/goconvey@v0.0.0-20190410193231-58a59202ab31/convey/context.go:264 +0x2d1 fp=0xc20004b9d0 sp=0xc20004b888 pc=0xe5e091
github.com/glycerine/goconvey/convey.rootConvey.func1()
	/home/jaten/go/pkg/mod/github.com/glycerine/goconvey@v0.0.0-20190410193231-58a59202ab31/convey/context.go:113 +0x1a5 fp=0xc20004ba88 sp=0xc20004b9d0 pc=0xe5c3e5
github.com/jtolds/gls.(*ContextManager).SetValues.func1(0x0)
	/home/jaten/go/pkg/mod/github.com/jtolds/gls@v4.20.0+incompatible/context.go:97 +0x66a fp=0xc20004bc58 sp=0xc20004ba88 pc=0xe58aaa
github.com/jtolds/gls.EnsureGoroutineId.func1()
	/home/jaten/go/pkg/mod/github.com/jtolds/gls@v4.20.0+incompatible/gid.go:24 +0x39 fp=0xc20004bc80 sp=0xc20004bc58 pc=0xe590d9
github.com/jtolds/gls._m(0x0, 0xc2000104c8)
	/home/jaten/go/pkg/mod/github.com/jtolds/gls@v4.20.0+incompatible/stack_tags.go:108 +0x39 fp=0xc20004bca8 sp=0xc20004bc80 pc=0xe5b959
github.com/jtolds/gls.github_com_jtolds_gls_markS(0x0, 0xc2000104c8)
	/home/jaten/go/pkg/mod/github.com/jtolds/gls@v4.20.0+incompatible/stack_tags.go:56 +0x34 fp=0xc20004bcc8 sp=0xc20004bca8 pc=0xe5b2f4
github.com/jtolds/gls.addStackTag(...)
	/home/jaten/go/pkg/mod/github.com/jtolds/gls@v4.20.0+incompatible/stack_tags.go:49
github.com/jtolds/gls.EnsureGoroutineId(0xc20037e5a0)
	/home/jaten/go/pkg/mod/github.com/jtolds/gls@v4.20.0+incompatible/gid.go:24 +0x165 fp=0xc20004bd48 sp=0xc20004bcc8 pc=0xe59045
github.com/jtolds/gls.(*ContextManager).SetValues(0xc200046850, 0xc20037e540, 0xc200374280)
	/home/jaten/go/pkg/mod/github.com/jtolds/gls@v4.20.0+incompatible/context.go:63 +0x285 fp=0xc20004bd98 sp=0xc20004bd48 pc=0xe583e5
github.com/glycerine/goconvey/convey.rootConvey({0xc2003d6ea0, 0x3, 0x3})
	/home/jaten/go/pkg/mod/github.com/glycerine/goconvey@v0.0.0-20190410193231-58a59202ab31/convey/context.go:108 +0x465 fp=0xc20004be50 sp=0xc20004bd98 pc=0xe5c165
github.com/glycerine/goconvey/convey.Convey({0xc2003d6ea0, 0x3, 0x3})
	/home/jaten/go/pkg/mod/github.com/glycerine/goconvey@v0.0.0-20190410193231-58a59202ab31/convey/doc.go:73 +0x9b fp=0xc20004be80 sp=0xc20004be50 pc=0xe5f11b
github.com/glycerine/rpc25519.Test045_upload(0xc2003ad340)
	/home/jaten/rpc25519/cli_test.go:1050 +0x10f fp=0xc20004bee0 sp=0xc20004be80 pc=0xf7f02f
testing.tRunner(0xc2003ad340, 0x1273f98)
	/mnt/oldrog/usr/local/go1.24.3/src/testing/testing.go:1792 +0x226 fp=0xc20004bfb0 sp=0xc20004bee0 pc=0x60c8c6
testing.(*T).Run.gowrap1()
	/mnt/oldrog/usr/local/go1.24.3/src/testing/testing.go:1851 +0x45 fp=0xc20004bfe0 sp=0xc20004bfb0 pc=0x60e625
runtime.goexit({})
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc20004bfe8 sp=0xc20004bfe0 pc=0x4bd101
created by testing.(*T).Run in goroutine 1
	/mnt/oldrog/usr/local/go1.24.3/src/testing/testing.go:1851 +0x8f3

goroutine 28 gp=0xc2002ff880 m=nil [IO wait]:
runtime.gopark(0x5541a5216d44b361?, 0x6d30eac961e69813?, 0xd1?, 0x61?, 0xc200300300?)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:435 +0xce fp=0xc80142b7d0 sp=0xc80142b7b0 pc=0x4b4d6e
runtime.netpollblock(0x4c0431?, 0x527da5?, 0x0?)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/netpoll.go:575 +0xf7 fp=0xc80142b808 sp=0xc80142b7d0 pc=0x476037
internal/poll.runtime_pollWait(0x20c3022b8e90, 0x72)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/netpoll.go:351 +0x85 fp=0xc80142b828 sp=0xc80142b808 pc=0x4b3f65
internal/poll.(*pollDesc).wait(0xc200300320, 0x72, 0x0)
	/mnt/oldrog/usr/local/go1.24.3/src/internal/poll/fd_poll_runtime.go:84 +0xb1 fp=0xc80142b868 sp=0xc80142b828 pc=0x51fe11
internal/poll.(*pollDesc).waitRead(...)
	/mnt/oldrog/usr/local/go1.24.3/src/internal/poll/fd_poll_runtime.go:89
internal/poll.(*FD).Accept(0xc200300300)
	/mnt/oldrog/usr/local/go1.24.3/src/internal/poll/fd_unix.go:620 +0x505 fp=0xc80142b970 sp=0xc80142b868 pc=0x527da5
net.(*netFD).accept(0xc200300300)
	/mnt/oldrog/usr/local/go1.24.3/src/net/fd_unix.go:172 +0x45 fp=0xc80142ba80 sp=0xc80142b970 pc=0x83c765
net.(*TCPListener).accept(0xc200023540)
	/mnt/oldrog/usr/local/go1.24.3/src/net/tcpsock_posix.go:159 +0x4e fp=0xc80142bb20 sp=0xc80142ba80 pc=0x861f6e
net.(*TCPListener).Accept(0xc200023540)
	/mnt/oldrog/usr/local/go1.24.3/src/net/tcpsock.go:380 +0x65 fp=0xc80142bb90 sp=0xc80142bb20 pc=0x860325
crypto/tls.(*listener).Accept(0xc200010840)
	/mnt/oldrog/usr/local/go1.24.3/src/crypto/tls/tls.go:67 +0x43 fp=0xc80142bbe8 sp=0xc80142bb90 pc=0x9762a3
github.com/glycerine/rpc25519.(*Server).runServerMain(0xc2003ad500, {0x123495a, 0xb}, 0x0, {0x0, 0x0}, 0xc200334a10)
	/home/jaten/rpc25519/srv.go:218 +0x12e3 fp=0xc80142bf60 sp=0xc80142bbe8 pc=0xf38f03
github.com/glycerine/rpc25519.(*Server).Start.gowrap1()
	/home/jaten/rpc25519/srv.go:2281 +0x8f fp=0xc80142bfe0 sp=0xc80142bf60 pc=0xf4afcf
runtime.goexit({})
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc80142bfe8 sp=0xc80142bfe0 pc=0x4bd101
created by github.com/glycerine/rpc25519.(*Server).Start in goroutine 27
	/home/jaten/rpc25519/srv.go:2281 +0xa7e

goroutine 12 gp=0xc2002ffa40 m=nil [runnable]:
github.com/klauspost/compress/zstd.(*bestFastEncoder).Reset(0xc801ca0000, 0x0, 0x0)
	/home/jaten/go/pkg/mod/github.com/klauspost/compress@v1.17.11/zstd/enc_best.go:488 +0x1045 fp=0xc200365368 sp=0xc200365360 pc=0xcbdac5
github.com/klauspost/compress/zstd.(*Encoder).Reset(0xc200328280, {0x13cff40, 0x19aa2c0})
	/home/jaten/go/pkg/mod/github.com/klauspost/compress@v1.17.11/zstd/encoder.go:122 +0x6ec fp=0xc200365488 sp=0xc200365368 pc=0xcdc84c
github.com/klauspost/compress/zstd.NewWriter({0x13cff40, 0x19aa2c0}, {0xc2003655d8, 0x1, 0xc200365548?})
	/home/jaten/go/pkg/mod/github.com/klauspost/compress@v1.17.11/zstd/encoder.go:78 +0x1ea fp=0xc200365510 sp=0xc200365488 pc=0xcdbe8a
github.com/glycerine/rpc25519.newPressor(0x140000)
	/home/jaten/rpc25519/compress.go:134 +0x36f fp=0xc200365630 sp=0xc200365510 pc=0xe9d6cf
github.com/glycerine/rpc25519.newWorkspace({0xc202b7e030, 0x14}, 0x13ffb0, 0x0, 0xc2002fa420, 0x0, 0xc2000308e8)
	/home/jaten/rpc25519/common.go:148 +0x445 fp=0xc2003656e0 sp=0xc200365630 pc=0xe9afa5
github.com/glycerine/rpc25519.newBlabber({0x1237ff3, 0x10}, {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, ...}, ...)
	/home/jaten/rpc25519/chacha.go:246 +0x912 fp=0xc200365808 sp=0xc2003656e0 pc=0xe62e12
github.com/glycerine/rpc25519.(*Client).runReadLoop(0xc200004808, {0x13d9508, 0xc2001c7888}, 0xc2000308e8)
	/home/jaten/rpc25519/cli.go:342 +0x316 fp=0xc200365a08 sp=0xc200365808 pc=0xe88a56
github.com/glycerine/rpc25519.(*Client).runClientMain(0xc200004808, {0xc20035c7e1, 0xf}, 0x0, {0x0, 0x0})
	/home/jaten/rpc25519/cli.go:248 +0x16ef fp=0xc200365f70 sp=0xc200365a08 pc=0xe871cf
github.com/glycerine/rpc25519.(*Client).Start.gowrap1()
	/home/jaten/rpc25519/cli.go:1665 +0x79 fp=0xc200365fe0 sp=0xc200365f70 pc=0xe90019
runtime.goexit({})
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc200365fe8 sp=0xc200365fe0 pc=0x4bd101
created by github.com/glycerine/rpc25519.(*Client).Start in goroutine 10
	/home/jaten/rpc25519/cli.go:1665 +0x3f6

goroutine 31 gp=0xc2003ac000 m=nil [select]:
runtime.gopark(0xc2003d3fa0?, 0x2?, 0x0?, 0x3e?, 0xc2003d3f64?)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:435 +0xce fp=0xc2003d3dc0 sp=0xc2003d3da0 pc=0x4b4d6e
runtime.selectgo(0xc2003d3fa0, 0xc2003d3f60, 0x29?, 0x0, 0x1?, 0x1)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/select.go:351 +0x985 fp=0xc2003d3f30 sp=0xc2003d3dc0 pc=0x491265
crypto/tls.(*Conn).handshakeContext.func2()
	/mnt/oldrog/usr/local/go1.24.3/src/crypto/tls/conn.go:1544 +0xba fp=0xc2003d3fe0 sp=0xc2003d3f30 pc=0x8fa3fa
runtime.goexit({})
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc2003d3fe8 sp=0xc2003d3fe0 pc=0x4bd101
created by crypto/tls.(*Conn).handshakeContext in goroutine 30
	/mnt/oldrog/usr/local/go1.24.3/src/crypto/tls/conn.go:1543 +0x445

goroutine 20 gp=0xc2003ac1c0 m=nil [select]:
runtime.gopark(0xc200115eb0?, 0x3?, 0xb8?, 0x5b?, 0xc200115daa?)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:435 +0xce fp=0xc200115b80 sp=0xc200115b60 pc=0x4b4d6e
runtime.selectgo(0xc200115eb0, 0xc200115da4, 0x0?, 0x0, 0x0?, 0x1)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/select.go:351 +0x985 fp=0xc200115cf0 sp=0xc200115b80 pc=0x491265
github.com/glycerine/rpc25519.(*rwPair).runSendLoop(0xc2003ea008, {0x13d9508, 0xc2001c7c08})
	/home/jaten/rpc25519/srv.go:498 +0x7c5 fp=0xc200115fa0 sp=0xc200115cf0 pc=0xf3b245
github.com/glycerine/rpc25519.(*Server).serveConn.gowrap1()
	/home/jaten/rpc25519/srv.go:361 +0x50 fp=0xc200115fe0 sp=0xc200115fa0 pc=0xf3a770
runtime.goexit({})
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc200115fe8 sp=0xc200115fe0 pc=0x4bd101
created by github.com/glycerine/rpc25519.(*Server).serveConn in goroutine 13
	/home/jaten/rpc25519/srv.go:361 +0x3b7

goroutine 29 gp=0xc2003ac380 m=9 mp=0xc707ffe008 [running]:
	goroutine running on other thread; stack unavailable
created by github.com/glycerine/rpc25519.(*Client).Start in goroutine 27
	/home/jaten/rpc25519/cli.go:1665 +0x3f6

goroutine 16 gp=0xc2003ac540 m=nil [GC worker (idle)]:
runtime.gopark(0x3b7046f4e7f?, 0xc2003cc160?, 0x1b?, 0xa?, 0x0?)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:435 +0xce fp=0xc2003d7f38 sp=0xc2003d7f18 pc=0x4b4d6e
runtime.gcBgMarkWorker(0xc20003cee0)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mgc.go:1423 +0xe9 fp=0xc2003d7fc8 sp=0xc2003d7f38 pc=0x45c0e9
runtime.gcBgMarkStartWorkers.gowrap1()
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mgc.go:1339 +0x25 fp=0xc2003d7fe0 sp=0xc2003d7fc8 pc=0x45bfc5
runtime.goexit({})
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc2003d7fe8 sp=0xc2003d7fe0 pc=0x4bd101
created by runtime.gcBgMarkStartWorkers in goroutine 12
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/mgc.go:1339 +0x105

goroutine 30 gp=0xc2003ac700 m=nil [runnable]:
runtime.gopark(0x0?, 0x9?, 0x0?, 0x10?, 0x800?)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:435 +0xce fp=0xc80144c9b0 sp=0xc80144c990 pc=0x4b4d6e
runtime.netpollblock(0x4c0431?, 0x5219f3?, 0x0?)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/netpoll.go:575 +0xf7 fp=0xc80144c9e8 sp=0xc80144c9b0 pc=0x476037
internal/poll.runtime_pollWait(0x20c3022b8a30, 0x72)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/netpoll.go:351 +0x85 fp=0xc80144ca08 sp=0xc80144c9e8 pc=0x4b3f65
internal/poll.(*pollDesc).wait(0xc200300620, 0x72, 0x0)
	/mnt/oldrog/usr/local/go1.24.3/src/internal/poll/fd_poll_runtime.go:84 +0xb1 fp=0xc80144ca48 sp=0xc80144ca08 pc=0x51fe11
internal/poll.(*pollDesc).waitRead(...)
	/mnt/oldrog/usr/local/go1.24.3/src/internal/poll/fd_poll_runtime.go:89
internal/poll.(*FD).Read(0xc200300600, {0xc801431000, 0x800, 0x800})
	/mnt/oldrog/usr/local/go1.24.3/src/internal/poll/fd_unix.go:165 +0x453 fp=0xc80144cb38 sp=0xc80144ca48 pc=0x5219f3
net.(*netFD).Read(0xc200300600, {0xc801431000, 0x800, 0x800})
	/mnt/oldrog/usr/local/go1.24.3/src/net/fd_posix.go:55 +0x4b fp=0xc80144cb88 sp=0xc80144cb38 pc=0x8391eb
net.(*conn).Read(0xc200068240, {0xc801431000, 0x800, 0x800})
	/mnt/oldrog/usr/local/go1.24.3/src/net/net.go:194 +0xad fp=0xc80144cc08 sp=0xc80144cb88 pc=0x85220d
crypto/tls.(*atLeastReader).Read(0xc200010d08, {0xc801431000, 0x800, 0x800})
	/mnt/oldrog/usr/local/go1.24.3/src/crypto/tls/conn.go:809 +0x84 fp=0xc80144cc70 sp=0xc80144cc08 pc=0x8f1fc4
bytes.(*Buffer).ReadFrom(0xc2001c62b8, {0x13d0680, 0xc200010d08})
	/mnt/oldrog/usr/local/go1.24.3/src/bytes/buffer.go:211 +0x10f fp=0xc80144cce0 sp=0xc80144cc70 pc=0x5d662f
crypto/tls.(*Conn).readFromUntil(0xc2001c6008, {0x13d0860, 0xc200068240}, 0x5)
	/mnt/oldrog/usr/local/go1.24.3/src/crypto/tls/conn.go:831 +0x1d6 fp=0xc80144cd50 sp=0xc80144cce0 pc=0x8f2376
crypto/tls.(*Conn).readRecordOrCCS(0xc2001c6008, 0x0)
	/mnt/oldrog/usr/local/go1.24.3/src/crypto/tls/conn.go:629 +0x359 fp=0xc80144d220 sp=0xc80144cd50 pc=0x8ed439
crypto/tls.(*Conn).readRecord(...)
	/mnt/oldrog/usr/local/go1.24.3/src/crypto/tls/conn.go:591
crypto/tls.(*Conn).readHandshakeBytes(0xc2001c6008, 0x4)
	/mnt/oldrog/usr/local/go1.24.3/src/crypto/tls/conn.go:1078 +0xcc fp=0xc80144d270 sp=0xc80144d220 pc=0x8f4bec
crypto/tls.(*Conn).readHandshake(0xc2001c6008, {0x2d6d35afa218, 0xc2003c6280})
	/mnt/oldrog/usr/local/go1.24.3/src/crypto/tls/conn.go:1089 +0x57 fp=0xc80144d3c0 sp=0xc80144d270 pc=0x8f4cb7
crypto/tls.(*serverHandshakeStateTLS13).readClientCertificate(0xc80144d870)
	/mnt/oldrog/usr/local/go1.24.3/src/crypto/tls/handshake_server_tls13.go:1053 +0x29f fp=0xc80144d790 sp=0xc80144d3c0 pc=0x963e7f
crypto/tls.(*serverHandshakeStateTLS13).handshake(0xc80144d870)
	/mnt/oldrog/usr/local/go1.24.3/src/crypto/tls/handshake_server_tls13.go:94 +0x125 fp=0xc80144d848 sp=0xc80144d790 pc=0x956ec5
crypto/tls.(*Conn).serverHandshake(0xc2001c6008, {0x13d6ae0, 0xc2001f9040})
	/mnt/oldrog/usr/local/go1.24.3/src/crypto/tls/handshake_server.go:56 +0x25d fp=0xc80144da30 sp=0xc80144d848 pc=0x94b75d
crypto/tls.(*Conn).serverHandshake-fm({0x13d6ae0, 0xc2001f9040})
	<autogenerated>:1 +0x48 fp=0xc80144da70 sp=0xc80144da30 pc=0x97f488
crypto/tls.(*Conn).handshakeContext(0xc2001c6008, {0x13d6b50, 0xc200335180})
	/mnt/oldrog/usr/local/go1.24.3/src/crypto/tls/conn.go:1568 +0x603 fp=0xc80144dcd8 sp=0xc80144da70 pc=0x8f97a3
crypto/tls.(*Conn).HandshakeContext(...)
	/mnt/oldrog/usr/local/go1.24.3/src/crypto/tls/conn.go:1508
github.com/glycerine/rpc25519.(*Server).handleTLSConnection(0xc2003ad500, 0xc2001c6008)
	/home/jaten/rpc25519/srv.go:388 +0xa5 fp=0xc80144dfb0 sp=0xc80144dcd8 pc=0xf3a845
github.com/glycerine/rpc25519.(*Server).runServerMain.gowrap3()
	/home/jaten/rpc25519/srv.go:236 +0x45 fp=0xc80144dfe0 sp=0xc80144dfb0 pc=0xf393a5
runtime.goexit({})
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc80144dfe8 sp=0xc80144dfe0 pc=0x4bd101
created by github.com/glycerine/rpc25519.(*Server).runServerMain in goroutine 28
	/home/jaten/rpc25519/srv.go:236 +0x13cb

goroutine 21 gp=0xc2003aca80 m=nil [IO wait]:
runtime.gopark(0x0?, 0x8?, 0x0?, 0xd0?, 0x800?)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:435 +0xce fp=0xc200359278 sp=0xc200359258 pc=0x4b4d6e
runtime.netpollblock(0x4c0431?, 0x5219f3?, 0x0?)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/netpoll.go:575 +0xf7 fp=0xc2003592b0 sp=0xc200359278 pc=0x476037
internal/poll.runtime_pollWait(0x20c3022b8c60, 0x72)
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/netpoll.go:351 +0x85 fp=0xc2003592d0 sp=0xc2003592b0 pc=0x4b3f65
internal/poll.(*pollDesc).wait(0xc200300ca0, 0x72, 0x0)
	/mnt/oldrog/usr/local/go1.24.3/src/internal/poll/fd_poll_runtime.go:84 +0xb1 fp=0xc200359310 sp=0xc2003592d0 pc=0x51fe11
internal/poll.(*pollDesc).waitRead(...)
	/mnt/oldrog/usr/local/go1.24.3/src/internal/poll/fd_poll_runtime.go:89
internal/poll.(*FD).Read(0xc200300c80, {0xc20037d000, 0x800, 0x800})
	/mnt/oldrog/usr/local/go1.24.3/src/internal/poll/fd_unix.go:165 +0x453 fp=0xc200359400 sp=0xc200359310 pc=0x5219f3
net.(*netFD).Read(0xc200300c80, {0xc20037d000, 0x800, 0x800})
	/mnt/oldrog/usr/local/go1.24.3/src/net/fd_posix.go:55 +0x4b fp=0xc200359450 sp=0xc200359400 pc=0x8391eb
net.(*conn).Read(0xc200068480, {0xc20037d000, 0x800, 0x800})
	/mnt/oldrog/usr/local/go1.24.3/src/net/net.go:194 +0xad fp=0xc2003594d0 sp=0xc200359450 pc=0x85220d
crypto/tls.(*atLeastReader).Read(0xc2000103f0, {0xc20037d000, 0x800, 0x800})
	/mnt/oldrog/usr/local/go1.24.3/src/crypto/tls/conn.go:809 +0x84 fp=0xc200359538 sp=0xc2003594d0 pc=0x8f1fc4
bytes.(*Buffer).ReadFrom(0xc2001c7eb8, {0x13d0680, 0xc2000103f0})
	/mnt/oldrog/usr/local/go1.24.3/src/bytes/buffer.go:211 +0x10f fp=0xc2003595a8 sp=0xc200359538 pc=0x5d662f
crypto/tls.(*Conn).readFromUntil(0xc2001c7c08, {0x13d0860, 0xc200068480}, 0x5)
	/mnt/oldrog/usr/local/go1.24.3/src/crypto/tls/conn.go:831 +0x1d6 fp=0xc200359618 sp=0xc2003595a8 pc=0x8f2376
crypto/tls.(*Conn).readRecordOrCCS(0xc2001c7c08, 0x0)
	/mnt/oldrog/usr/local/go1.24.3/src/crypto/tls/conn.go:629 +0x359 fp=0xc200359ae8 sp=0xc200359618 pc=0x8ed439
crypto/tls.(*Conn).readRecord(...)
	/mnt/oldrog/usr/local/go1.24.3/src/crypto/tls/conn.go:591
crypto/tls.(*Conn).Read(0xc2001c7c08, {0xc20035c568, 0x8, 0x0?})
	/mnt/oldrog/usr/local/go1.24.3/src/crypto/tls/conn.go:1385 +0x2d0 fp=0xc200359bc0 sp=0xc200359ae8 pc=0x8f8470
github.com/glycerine/rpc25519.readFull({0x13d69c8, 0xc2001c7c08}, {0xc20035c568, 0x8, 0x8})
	/home/jaten/rpc25519/common.go:391 +0xe6 fp=0xc200359c08 sp=0xc200359bc0 pc=0xe9c786
github.com/glycerine/rpc25519.(*workspace).readMessage(0xc20037a270, {0x13d69c8, 0xc2001c7c08})
	/home/jaten/rpc25519/common.go:193 +0x86 fp=0xc200359d10 sp=0xc200359c08 pc=0xe9b386
github.com/glycerine/rpc25519.(*blabber).readMessage(0xc2001f8500, {0x13d69c8, 0xc2001c7c08})
	/home/jaten/rpc25519/chacha.go:292 +0xe5 fp=0xc200359d58 sp=0xc200359d10 pc=0xe63665
github.com/glycerine/rpc25519.(*rwPair).runReadLoop(0xc2003ea008, {0x13d9508, 0xc2001c7c08})
	/home/jaten/rpc25519/srv.go:582 +0x525 fp=0xc200359fa0 sp=0xc200359d58 pc=0xf3bf65
github.com/glycerine/rpc25519.(*Server).serveConn.gowrap2()
	/home/jaten/rpc25519/srv.go:362 +0x50 fp=0xc200359fe0 sp=0xc200359fa0 pc=0xf3a6f0
runtime.goexit({})
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc200359fe8 sp=0xc200359fe0 pc=0x4bd101
created by github.com/glycerine/rpc25519.(*Server).serveConn in goroutine 13
	/home/jaten/rpc25519/srv.go:362 +0x485

goroutine 32 gp=0xc2003ad6c0 m=nil [runnable]:
github.com/glycerine/rpc25519.(*Client).runClientMain.gowrap2()
	/home/jaten/rpc25519/cli.go:247 fp=0xc200062fe0 sp=0xc200062fd8 pc=0xe87580
runtime.goexit({})
	/mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:1700 +0x1 fp=0xc200062fe8 sp=0xc200062fe0 pc=0x4bd101
created by github.com/glycerine/rpc25519.(*Client).runClientMain in goroutine 29
	/home/jaten/rpc25519/cli.go:247 +0x16cb

rax    0xc200055750
rbx    0x218e08000000
rcx    0x0
rdx    0x0
rdi    0x218e08010800
rsi    0x8
rbp    0x7ba951eb3cd0
rsp    0x7ba951eb3c58
r8     0x218e080108ff
r9     0x0
r10    0x0
r11    0x615ff
r12    0x7ba951eb3d60
r13    0xb
r14    0xc200003180
r15    0x3
rip    0x45e574
rflags 0x10287
cs     0x33
fs     0x0
gs     0x0

Compilation exited abnormally with code 2 at Sun Jun  8 18:10:21

~~~

I analyzed the recorded trace in the following rr replay 
session. rr is integrated with gdb, despite the (rr) prompt.

See the full back and forth with gemini in the 
debug session link for the step-by-step
detective work. It has the play-by-play commentary.
Gemini was pretty essential to navigating around
gdb and rr.

~~~
jaten@rog ~/rpc25519 (master) $ rr pack
rr: Packed trace directory `/mnt/oldrog/home/jaten/go/src/github.com/glycerine/rr_binary_for_issue74019/rr/rpc.test-24'.
jaten@rog ~/rpc25519 (master) $ rr replay
GNU gdb (Ubuntu 15.0.50.20240403-0ubuntu1) 15.0.50.20240403-git
Copyright (C) 2024 Free Software Foundation, Inc.
License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses/gpl.html>
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.
Type "show copying" and "show warranty" for details.
This GDB was configured as "x86_64-linux-gnu".
Type "show configuration" for configuration details.
For bug reporting instructions, please see:
<https://www.gnu.org/software/gdb/bugs/>.
Find the GDB manual and other documentation resources online at:
    <http://www.gnu.org/software/gdb/documentation/>.

For help, type "help".
Type "apropos word" to search for commands related to "word"...
Reading symbols from /mnt/oldrog/home/jaten/go/src/github.com/glycerine/rr_binary_for_issue74019/rr/rpc.test-24/mmap_pack_0_rpc.test...
Redefine command "hook-stop"? (y or n) [answered Y; input not from terminal]
Loading Go Runtime support.
Remote debugging using 127.0.0.1:7306
Reading symbols from /lib64/ld-linux-x86-64.so.2...
Reading symbols from /usr/lib/debug/.build-id/1c/8db5f83bba514f8fd5f1fb6d7be975be1bb855.debug...
BFD: warning: system-supplied DSO at 0x6fffd000 has a section extending past end of file
Downloading separate debug info for system-supplied DSO at 0x6fffd000
0x00007d9f956e2540 in _start () from /lib64/ld-linux-x86-64.so.2
(rr) c

Continuing.
faketime = false
=== RUN   Test040_remote_cancel_by_context

cli_test.go:948 [goID 10] 2025-06-08 23:10:18.343620423 +0000 UTC server Start() returned serverAddr = '127.0.0.1:34869'

cli.go:211 [goID 12] 2025-06-08 23:10:18.366232721 +0000 UTC connected to server 127.0.0.1:34869

cli_test.go:982 [goID 10] 2025-06-08 23:10:18.366636546 +0000 UTC cli_test 040: about to block on test040callStarted

cli_test.go:974 [goID 18] 2025-06-08 23:10:18.367014513 +0000 UTC client.Call() goro top about to call over net/rpc: MustBeCancelled.WillHangUntilCancel()
example.go:300 server-side: WillHangUntilCancel() is running
example.go:304 net.rpc API: our net.Conn has 
local = '127.0.0.1:34869'
remote = '127.0.0.1:51946'

cli_test.go:984 [goID 10] 2025-06-08 23:10:21.227595976 +0000 UTC cli_test 040: we got past test040callStarted

cli_test.go:988 [goID 10] 2025-06-08 23:10:21.227787850 +0000 UTC past cancelFunc()

cli_test.go:977 [goID 18] 2025-06-08 23:10:21.231561912 +0000 UTC client.Call() returned with cliErr = 'cancellation request sent'

cli_test.go:991 [goID 10] 2025-06-08 23:10:21.231755899 +0000 UTC past cliErrIsSet channel; cliErr40 = 'cancellation request sent'

cli_test.go:998 [goID 10] 2025-06-08 23:10:21.231934678 +0000 UTC about to verify that server side context was cancelled.
example.go:310 MustBeCancelled.WillHangUntilCancel(): ctx.Done() was closed!

cli_test.go:1000 [goID 10] 2025-06-08 23:10:21.237329200 +0000 UTC server side saw the cancellation request: confirmed.
example.go:324 server-side: MessageAPI_HangUntilCancel() is running
example.go:326: Message API: our net.Conn has local = '127.0.0.1:34869'; remote = '127.0.0.1:51946'

cli_test.go:1020 [goID 10] 2025-06-08 23:10:21.243213831 +0000 UTC cli_test 041: we got past test041callStarted

cli_test.go:1024 [goID 10] 2025-06-08 23:10:21.243403981 +0000 UTC past cancelFunc()

cli_test.go:1014 [goID 24] 2025-06-08 23:10:21.243850237 +0000 UTC client.Call() returned with cliErr = 'cancellation request sent'

cli_test.go:1027 [goID 10] 2025-06-08 23:10:21.244036540 +0000 UTC past cliErrIsSet channel; cliErr = 'cancellation request sent'

cli_test.go:1038 [goID 10] 2025-06-08 23:10:21.244213305 +0000 UTC about to verify that server side context was cancelled.
example.go: MustBeCancelled.MessageAPI_HangUntilCancel(): ctx.Done() was closed!

0 assertions thus far

--- PASS: Test040_remote_cancel_by_context (2.91s)
=== RUN   Test045_upload

cli.go:211 [goID 29] 2025-06-08 23:10:21.313925974 +0000 UTC connected to server 127.0.0.1:38633
[New Thread 7129.7131]
[New Thread 7129.7130]
[New Thread 7129.7132]
[New Thread 7129.7133]
[New Thread 7129.7134]
[New Thread 7129.7135]
[New Thread 7129.7136]
[New Thread 7129.7137]

Thread 2 received signal SIGSEGV, Segmentation fault.
[Switching to Thread 7129.7131]
runtime.markrootSpans (gcw=0xc200055750, shard=<optimized out>)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/mgcmark.go:370
370				if state := s.state.get(); state != mSpanInUse {

(rr) list

365				// about the span being freed and re-used.
366				s := ha.spans[arenaPage+uint(i)*8+j]
367	
368				// The state must be mSpanInUse if the specials bit is set, so
369				// sanity check that.
370				if state := s.state.get(); state != mSpanInUse {
371					print("s.state = ", state, "\n")
372					throw("non in-use span found with specials bit set")
373				}
374				// Check that this span was swept (it may be cached or uncached).
(rr) p s

$1 = (runtime.mspan *) 0x615ff
(rr) p s.state

Cannot access memory at address 0x61662
(rr) b 366

Breakpoint 1 at 0x45e55b: file /mnt/oldrog/usr/local/go1.24.3/src/runtime/mgcmark.go, line 366.
(rr) # We just set a breakpoint at the line that assigns the bad pointer

(rr) # Now, run backwards until we hit the breakpoint at mgcmark.go:366

(rr) reverse-continue

Continuing.

Thread 2 received signal SIGSEGV, Segmentation fault.
runtime.markrootSpans (gcw=0xc200055750, shard=<optimized out>)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/mgcmark.go:370
370				if state := s.state.get(); state != mSpanInUse {

(rr) # use reverse-step (rs) instead because reverse-continue can sometimes be too coarse when the crash and the breakpoint are very close together.

(rr) rs

370				if state := s.state.get(); state != mSpanInUse {

(rr) rs

runtime.(*mSpanStateBox).get (b=<optimized out>)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/mheap.go:392
392		return mSpanState(b.s.Load())
(rr) list

387	// It is nosplit because it's called indirectly by typedmemclr,
388	// which must not be preempted.
389	
390	//go:nosplit
391	func (b *mSpanStateBox) get() mSpanState {
392		return mSpanState(b.s.Load())
393	}
394	
395	// mSpanList heads a linked list of spans.
396	type mSpanList struct {
(rr) rs


Thread 2 hit Breakpoint 1, runtime.markrootSpans (gcw=0xc200055750, 
    shard=<optimized out>)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/mgcmark.go:366
366				s := ha.spans[arenaPage+uint(i)*8+j]

(rr) ## looks like gdb just got a little confused about how to map program counter to source for a moment, but we're back on track now

(rr) p i

$2 = 0
(rr) p j

$3 = 0
(rr) p arenaPage

$4 = 0
(rr) p ha

$5 = (runtime.heapArena *) 0x218e08000000
(rr) p ha.spans

$6 = {0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 
  0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 
  0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 
  0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 
  0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 
  0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 
  0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 
  0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 
  0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 
  0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 
  0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 
  0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 
  0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 
  0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 
  0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 
  0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 
  0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 
  0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 0x0, 0x615ff, 
  0x0, 0x615ff, 0x0...}
(rr) arenaPage + i*8 + j

Undefined command: "arenaPage".  Try "help".
(rr) p arenaPage + i*8 + j

$7 = 0
(rr) x/gx ha.spans.array

There is no member named array.
(rr) x/gx ha.spans

0x218e08000000:	0x00000000000615ff
(rr) # Set a Memory Watchpoint on the contents of 0x218e08000000

(rr) watch *0x218e08000000

Hardware watchpoint 2: *0x218e08000000
(rr) # run backwards to find who wrote the 0x615ff

(rr) reverse-continue

Continuing.

Thread 2 hit Breakpoint 1, runtime.markrootSpans (gcw=0xc200055750, 
    shard=<optimized out>)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/mgcmark.go:366
366				s := ha.spans[arenaPage+uint(i)*8+j]

(rr) disable 1

(rr) reverse-continue

Continuing.
[Thread 7129.7136 exited]
[Thread 7129.7137 exited]
[Switching to Thread 7129.7129]

Thread 1 hit Hardware watchpoint 2: *0x218e08000000

Old value = 398847
New value = 0
0x0000000000435468 in void __tsan::MemoryAccessRangeT<false>(__tsan::ThreadState*, unsigned long, unsigned long, unsigned long) ()

(rr) list

361				//
362				// This value is guaranteed to be non-nil because having
363				// specials implies that the span is in-use, and since we're
364				// currently marking we can be sure that we don't have to worry
365				// about the span being freed and re-used.
366				s := ha.spans[arenaPage+uint(i)*8+j]
367	
368				// The state must be mSpanInUse if the specials bit is set, so
369				// sanity check that.
370				if state := s.state.get(); state != mSpanInUse {
(rr) ## 398847 in decimal is 0x615ff in hex

(rr) bt

#0  0x0000000000435468 in void __tsan::MemoryAccessRangeT<false>(__tsan::ThreadState*, unsigned long, unsigned long, unsigned long) ()
#1  0x0000000002200078 in __tsan::ctx_placeholder ()
#2  0x000000c2003598a8 in ?? ()
#3  0x000000c2003aca80 in ?? ()
#4  0x00000000004be8ba in racecall ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/race_amd64.s:447
#5  0x00000000004bb14a in runtime.systemstack ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:514
#6  0x00007ffe1ffb4c38 in ?? ()
#7  0x00000000004c03bf in runtime.newproc (fn=0x4bafcf <runtime.rt0_go+303>)
    at <autogenerated>:1
#8  0x00000000004bb045 in runtime.mstart ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:395
#9  0x00000000004bafcf in runtime.rt0_go ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:358
#10 0x0000000000000004 in ?? ()
#11 0x00007ffe1ffb4c88 in ?? ()
#12 0x0000000000000000 in ?? ()
(rr) i thr

  Id   Target Id                   Frame 
* 1    Thread 7129.7129 (rpc.test) 0x0000000000435468 in void __tsan::MemoryAccessRangeT<false>(__tsan::ThreadState*, unsigned long, unsigned long, unsigned long) ()
  2    Thread 7129.7131 (rpc.test) 0x0000000000435468 in void __tsan::MemoryAccessRangeT<false>(__tsan::ThreadState*, unsigned long, unsigned long, unsigned long) ()
  3    Thread 7129.7130 (rpc.test) 0x000000000048aa27 in runtime.retake (
    now=4084794783364, ~r0=<optimized out>)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:6258
  4    Thread 7129.7132 (rpc.test) runtime.futex ()
--Type <RET> for more, q to quit, c to continue without paging-- 

    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/sys_linux_amd64.s:558
  5    Thread 7129.7133 (rpc.test) runtime.futex ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/sys_linux_amd64.s:558
  6    Thread 7129.7134 (rpc.test) runtime.futex ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/sys_linux_amd64.s:558
  7    Thread 7129.7135 (rpc.test) 0x0000000070000002 in syscall_traced ()
(rr) # gemini argues: "the backtrace showing runtime.mstart -> runtime.rt0_go has a very specific meaning."

(rr) # gemini: "runtime.mstart is the function that an OS thread (M in Go's M:N scheduler model) executes when it first starts up. It then calls runtime.rt0_go to bootstrap the execution of goroutines (G) on that thread. So, while the process has been running, the backtrace tells us that this specific OS thread (Thread 1) is just starting its life and is being added to the Go runtime's pool of worker threads. The race is happening when a brand new OS thread is being initialized and is creating its first goroutine, while the GC is concurrently active."

(rr) # gemini: "This is still a form of "initialization," but it's thread-level, not process-level. This can happen long after your main() function has started."

(rr) # gemini: "How to Be Certain About Timing

(rr) # gemini: "You are not certain how close to the crash we are. rr can tell you this with precision. The execution trace is a linear sequence of events. Let's check the time difference between the corruption and the crash."

(rr) when

Completed event: 13902
(rr) list

371					print("s.state = ", state, "\n")
372					throw("non in-use span found with specials bit set")
373				}
374				// Check that this span was swept (it may be cached or uncached).
375				if !useCheckmark && !(s.sweepgen == sg || s.sweepgen == sg+3) {
376					// sweepgen was updated (+2) during non-che--Type <RET> for more, q to quit, c to continue without paging-- 

ckmark GC pass
377					print("sweep ", s.sweepgen, " ", sg, "\n")
378					throw("gc: unswept span")
379				}
380	
(rr) bt

#0  0x0000000000435468 in void __tsan::MemoryAccessRangeT<false>(__tsan::ThreadState*, unsigned long, unsigned long, unsigned long) ()
#1  0x0000000002200078 in __tsan::ctx_placeholder ()
#2  0x000000c2003598a8 in ?? ()
#3  0x000000c2003aca80 in ?? ()
#4  0x00000000004be8ba in racecall ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/race_amd64.s:447
#5  0x00000000004bb14a in runtime.systemstack ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:514
#6  0x00007ffe1ffb4c38 in ?? ()
#7  0x00000000004c03bf in runtime.newproc (fn=0x4bafcf <runtime.rt0_go+303>)
    at <autogenerated>:1
#8  0x00000000004bb045 in runtime.mstart ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:395
#9  0x00000000004bafcf in runtime.rt0_go ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:358
#10 0x0000000000000004 in ?? ()
#11 0x00007ffe1ffb4c88 in ?? ()
#12 0x0000000000000000 in ?? ()
(rr) i thr

  Id   Target Id                   Frame 
* 1    Thread 7129.7129 (rpc.test) 0x0000000000435468 in void __tsan::MemoryAccessRangeT<false>(__tsan::ThreadState*, unsigned long, unsigned long, unsigned long) ()
  2    Thread 7129.7131 (rpc.test) 0x0000000000435468 in void __tsan::MemoryAccessRangeT<false>(__tsan::ThreadState*, unsigned long, unsigned long, unsigned long) ()
  3    Thread 7129.7130 (rpc.test) 0x000000000048aa27 in runtime.retake (
    now=4084794783364, ~r0=<optimized out>)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:6258
  4    Thread 7129.7132 (rpc.test) runtime.futex ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/sys_linux_amd64.s:558
  5    Thread 7129.7133 (rpc.test) runtime.futex ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/sys_linux_amd64.s:558
  6    Thread 7129.7134 (rpc.test) runtime.futex ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/sys_linux_amd64.s:558
  7    Thread 7129.7135 (rpc.test) 0x0000000070000002 in syscall_traced ()
(rr) i goro

  1 waiting  runtime.gopark
  2 waiting  runtime.gopark
  3 runnable runtime.goschedIfBusy
  4 waiting  runtime.gopark
  5 waiting  runtime.gopark
  6 waiting  runtime.gopark
  7 waiting  runtime.gopark
* 8 syscall  runtime.notetsleepg
  9 waiting  runtime.gopark
  10 waiting  runtime.gopark
  11 waiting  runtime.gopark
* 12 running  runtime.systemstack_switch
  20 runnable github.com/klauspost/compress/zstd.(*bestFastEncoder).Reset
  15 waiting  runtime.gopark
  16 waiting  runtime.gopark
  18 waiting  runtime.gopark
  19 waiting  runtime.gopark
* 21 running  runtime.systemstack_switch
(rr) bt

#0  0x0000000000435468 in void __tsan::MemoryAccessRangeT<false>(__tsan::ThreadState*, unsigned long, unsigned long, unsigned long) ()
#1  0x0000000002200078 in __tsan::ctx_placeholder ()
#2  0x000000c2003598a8 in ?? ()
#3  0x000000c2003aca80 in ?? ()
#4  0x00000000004be8ba in racecall ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/race_amd64.s:447
#5  0x00000000004bb14a in runtime.systemstack ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:514
#6  0x00007ffe1ffb4c38 in ?? ()
#7  0x00000000004c03bf in runtime.newproc (fn=0x4bafcf <runtime.rt0_go+303>)
    at <autogenerated>:1
#8  0x00000000004bb045 in runtime.mstart ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:395
#9  0x00000000004bafcf in runtime.rt0_go ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:358
#10 0x0000000000000004 in ?? ()
#11 0x00007ffe1ffb4c88 in ?? ()
#12 0x0000000000000000 in ?? ()
(rr) # Continue forward to the crash:

(rr) c

Continuing.

Thread 1 hit Hardware watchpoint 2: *0x218e08000000

Old value = 0
New value = 398847
0x000000000043546a in void __tsan::MemoryAccessRangeT<false>(__tsan::ThreadState*, unsigned long, unsigned long, unsigned long) ()

(rr) c

Continuing.
example.go:300 server-side: WillHangUntilCancel() is running
example.go:304 net.rpc API: our net.Conn has 
local = '127.0.0.1:34869'
remote = '127.0.0.1:51946'

cli_test.go:984 [goID 10] 2025-06-08 23:10:21.227595976 +0000 UTC cli_test 040: we got past test040callStarted

cli_test.go:988 [goID 10] 2025-06-08 23:10:21.227787850 +0000 UTC past cancelFunc()

cli_test.go:977 [goID 18] 2025-06-08 23:10:21.231561912 +0000 UTC client.Call() returned with cliErr = 'cancellation request sent'

cli_test.go:991 [goID 10] 2025-06-08 23:10:21.231755899 +0000 UTC past cliErrIsSet channel; cliErr40 = 'cancellation request sent'

cli_test.go:998 [goID 10] 2025-06-08 23:10:21.231934678 +0000 UTC about to verify that server side context was cancelled.
example.go:310 MustBeCancelled.WillHangUntilCancel(): ctx.Done() was closed!

cli_test.go:1000 [goID 10] 2025-06-08 23:10:21.237329200 +0000 UTC server side saw the cancellation request: confirmed.
example.go:324 server-side: MessageAPI_HangUntilCancel() is running
example.go:326: Message API: our net.Conn has local = '127.0.0.1:34869'; remote = '127.0.0.1:51946'

cli_test.go:1020 [goID 10] 2025-06-08 23:10:21.243213831 +0000 UTC cli_test 041: we got past test041callStarted

cli_test.go:1024 [goID 10] 2025-06-08 23:10:21.243403981 +0000 UTC past cancelFunc()

cli_test.go:1014 [goID 24] 2025-06-08 23:10:21.243850237 +0000 UTC client.Call() returned with cliErr = 'cancellation request sent'

cli_test.go:1027 [goID 10] 2025-06-08 23:10:21.244036540 +0000 UTC past cliErrIsSet channel; cliErr = 'cancellation request sent'

cli_test.go:1038 [goID 10] 2025-06-08 23:10:21.244213305 +0000 UTC about to verify that server side context was cancelled.
example.go: MustBeCancelled.MessageAPI_HangUntilCancel(): ctx.Done() was closed!

0 assertions thus far

--- PASS: Test040_remote_cancel_by_context (2.91s)
=== RUN   Test045_upload

cli.go:211 [goID 29] 2025-06-08 23:10:21.313925974 +0000 UTC connected to server 127.0.0.1:38633
[New Thread 7129.7136]
[New Thread 7129.7137]

Thread 2 received signal SIGSEGV, Segmentation fault.
[Switching to Thread 7129.7131]
runtime.markrootSpans (gcw=0xc200055750, shard=<optimized out>)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/mgcmark.go:370
370				if state := s.state.get(); state != mSpanInUse {

(rr) when

Completed event: 15881
(rr) reverse-continue

Continuing.

Thread 2 received signal SIGSEGV, Segmentation fault.
runtime.markrootSpans (gcw=0xc200055750, shard=<optimized out>)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/mgcmark.go:370
370				if state := s.state.get(); state != mSpanInUse {

(rr) rs

370				if state := s.state.get(); state != mSpanInUse {

(rr) rs

runtime.(*mSpanStateBox).get (b=<optimized out>)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/mheap.go:392
392		return mSpanState(b.s.Load())
(rr) rs

rs
366				s := ha.spans[arenaPage+uint(i)*8+j]

(rr) 
357				if specials&(1<<j) == 0 {

(rr) rs

356			for j := uint(0); j < 8; j++ {

(rr) reverse-continue

Continuing.
[Thread 7129.7136 exited]
[Thread 7129.7137 exited]
[Switching to Thread 7129.7129]

Thread 1 hit Hardware watchpoint 2: *0x218e08000000

Old value = 398847
New value = 0
0x0000000000435468 in void __tsan::MemoryAccessRangeT<false>(__tsan::ThreadState*, unsigned long, unsigned long, unsigned long) ()

(rr) i thr

  Id   Target Id                   Frame 
* 1    Thread 7129.7129 (rpc.test) 0x0000000000435468 in void __tsan::MemoryAccessRangeT<false>(__tsan::ThreadState*, unsigned long, unsigned long, unsigned long) ()
  2    Thread 7129.7131 (rpc.test) 0x0000000000435468 in void __tsan::MemoryAccessRangeT<false>(__tsan::ThreadState*, unsigned long, unsigned long, unsigned long) ()
  3    Thread 7129.7130 (rpc.test) 0x000000000048aa27 in runtime.retake (
    now=4084794783364, ~r0=<optimized out>)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/proc.go:6258
  4    Thread 7129.7132 (rpc.test) runtime.futex ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/sys_linux_amd64.s:558
  5    Thread 7129.7133 (rpc.test) runtime.futex ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/sys_linux_amd64.s:558
  6    Thread 7129.7134 (rpc.test) runtime.futex ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/sys_linux_amd64.s:558
  7    Thread 7129.7135 (rpc.test) 0x0000000070000002 in syscall_traced ()
(rr) bt

#0  0x0000000000435468 in void __tsan::MemoryAccessRangeT<false>(__tsan::ThreadState*, unsigned long, unsigned long, unsigned long) ()
#1  0x0000000002200078 in __tsan::ctx_placeholder ()
#2  0x000000c2003598a8 in ?? ()
#3  0x000000c2003aca80 in ?? ()
#4  0x00000000004be8ba in racecall ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/race_amd64.s:447
#5  0x00000000004bb14a in runtime.systemstack ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:514
#6  0x00007ffe1ffb4c38 in ?? ()
#7  0x00000000004c03bf in runtime.newproc (fn=0x4bafcf <runtime.rt0_go+303>)
    at <autogenerated>:1
#8  0x00000000004bb045 in runtime.mstart ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:395
#9  0x00000000004bafcf in runtime.rt0_go ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:358
#10 0x0000000000000004 in ?? ()
#11 0x00007ffe1ffb4c88 in ?? ()
#12 0x0000000000000000 in ?? ()
(rr) i goro

  1 waiting  runtime.gopark
  2 waiting  runtime.gopark
  3 runnable runtime.goschedIfBusy
  4 waiting  runtime.gopark
  5 waiting  runtime.gopark
  6 waiting  runtime.gopark
  7 waiting  runtime.gopark
* 8 syscall  runtime.notetsleepg
  9 waiting  runtime.gopark
  10 waiting  runtime.gopark
  11 waiting  runtime.gopark
* 12 running  runtime.systemstack_switch
  20 runnable github.com/klauspost/compress/zstd.(*bestFastEncoder).Reset
  15 waiting  runtime.gopark
  16 waiting  runtime.gopark
  18 waiting  runtime.gopark
  19 waiting  runtime.gopark
* 21 running  runtime.systemstack_switch
(rr) frame 7

Download failed: Invalid argument.  Continuing without source file ./<autogenerated>.
#7  0x00000000004c03bf in runtime.newproc (fn=0x4bafcf <runtime.rt0_go+303>)
    at <autogenerated>:1
warning: 1	<autogenerated>: No such file or directory
(rr) info args

fn = 0x4bafcf <runtime.rt0_go+303>
(rr) list

1	in <autogenerated>
(rr) thread 7131

Unknown thread 7131.
(rr) thread 2

[Switching to thread 2 (Thread 7129.7131)]
#0  0x0000000000435468 in void __tsan::MemoryAccessRangeT<false>(__tsan::ThreadState*, unsigned long, unsigned long, unsigned long) ()
(rr) bt

#0  0x0000000000435468 in void __tsan::MemoryAccessRangeT<false>(__tsan::ThreadState*, unsigned long, unsigned long, unsigned long) ()
#1  0x0000000000410048 in __sanitizer::DD::MutexBeforeLock(__sanitizer::DDCallback*, __sanitizer::DDMutex*, bool) ()
#2  0x000000c200365408 in ?? ()
#3  0x000000c2002ffa40 in ?? ()
#4  0x00000000004be8ba in racecall ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/race_amd64.s:447
#5  0x0000000000000000 in ?? ()
(rr) up

#1  0x0000000000410048 in __sanitizer::DD::MutexBeforeLock(__sanitizer::DDCallback*, __sanitizer::DDMutex*, bool) ()
(rr) info args

No symbol table info available.
(rr) up

#2  0x000000c200365408 in ?? ()
(rr) info frame

Stack level 2, frame at 0x7ba951eb3e38:
 rip = 0xc200365408; saved rip = 0xc2002ffa40
 called by frame at 0x7ba951eb3e40, caller of frame at 0x7ba951eb3e30
 Arglist at 0x7ba951eb3e28, args: 
 Locals at 0x7ba951eb3e28, Previous frame's sp is 0x7ba951eb3e38
 Saved registers:
  rip at 0x7ba951eb3e30
(rr) disassemble  0x000000c200365408

No function contains specified address.
(rr) # gemini explained that the optimized c++ tsan and Go trampolines are likely preventing gdb from interpretting the code, it suggested:

(rr) # In Thread 2, examine 10 instructions starting at the problematic address

(rr) x/10i 0xc200365408

=> 0xc200365408:	sub    %dl,0x36(%rbp)
   0xc20036540b:	add    %al,%dl
   0xc20036540d:	add    %al,(%rax)
   0xc20036540f:	add    %bh,-0x39(%rdi)
   0xc200365412:	int    $0x0
   0xc200365414:	add    %al,(%rax)
   0xc200365416:	add    %al,(%rax)
   0xc200365418:	add    (%rax),%al
   0xc20036541a:	add    %al,(%rax)
   0xc20036541c:	add    %al,(%rax)
(rr) disassemble  0xc200365408

No function contains specified address.
(rr) x/100i 0xc200365408

=> 0xc200365408:	sub    %dl,0x36(%rbp)
   0xc20036540b:	add    %al,%dl
   0xc20036540d:	add    %al,(%rax)
   0xc20036540f:	add    %bh,-0x39(%rdi)
   0xc200365412:	int    $0x0
   0xc200365414:	add    %al,(%rax)
   0xc200365416:	add    %al,(%rax)
   0xc200365418:	add    (%rax),%al
   0xc20036541a:	add    %al,(%rax)
   0xc20036541c:	add    %al,(%rax)
   0xc20036541e:	add    %al,(%rax)
   0xc200365420:	add    (%rax),%eax
   0xc200365422:	add    %al,(%rax)
   0xc200365424:	add    %al,(%rax)
   0xc200365426:	add    %al,(%rax)
   0xc200365428:	add    %al,(%rax)
   0xc20036542a:	add    %al,(%rax)
   0xc20036542c:	add    %al,(%rax)
   0xc20036542e:	add    %al,(%rax)
   0xc200365430:	add    %al,(%rax)
   0xc200365432:	add    %al,(%rax)
   0xc200365434:	add    %al,(%rax)
   0xc200365436:	add    %al,(%rax)
   0xc200365438:	add    %al,(%rax)
   0xc20036543a:	add    (%rax),%al
--Type <RET> for more, q to quit, c to continue without paging--c

   0xc20036543c:	add    %al,(%rax)
   0xc20036543e:	add    %al,(%rax)
   0xc200365440:	add    %al,(%rax)
   0xc200365442:	addb   $0x0,(%rax)
   0xc200365445:	add    %al,(%rax)
   0xc200365447:	add    %al,(%rcx)
   0xc200365449:	add    %al,(%rax)
   0xc20036544b:	add    %eax,(%rax)
   0xc20036544d:	add    %al,(%rax)
   0xc20036544f:	add    %al,(%rax)
   0xc200365451:	add    %al,(%rax)
   0xc200365453:	add    %al,(%rax)
   0xc200365455:	add    %al,(%rax)
   0xc200365457:	add    %bl,0x25(%rdx)
   0xc20036545a:	add    %r8b,(%r8)
   0xc20036545d:	add    %al,(%rax)
   0xc20036545f:	add    %bl,%al
   0xc200365461:	pushf
   0xc200365462:	cmp    $0x1,%eax
   0xc200365467:	add    %al,(%rax)
   0xc200365469:	add    %al,(%rdx)
   0xc20036546b:	add    %al,(%rax)
   0xc20036546d:	add    %al,(%rax)
   0xc20036546f:	add    %al,%al
   0xc200365471:	ss (bad)
   0xc200365473:	add    %al,%dl
   0xc200365475:	add    %al,(%rax)
   0xc200365477:	add    %al,(%rax)
   0xc200365479:	add    %al,(%rax)
   0xc20036547b:	add    %al,(%rax)
   0xc20036547d:	add    %al,(%rax)
   0xc20036547f:	add    %al,0x12167(%rax)
   0xc200365485:	add    %al,(%rax)
   0xc200365487:	add    %ah,-0x3dffc6ac(%rax)
   0xc20036548d:	add    %al,(%rax)
   0xc20036548f:	add    %cl,(%rax)
   0xc200365491:	rex.WX adc %cl,(%rbx)
   0xc200365494:	(bad)
   0xc200365497:	add    %bh,%al
   0xc200365499:	ss (bad)
   0xc20036549b:	add    %al,%dl
   0xc20036549d:	add    %al,(%rax)
   0xc20036549f:	add    %ch,(%rax)
   0xc2003654a1:	(bad)
   0xc2003654a2:	(bad)
   0xc2003654a3:	add    %al,%dl
   0xc2003654a5:	add    %al,(%rax)
   0xc2003654a7:	add    %bl,-0x3dffe0c9(%rcx)
   0xc2003654ad:	add    %al,(%rax)
   0xc2003654af:	add    %al,%al
   0xc2003654b1:	ss (bad)
   0xc2003654b3:	add    %al,%dl
   0xc2003654b5:	add    %al,(%rax)
   0xc2003654b7:	add    %dh,-0x3dffc6ac(%rax)
   0xc2003654bd:	add    %al,(%rax)
   0xc2003654bf:	add    %cl,-0x3dffc6ac(%rax)
   0xc2003654c5:	add    %al,(%rax)
   0xc2003654c7:	add    %dh,0x54(%rax)
   0xc2003654ca:	cmp    %eax,(%rax)
   0xc2003654cc:	ret    $0x0
   0xc2003654cf:	add    %al,(%rax)
   0xc2003654d1:	movabs 0x5420000000c50b09,%al
   0xc2003654da:	cmp    %eax,(%rax)
   0xc2003654dc:	ret    $0x0
   0xc2003654df:	add    %bl,0x54(%rax)
   0xc2003654e2:	cmp    %eax,(%rax)
   0xc2003654e4:	ret    $0x0
   0xc2003654e7:	add    %al,(%rdx)
   0xc2003654e9:	add    %al,(%rax)
   0xc2003654eb:	add    %al,(%rax)
   0xc2003654ed:	add    %al,(%rax)
   0xc2003654ef:	add    %al,(%rbx)
   0xc2003654f1:	add    %al,(%rax)
   0xc2003654f3:	add    %al,(%rax)
   0xc2003654f5:	add    %al,(%rax)
(rr) bt

#0  0x0000000000435468 in void __tsan::MemoryAccessRangeT<false>(__tsan::ThreadState*, unsigned long, unsigned long, unsigned long) ()
#1  0x0000000000410048 in __sanitizer::DD::MutexBeforeLock(__sanitizer::DDCallback*, __sanitizer::DDMutex*, bool) ()
#2  0x000000c200365408 in ?? ()
#3  0x000000c2002ffa40 in ?? ()
#4  0x00000000004be8ba in racecall ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/race_amd64.s:447
#5  0x0000000000000000 in ?? ()
(rr) frame 4

#4  0x00000000004be8ba in racecall ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/race_amd64.s:447
447		CALL	AX
(rr) info reg rsp

rsp            0x7ba951eb3e40      0x7ba951eb3e40
(rr) # gemini: # Watch the 64 bytes of memory starting at the stack pointer address.

(rr) # gemini: "This region almost certainly contains the return address that was corrupted."

(rr) watch -l * (char[64] *) 0x7ba951eb3e40

A syntax error in expression, near `[64] *) 0x7ba951eb3e40'.
(rr) watch *(unsigned long long *)0x7ba951eb3e40

No symbol "unsigned" in current context.
(rr) watch *(long long *)0x7ba951eb3e40

A syntax error in expression, near `long *)0x7ba951eb3e40'.
(rr) watch *0x7ba951eb3e40

Hardware watchpoint 3: *0x7ba951eb3e40
(rr)  awatch -l 0x7ba951eb3e40, 64

Cannot watch constant value `0x7ba951eb3e40, 64'.
(rr)  awatch -l *0x7ba951eb3e40, 64

Attempt to take address of value not located in memory.
(rr) # lol. gemini is such a drama queen: "The Final Step: Find the Culprit. You have now completed the setup. The trap is laid. Hardware watchpoint 3 is active and waiting. The one and only thing left to do is to run the program backward and see what springs the trap." 

(rr) reverse-continue

Continuing.
[Switching to Thread 7129.7129]

Thread 1 hit Hardware watchpoint 2: *0x218e08000000

Old value = 0
New value = 2042251376
0x000000000040482b in __sanitizer::internal_mmap(void*, unsigned long, int, int, int, unsigned long long) ()

(rr) bt

#0  0x000000000040482b in __sanitizer::internal_mmap(void*, unsigned long, int, int, int, unsigned long long) ()
#1  0x0000218e05219000 in ?? ()
#2  0x00000000043ff000 in ?? ()
#3  0x0000218e05219000 in ?? ()
#4  0x0000000000404880 in __sanitizer::MmapNamed(void*, unsigned long, int, int, char const*)
    ()
#5  0x0000000001922e80 in ?? ()
#6  0x000040321ffb4b70 in ?? ()
#7  0x0000000000465c71 in runtime.(*mspan).ensureSwept (s=<optimized out>)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/mgcsweep.go:477
#8  0x000000000046a73e in runtime.unlockWithRank (l=<optimized out>)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/lockrank_off.go:35
#9  runtime.unlock (l=<optimized out>)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/lock_spinbit.go:253
#10 runtime.addspecial (p=0x218e09618000, s=0x0, force=189, ~r0=<optimized out>)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/mheap.go:1895
#11 0x000000000046b187 in runtime.setprofilebucket (p=0x1989a00 <runtime.m0>, b=0xc2003aca80)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/mheap.go:2311
#12 0x000000c2003597d0 in ?? ()
#13 0x0000000001989a00 in ?? ()
#14 0x000000c2003aca80 in ?? ()
#15 0x00000000004be8ba in racecall ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/race_amd64.s:447
#16 0x00000000004bb14a in runtime.systemstack ()
--Type <RET> for more, q to quit, c to continue without paging--c

    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:514
#17 0x00007ffe1ffb4c38 in ?? ()
#18 0x00000000004c03bf in runtime.newproc (fn=0x4bafcf <runtime.rt0_go+303>)
    at <autogenerated>:1
#19 0x00000000004bb045 in runtime.mstart ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:395
#20 0x00000000004bafcf in runtime.rt0_go ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:358
#21 0x0000000000000004 in ?? ()
#22 0x00007ffe1ffb4c88 in ?? ()
#23 0x0000000000000000 in ?? ()
(rr) when

Completed event: 13867
(rr) reverse-continue

Continuing.

Thread 1 hit Hardware watchpoint 2: *0x218e08000000

Old value = 2042251376
New value = 0
0x0000000000468f6c in runtime.(*mheap).setSpans (h=0x1993760 <runtime.mheap_>, 
    base=<optimized out>, npage=4353, s=0x2bee79ba4870)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/mheap.go:1011
1011			ha.spans[i] = s

(rr) bt

#0  0x0000000000468f6c in runtime.(*mheap).setSpans (h=0x1993760 <runtime.mheap_>, 
    base=<optimized out>, npage=4353, s=0x2bee79ba4870)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/mheap.go:1011
#1  0x0000000000469a3a in runtime.(*mheap).initSpan (h=0x1993760 <runtime.mheap_>, 
    s=0x2bee79ba4870, typ=0 '\000', spanclass=0 '\000', base=854741532672, npages=4353)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/mheap.go:1454
#2  0x0000000000469657 in runtime.(*mheap).allocSpan (h=0x1993760 <runtime.mheap_>, 
    npages=4353, typ=0 '\000', spanclass=0 '\000', s=<optimized out>)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/mheap.go:1346
#3  0x0000000000468e9c in runtime.(*mheap).alloc.func1 ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/mheap.go:970
#4  0x00000000004bb14a in runtime.systemstack ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:514
#5  0x00007ffe1ffb4c38 in ?? ()
#6  0x00000000004c03bf in runtime.newproc (fn=0x4bafcf <runtime.rt0_go+303>)
    at <autogenerated>:1
#7  0x00000000004bb045 in runtime.mstart ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:395
#8  0x00000000004bafcf in runtime.rt0_go ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/asm_amd64.s:358
#9  0x0000000000000004 in ?? ()
#10 0x00007ffe1ffb4c88 in ?? ()
#11 0x0000000000000000 in ?? ()
(rr) when

Completed event: 13867
(rr) frame 0

#0  0x0000000000468f6c in runtime.(*mheap).setSpans (h=0x1993760 <runtime.mheap_>, 
    base=<optimized out>, npage=4353, s=0x2bee79ba4870)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/mheap.go:1011
1011			ha.spans[i] = s
(rr) info args

h = 0x1993760 <runtime.mheap_>
base = <optimized out>
npage = 4353
s = 0x2bee79ba4870
(rr) p h.lock

$8 = {lockRankStruct = {<No data fields>}, key = 0}
(rr) ## gemini: "The output key = 0 is unambiguous. It proves that at the moment setSpans is executing this write, the main heap lock (h.lock) is not held by this thread."

(rr) x/gx &h.lock

0x1993760 <runtime.mheap_>:	0x0000000000000000
(rr) thread 2

[Switching to thread 2 (Thread 7129.7131)]
#0  0x0000000000435468 in void __tsan::MemoryAccessRangeT<false>(__tsan::ThreadState*, unsigned long, unsigned long, unsigned long) ()
(rr) bt

#0  0x0000000000435468 in void __tsan::MemoryAccessRangeT<false>(__tsan::ThreadState*, unsigned long, unsigned long, unsigned long) ()
#1  0x0000000000410048 in __sanitizer::DD::MutexBeforeLock(__sanitizer::DDCallback*, __sanitizer::DDMutex*, bool) ()
#2  0x000000c200365408 in ?? ()
#3  0x000000c2002ffa40 in ?? ()
#4  0x00000000004be8ba in racecall ()
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/race_amd64.s:447
#5  0x0000000000000000 in ?? ()
(rr) frame 1

#1  0x0000000000410048 in __sanitizer::DD::MutexBeforeLock(__sanitizer::DDCallback*, __sanitizer::DDMutex*, bool) ()
(rr) info reg rdi

rdi            0x2d6d11601500      49946466194688
(rr) thread 1

[Switching to thread 1 (Thread 7129.7129)]
#0  0x0000000000468f6c in runtime.(*mheap).setSpans (h=0x1993760 <runtime.mheap_>, 
    base=<optimized out>, npage=4353, s=0x2bee79ba4870)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/mheap.go:1011
1011			ha.spans[i] = s
(rr) up

#1  0x0000000000469a3a in runtime.(*mheap).initSpan (h=0x1993760 <runtime.mheap_>, 
    s=0x2bee79ba4870, typ=0 '\000', spanclass=0 '\000', base=854741532672, npages=4353)
    at /mnt/oldrog/usr/local/go1.24.3/src/runtime/mheap.go:1454
1454		h.setSpans(s.base(), npages, s)
(rr) list

1449		// This is safe to call without the lock held because the slots
1450		// related to this span will only ever be read or modified by
1451		// this thread until pointers into the span are published (and
1452		// we execute a publication barrier at the end of this function
1453		// before that happens) or pageInUse is updated.
1454		h.setSpans(s.base(), npages, s)
1455	
1456		if !typ.manual() {
1457			// Mark in-use span in arena page bitmap.
1458			//
(rr) list 30

25		// maxPhysPageSize is the maximum page size the runtime supports.
26		maxPhysPageSize = 512 << 10
27	
28		// maxPhysHugePageSize sets an upper-bound on the maximum huge page size
29		// that the runtime supports.
30		maxPhysHugePageSize = pallocChunkBytes
31	
32		// pagesPerReclaimerChunk indicates how many pages to scan from the
33		// pageInUse bitmap at a time. Used by the page reclaimer.
34		//
(rr) 
~~~

