.PHONY: all s

all:
	go test -c -o hello.test
	rr record -h ./hello.test

s:
	GOTRACEBACK=all GOEXPERIMENT=synctest go test -c -o hello.test
	GOTRACEBACK=all GOEXPERIMENT=synctest rr record -h ./hello.test

clean:
	rm -rf ../rr
	mkdir ../rr
