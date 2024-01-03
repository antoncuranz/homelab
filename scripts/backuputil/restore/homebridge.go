package restore

import (
	. "backuputil/common"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"log"
)

func Homebridge(client *kubernetes.Clientset, snapshotMap NamespacedSnapshotMap) {
	const namespace = "homebridge"
	const dataPath = "/data/homebridge-homebridge"

	// Input: snapshot ids
	dataSnapshot := ResticSnapshotSelectionPrompt(snapshotMap, dataPath)

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
	fmt.Println("Restoring homebridge PV...")
	if err := RestorePvc(client, namespace, dataSnapshot, "homebridge-homebridge", "250Mi", "local-path"); err != nil {
		log.Fatal(err)
	}

	// 4. Full Argo Sync
	fmt.Println("Please perform an ArgoCD sync of the application once restored.")
}
