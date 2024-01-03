package restore

import (
	. "backuputil/common"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"log"
)

func Changedetection(client *kubernetes.Clientset, snapshotMap NamespacedSnapshotMap) {
	const namespace = "changedetection"
	const signalPath = "/data/changedetection-datastore"

	// Input: snapshot ids => download postgres dump
	signalSnapshot := ResticSnapshotSelectionPrompt(snapshotMap, signalPath)

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
	fmt.Println("Restoring PV...")
	if err := RestorePvc(client, namespace, signalSnapshot, "changedetection-datastore", "250Mi", "local-path"); err != nil {
		log.Fatal(err)
	}

	// 4. Full Argo Sync
	fmt.Println("Please perform an ArgoCD sync of the application once restored.")
}
