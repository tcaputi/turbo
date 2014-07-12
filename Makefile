CC=go
DB=gdb

all: clean build test run

clean:
	go clean
	rm test/test
build:
	go install
live: build
	cd ./js/test && clear && go run ./main.go
init:
	go get && cd ./test && go get
commit:
	git add -A && git commit -m '$(filter-out $@,$(MAKECMDGOALS))' && git pull && git push
test:
	go test