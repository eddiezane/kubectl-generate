.PHONY: clean build run install uninstall

clean:
	rm -rf _out

build: clean
	go build -o _out/kubectl-generate cmd/kubectl-generate.go

run: build
	PATH=$(PWD)/_out:$(PATH) kubectl generate deployment

install: build
	mv _out/kubectl-generate /usr/local/bin/

uninstall:
	rm /usr/local/bin/kubectl-generate
