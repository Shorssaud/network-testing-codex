# Create a makefile that runs all the tests in tests/

SRC = $(wildcard tests/*.go)

all:
	go test $(SRC)

verbose:
	go test -v $(SRC)