package restore

import (
	. "backuputil/common"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"log"
)

func Paperless(client *kubernetes.Clientset, snapshotMap NamespacedSnapshotMap) {
	const namespace = "paperless"
	const dataPath = "/data/paperless-data"
	const mediaPath = "/data/paperless-media"

	// Input: snapshot ids => download postgres dump
	dataSnapshot := ResticSnapshotSelectionPrompt(snapshotMap, dataPath)
	mediaSnapshot := ResticSnapshotSelectionPrompt(snapshotMap, mediaPath)

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
	fmt.Println("Restoring paperless data PV...")
	if err := RestorePvc(client, namespace, dataSnapshot, "paperless-data", "5Gi", "nfs-paperless"); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Restoring paperless media PV...")
	if err := RestorePvc(client, namespace, mediaSnapshot, "paperless-media", "5Gi", "nfs-paperless"); err != nil {
		log.Fatal(err)
	}

	// 4. Full Argo Sync
	fmt.Println("Please perform an ArgoCD sync of the application once restored.")
}
