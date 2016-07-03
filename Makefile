all: pgdocserve

pgdocserve: pgdocserve.go
	go build

clean:
	rm -f pgdocserve

.PHONY: clean
