# Forgejo Actions runner

This runner accepts `runs-on: ubuntu-latest` jobs, one at a time. Jobs run in a
Node.js Debian container, not a full GitHub-hosted Ubuntu VM. Docker builds and
Docker-based actions are not supported by default. Cache is disabled, and
images and Docker data are discarded when the pod is replaced.

The Docker-in-Docker sidecar is privileged and shares the cluster node's kernel.
Use this runner only for trusted workflows; it is not an isolation boundary for
untrusted pull requests. A dedicated VM is preferable for untrusted jobs. The
runner does not mount the node's Docker socket or a Kubernetes service account.

## Registration

Before merging, create a persistent runner at Forgejo's **Site Administration →
Actions → Runners** (or at a narrower repository/organization scope). Store its
UUID and token in the 1Password item `forgejo-runner`, in the vault used by the
`secret-store` ClusterSecretStore, with fields `uuid` and `token`. The
ExternalSecret creates the runner configuration from those fields; do not
commit the token. Once the secret has synchronized and the pod is ready, check
that the runner is online in Forgejo and run a test workflow. The runner uses
`https://git.cura.nz/`, so DNS and TLS must work from the pod and job containers.
