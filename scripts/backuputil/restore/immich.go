package restore

import (
	. "backuputil/common"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"log"
	"os/exec"
)

func Immich(client *kubernetes.Clientset) {
	// IMPORTANT: new postgresPassword must match old deployment's postgresPassword!

	const namespace = "immich"
	const postgresPod = "immich-postgresql-0"
	const sqlDumpPath = "/immich-postgresql.sql.gz"
	const dataPath = "/data/immich-data"
	const tmpDir = "./tmp"

	// Input: snapshot ids => download postgres dump
	snapshotMap, err := CreateResticSnapshotMap(namespace)
	if err != nil {
		log.Fatal(err)
	}
	dataSnapshot := ResticSnapshotSelectionPrompt(snapshotMap, dataPath)
	psqlSnapshot := ResticSnapshotSelectionPrompt(snapshotMap, sqlDumpPath)
	if err := RestoreResticSnapshot(namespace, sqlDumpPath, psqlSnapshot, tmpDir); err != nil {
		log.Fatal(err)
	}

	// 1. Scale down Deployments and Statefulsets
	fmt.Println("Scaling down Deployments and StatefulSets...")
	if err := ScaleDownDeploymentsAndStatefulSets(client, namespace); err != nil {
		log.Fatal(err)
	}

	// 2. Delete PVCs
	fmt.Println("Deleting PVCs...")
	if err := DeletePvcsInNamespace(client, namespace); err != nil {
		log.Fatal(err)
	}
	if err := WaitUntilPvcsAreDeleted(client, namespace); err != nil {
		log.Fatal(err)
	}

	// 3. Create and restore NFS PVC using k8up Restore CRD
	fmt.Println("Restoring immich-data PVC...")
	if err := RestorePvc(client, namespace, dataSnapshot); err != nil {
		log.Fatal(err)
	}

	// 4. (Argo-)Sync Postgres and run psql Restore
	fmt.Println("Syncing ArgoCD immich postgres STS...")
	if err := ArgoSyncResource("immich", "apps:StatefulSet:immich-postgresql"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Waiting for " + postgresPod + " to be ready...")
	if err := WaitUntilPodIsReady(client, namespace, postgresPod); err != nil {
		log.Fatal(err)
	}

	if err := RestorePostgresDatabase(namespace, postgresPod, tmpDir+sqlDumpPath); err != nil {
		fmt.Println("Error restoring postgres DB:")
		log.Fatal(err)
	}

	// 5. Full Argo Sync
	fmt.Println("Syncing ArgoCD immich Application...")
	if err := ArgoSyncApplication("immich"); err != nil {
		log.Fatal(err)
	}
}

func RestorePostgresDatabase(namespace string, postgresPod string, sqlDumpPath string) error {
	cmdStr := "gunzip < " + sqlDumpPath + " | kubectl exec -i " + postgresPod + " -n " + namespace +
		" -- sh -c 'PGPASSWORD=\"$POSTGRES_POSTGRES_PASSWORD\" psql -U postgres -d immich'"
	return exec.Command("sh", "-c", cmdStr).Run()
}

func RestorePvc(client *kubernetes.Clientset, namespace string, snapshot string) error {
	pvcName := "immich-data"
	pvcStorage := "50Gi"
	storageClass := "truenas-nfs"
	restoreName := "restore-immich"

	if err := CreatePvc(client, namespace, pvcName, pvcStorage, storageClass); err != nil {
		return err
	}

	if err := CreateRestore(client, namespace, restoreName, pvcName, snapshot); err != nil {
		log.Fatal(err)
	}

	return nil
}
