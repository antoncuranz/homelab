package restore

import (
	. "backuputil/common"
	"bufio"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"log"
	"os"
	"os/exec"
)

const namespace = "immich"
const postgresPod = "immich-postgresql-0"
const sqlDumpPath = "/immich-postgresql.sql.gz"
const dataPath = "/data/immich-data"
const tmpDir = "./tmp"

func RestoreImmich() {

	// TODO: pass client here?
	client, err := InitClient()
	if err != nil {
		log.Fatal(err)
	}

	snapshotMap, err := CreateResticSnapshotMap(namespace)
	if err != nil {
		log.Fatal(err)
	}

	// 0. Input: snapshot ids, sql backup path
	dataSnapshot := ResticSnapshotSelectionPrompt(snapshotMap, dataPath)
	psqlSnapshot := ResticSnapshotSelectionPrompt(snapshotMap, sqlDumpPath)
	if err := RestoreResticSnapshot(namespace, sqlDumpPath, psqlSnapshot, tmpDir); err != nil {
		log.Fatal(err)
	}

	// 1. Scale down Deployments and Statefulsets
	fmt.Println("Scaling down Deployments and StatefulSets...")
	if err := ScaleDownDeploymentsInNamespace(client, namespace); err != nil {
		log.Fatal(err)
	}
	if err := ScaleDownStatefulSetsInNamespace(client, namespace); err != nil {
		log.Fatal(err)
	}
	if err := WaitUntilPodsAreDeleted(client, namespace); err != nil {
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
	fmt.Println("IS IT DONE????") // TODO: use survey confirm or wait for pod to be ready
	bufio.NewScanner(os.Stdin).Scan()

	cmdStr := "gunzip < " + tmpDir + sqlDumpPath + " | kubectl exec -i " + postgresPod + " -n " + namespace +
		" -- sh -c 'PGPASSWORD=\"$POSTGRES_POSTGRES_PASSWORD\" psql -U postgres -d immich'"
	cmd := exec.Command("sh", "-c", cmdStr)
	if _, err = cmd.Output(); err != nil {
		log.Fatal(err)
	}

	// 5. Full Argo Sync
	fmt.Println("Syncing ArgoCD immich Application...")
	if err := ArgoSyncApplication("immich"); err != nil {
		log.Fatal(err)
	}
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
