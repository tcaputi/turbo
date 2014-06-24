CC=go
DB=gdb

all: clean build test run

clean:
	go clean
	rm test/test
build:
	go install
run: build
	cd ./test && go build && cd ..&& clear && ./test/test
init:
	go get && cd ./test && go get
commit:
	git add -A && git commit -m '$(filter-out $@,$(MAKECMDGOALS))' && git pull && git push
test:
	go test