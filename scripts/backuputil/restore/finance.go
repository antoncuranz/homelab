package restore

import (
	. "backuputil/common"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"log"
	"os/exec"
)

func Finance(client *kubernetes.Clientset) {
	// IMPORTANT: new postgresPassword must match old deployment's postgresPassword!

	const namespace = "finance"
	const postgresPod = "finance-postgresql-0"
	const sqlDumpPath = "/finance-postgresql.sql"
	const tmpDir = "./tmp"

	// Input: snapshot ids => download postgres dump
	snapshotMap, err := CreateResticSnapshotMap(namespace)
	if err != nil {
		log.Fatal(err)
	}
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

	// 4. (Argo-)Sync Postgres and run psql Restore
	fmt.Println("Syncing ArgoCD finance postgres STS...")
	if err := ArgoSyncResource("finance", "apps:StatefulSet:finance-postgresql"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Waiting for " + postgresPod + " to be ready...")
	if err := WaitUntilPodIsReady(client, namespace, postgresPod); err != nil {
		log.Fatal(err)
	}

	if err := RestoreFinancePostgresDatabase(namespace, postgresPod, tmpDir+sqlDumpPath); err != nil {
		fmt.Println("Error restoring postgres DB:")
		log.Fatal(err)
	}

	// 5. Full Argo Sync
	fmt.Println("Syncing ArgoCD finance Application...")
	if err := ArgoSyncApplication("finance"); err != nil {
		log.Fatal(err)
	}
}

func RestoreFinancePostgresDatabase(namespace string, postgresPod string, sqlDumpPath string) error {
	cmdStr := "cat " + sqlDumpPath + " | kubectl exec -i " + postgresPod + " -n " + namespace +
		" -- sh -c 'PGPASSWORD=\"$POSTGRES_PASSWORD\" psql -U postgres -d finance'"
	return exec.Command("sh", "-c", cmdStr).Run()
}
