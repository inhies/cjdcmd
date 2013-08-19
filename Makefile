PROGRAM_NAME := cjdcmd
GOCOMPILER := go build
GOFLAGS	+= -ldflags "-X main.Version $(shell git describe --tags --dirty=+)"

.PHONY: all clean

all: $(DEPS) $(PROGRAM_NAME)

$(PROGRAM_NAME): $(wildcard *.go)
	$(GOCOMPILER) $(GOFLAGS)

clean:
	@- $(RM) $(PROGRAM_NAME)
