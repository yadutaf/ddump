build:
	go build ./cmd/ddump-worker
	go build ./cmd/ddump

clean:
	rm -f ddump-worker
	rm -f ddump
