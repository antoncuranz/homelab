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
├─📁 bootstrap     # GitOps bootstrap with ArgoCD
└─📁 metal         # K3s installation using Ansible 
```

## Changes to [khuedoan/homelab](https://github.com/khuedoan/homelab)'s architecture

- followed instructions for single node operation
- replaced metal/ with ansible/ from [onedr0p's template](https://github.com/onedr0p/flux-cluster-template)
- removed cloudflared, zerotier and terraform cloud dependency
- removed gitea and tekton
- replaced vault with onepassword connect
- changed external-dns labels to target (cname) a single a record
- removed nix-shell stuff
- removed longhorn in favor of local-path-provisioner
- ...
- added my own apps

## Installation

(probably incomplete, but you'll figure it out!)


### 📄 Configuration

Run configure script to setup Ansible inventory and other things:
```sh
make configure
```

### ⛵ Installing k3s with Ansible

📍 Here we will be running a Ansible Playbook to install [k3s](https://k3s.io/) with [this](https://galaxy.ansible.com/xanmanning/k3s) wonderful k3s Ansible galaxy role. After completion, Ansible will drop a `kubeconfig` in `./kubeconfig` for use with interacting with your cluster with `kubectl`.

☢️ If you run into problems, you can run `make -C metal nuke` to destroy the k3s cluster and start over.

1. Ensure you are able to SSH into your nodes from your workstation using a private SSH key **without a passphrase**. This is how Ansible is able to connect to your remote nodes.

   [How to configure SSH key-based authentication](https://www.digitalocean.com/community/tutorials/how-to-configure-ssh-key-based-authentication-on-a-linux-server)

2. Verify Ansible can view your config

   ```sh
   make -C metal list
   ```

3. Verify Ansible can ping your nodes

   ```sh
   make -C metal ping
   ```

4. Run the Ansible prepare playbook

   ```sh
   make -C metal prepare
   ```

5. Install k3s with Ansible

   ```sh
   make -C metal install
   ```

### 🚀 Bootstraping ArgoCD

This will install ArgoCD with a root application, referencing all other services in system/, platform/ and apps/:
```sh
make bootstrap
```

### 🤫 Setting up initial secrets

The following will ask you for Cloudflare and 1Password credentials and create respective kubernetes secrets:
```sh
make external
```
