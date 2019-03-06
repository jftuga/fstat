
PROG = fstat
DEPS = render_number.go

$(PROG) : $(PROG).go
	go build -ldflags "-s -w" $(PROG).go $(DEPS)

clean:
	rm -f $(PROG) *~ .??*~
