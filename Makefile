clean:
	rm -rf _out

build: clean
	go build -o _out/kubectl-generate cmd/kubectl-generate.go

run: build
	PATH=$(PWD)/_out:$(PATH) kubectl generate deployment --schema localschema.yaml
