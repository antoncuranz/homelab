.POSIX:
.PHONY: *
.EXPORT_ALL_VARIABLES:

KUBECONFIG = $(shell pwd)/kubeconfig
KUBE_CONFIG_PATH = $(KUBECONFIG)

default: configure bootstrap

configure:
	./scripts/configure
	git status

bootstrap:
	make -C bootstrap

git-hooks:
	pre-commit install
