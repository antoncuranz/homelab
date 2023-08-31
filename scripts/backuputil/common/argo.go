package common

import "os/exec"

func ArgoSyncApplication(application string) error {
	return exec.Command("argocd", "app", "sync", application).Run()
}

func ArgoSyncResource(application string, resource string) error {
	return exec.Command("argocd", "app", "sync", application, "--resource", resource).Run()
}
