rpc.test built with

~~~
go test -race -c -o rpc.test
~~~

at v1.22.61 of

https://github.com/glycerine/rpc25519


~~~
jaten@rog ~/hello/rrlogs.rpc045 (master) $ go version -m ./rpc.test
./rpc.test: go1.24.3
	path	github.com/glycerine/rpc25519.test
	mod	github.com/glycerine/rpc25519	(devel)	
	build	-buildmode=exe
	build	-compiler=gc
	build	-race=true
	build	CGO_ENABLED=1
	build	CGO_CFLAGS=
	build	CGO_CPPFLAGS=
	build	CGO_CXXFLAGS=
	build	CGO_LDFLAGS=
	build	GOARCH=amd64
	build	GOOS=linux
	build	GOAMD64=v1
jaten@rog ~/hello/rrlogs.rpc045 (master) $ 
~~~
