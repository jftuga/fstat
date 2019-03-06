
PROG = fstat

$(PROG): $(wildcard *.go)
	go build -ldflags "-s -w" .

clean:
	rm -f $(PROG) *~ .??*~
