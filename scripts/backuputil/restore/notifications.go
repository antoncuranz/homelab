package restore

import (
	. "backuputil/common"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"log"
)

func Notifications(client *kubernetes.Clientset, snapshotMap NamespacedSnapshotMap) {
	const namespace = "notifications"
	const signalPath = "/data/notifications-signal-cli-config"

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
	if err := RestorePvc(client, namespace, signalSnapshot, "notifications-signal-cli-config", "250Mi", "local-path"); err != nil {
		log.Fatal(err)
	}

	// 4. Full Argo Sync
	fmt.Println("Please perform an ArgoCD sync of the application once restored.")
}
