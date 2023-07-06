.POSIX:

env ?= "prod"

default: install

prepare:
	ansible-playbook \
 		--inventory inventory/hosts.yml \
 		playbooks/cluster-prepare.yml

install:
	ansible-playbook \
 		--inventory inventory/hosts.yml \
 		playbooks/cluster-installation.yml

nuke:
	ansible-playbook \
 		--inventory inventory/hosts.yml \
 		playbooks/cluster-nuke.yml

reboot:
	ansible-playbook \
 		--inventory inventory/hosts.yml \
 		playbooks/cluster-reboot.yml

upgrade:
	ansible-playbook \
 		--inventory inventory/hosts.yml \
 		playbooks/cluster-upgrade.yml

console:
	ansible-console \
		--inventory inventory/hosts.yml

list:
	ansible all -i inventory/hosts.yml --list-hosts

ping:
	ansible all -i inventory/hosts.yml --one-line -m 'ping'
