SRCS = $(wildcard *.go)

all: sag

sag: $(SRCS)
	go build

clean:
	rm sag

.PHONY: all clean
