# Forgejo Actions runner

This runner accepts up to four `runs-on: ubuntu-latest` jobs at a time. Jobs run
in a Gitea Ubuntu 26.04 container, not a full GitHub-hosted Ubuntu VM. The job
image includes Node.js, Docker CLI and Buildx. All jobs have access to the shared
Docker daemon through `/var/run/docker.sock`. The runner stores Actions
caches in a dedicated pod-local volume with a 10 GiB limit. Replacing the
pod clears the cache, images, and Docker data.

The jobs share the Docker sidecar's 6 GiB memory limit. After changing runner
configuration, wait for the ExternalSecret to synchronize, then restart the
runner pod so the daemon reads the updated configuration.

The Docker-in-Docker sidecar is privileged and shares the cluster node's kernel.
Use this runner only for trusted workflows: jobs can control the shared daemon,
inspect or interfere with other jobs and potentially compromise the node. It is
not an isolation boundary for untrusted pull requests. A dedicated VM is
preferable for untrusted jobs. The runner does not mount the node's Docker
socket or a Kubernetes service account.

## Registration

Before merging, create a persistent runner at Forgejo's **Site Administration →
Actions → Runners** (or at a narrower repository/organization scope). Store its
UUID and token in the 1Password item `forgejo-runner`, in the vault used by the
`secret-store` ClusterSecretStore, with fields `uuid` and `token`. The
ExternalSecret creates the runner configuration from those fields; do not
commit the token. Once the secret has synchronized and the pod is ready, check
that the runner is online in Forgejo and run a test workflow. The runner uses
`https://git.cura.nz/`, so DNS and TLS must work from the pod and job containers.

## Git over SSH

Forgejo's built-in SSH server listens on container port 2222. The Envoy gateway
routes TCP port 22 on its existing address to Forgejo, so SSH and HTTPS use the
same DNS name. For external access, forward public TCP port 22 to the gateway's
address on the router. Add your public key under **Settings → SSH / GPG Keys**,
then test with `ssh -T git@git.cura.nz`. The built-in server's host keys are
stored in the persistent Forgejo data volume and should be included in backups.
