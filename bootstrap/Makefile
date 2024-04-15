.POSIX:

default: external namespace argocd root

.PHONY: external
external:
	make -C external

.PHONY: namespace
namespace:
	kubectl create namespace argocd --dry-run=client --output=yaml \
		| kubectl apply -f -

.PHONY: argocd
argocd:
	cd argocd && ./apply.sh

.PHONY: root
root:
	cd root && ./apply.sh
