package restore

import (
	. "backuputil/common"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"log"
	"os/exec"
)

func Keycloak(client *kubernetes.Clientset) {
	const namespace = "keycloak"
	const keycloakPod = "keycloak-0"
	const realmDumpPath = "/keycloak-keycloak.json"
	const tmpDir = "./tmp"

	// Input: snapshot id => download realm dump
	snapshotMap, err := CreateResticSnapshotMap(namespace)
	if err != nil {
		log.Fatal(err)
	}
	snapshot := ResticSnapshotSelectionPrompt(snapshotMap, realmDumpPath)
	if err := RestoreResticSnapshot(namespace, realmDumpPath, snapshot, tmpDir); err != nil {
		log.Fatal(err)
	}

	// 1. Copy realm dump into container
	fmt.Println("Copying realm dump into container...")
	if err := CopyFileIntoContainer(namespace, keycloakPod, tmpDir+realmDumpPath, "/opt/bitnami/keycloak/data/finance-realm.json"); err != nil {
		fmt.Println("Error copying realm dump into container:")
		log.Fatal(err)
	}

	// 2. Import realm using kc.sh import
	fmt.Println("Importing realm...")
	if err := KeycloakImportRealm(namespace, keycloakPod, "/opt/bitnami/keycloak/data/finance-realm.json"); err != nil {
		fmt.Println("Error importing realm:")
		log.Fatal(err)
	}
}

func CopyFileIntoContainer(namespace string, pod string, localPath string, containerPath string) error {
	return exec.Command("kubectl", "cp", localPath, namespace+"/"+pod+":"+containerPath).Run()
}

func KeycloakImportRealm(namespace string, keycloakPod string, realmDumpPath string) error {
	cmdStr := "kubectl exec -i " + keycloakPod + " -n " + namespace +
		" -- /opt/bitnami/keycloak/bin/kc.sh import --file " + realmDumpPath

	return exec.Command("sh", "-c", cmdStr).Run()
}
