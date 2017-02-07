DIRNAME=$(notdir $(shell pwd))
PROJECTNAME=$(DIRNAME)

.PHONY: all clean

all:
	go build .

test:
	go test -v -coverprofile=$(PROJECTNAME)-profile.out ./PersistentCounter && \
		go tool cover -html=$(PROJECTNAME)-profile.out

clean:
	rm $(PROJECTNAME) & rm $(PROJECTNAME)-profile.out
