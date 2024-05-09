# Anton's Homelab

Based on [khuedoan/homelab](https://github.com/khuedoan/homelab) with influence of [onedr0p/flux-cluster-template](https://github.com/onedr0p/flux-cluster-template).

## 📂 Repository structure

The Git repository contains the following directories under `kubernetes` and are ordered below by how Flux will apply them.

```sh
📁 homelab
├─📁 apps          # User facing applications (managed by argo)
├─📁 platform      # Essential components for services (managed by argo)
├─📁 system        # Critical system components (managed by argo)
├─📁 external      # Sets up initial secrets using Terraform
└─📁 bootstrap     # GitOps bootstrap with ArgoCD
```

## Changes to [khuedoan/homelab](https://github.com/khuedoan/homelab)'s architecture

- followed instructions for single node operation
- replaced metal/ with ansible/ from [onedr0p's template](https://github.com/onedr0p/flux-cluster-template)
- removed cloudflared, zerotier and terraform cloud dependencies
- removed gitea and tekton
- replaced vault with onepassword connect
- changed external-dns labels to target (cname) a single a record
- removed nix-shell stuff
- removed longhorn
- ...
- added my own apps

## Installation

(probably incomplete, but you'll figure it out!)

### 📄 Configuration

Run configure script to setup Ansible inventory and other things:
```sh
make configure
```

### 🚀 Bootstraping ArgoCD

This will install ArgoCD with a root application, referencing all other services in system/, platform/ and apps/:
```sh
make bootstrap
```

### 🤫 Setting up initial secrets

The following will ask you for Cloudflare and 1Password credentials and creates respective kubernetes secrets:
```sh
make external
```
