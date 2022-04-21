.PHONY: build
build:
	go build -o cvo main.go

.PHONY: install-plugin
install-plugin: build
	mkdir -p ${HOME}/bin
	cp cvo ${HOME}/bin/kubectl-cvo
