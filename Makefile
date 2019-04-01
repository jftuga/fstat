
ifeq ($(OS),Windows_NT)
	PROG = fstat.exe
	ERASE = del
else
	PROG = fstat
	ERASE = rm -f
endif

$(PROG): $(wildcard *.go)
	go build -ldflags "-s -w" .

clean:
	$(ERASE) $(PROG) *~ .??*~

