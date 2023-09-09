package restore

import (
	. "backuputil/common"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"log"
)

func Servarr(client *kubernetes.Clientset, snapshotMap NamespacedSnapshotMap) {
	const namespace = "servarr"
	const jellyfinPath = "/data/servarr-jellyfin-config"
	const jellyseerrPath = "/data/servarr-jellyseerr-config"
	const radarrExportPath = "/servarr-servarr-radarr.zip"
	const sonarrExportPath = "/servarr-servarr-sonarr.zip"
	const prowlarrExportPath = "/servarr-servarr-prowlarr.zip"
	exportPaths := []string{radarrExportPath, sonarrExportPath, prowlarrExportPath}
	const tmpDir = "./tmp"

	// Input: snapshot id => download realm dump
	jellyfinSnapshot := ResticSnapshotSelectionPrompt(snapshotMap, jellyfinPath)
	jellyseerrSnapshot := ResticSnapshotSelectionPrompt(snapshotMap, jellyseerrPath)

	for _, exportPath := range exportPaths {
		snapshot := ResticSnapshotSelectionPrompt(snapshotMap, exportPath)
		if err := RestoreResticSnapshot(namespace, exportPath, snapshot, tmpDir); err != nil {
			log.Fatal(err)
		}
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

	// 3. Restore Jellyfin and Jellyseerr PVs
	fmt.Println("Restoring jellyfin PV...")
	if err := RestorePvc(client, namespace, jellyfinSnapshot, "servarr-jellyfin-config", "250Mi", "truenas-iscsi"); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Restoring jellyseerr PV...")
	if err := RestorePvc(client, namespace, jellyseerrSnapshot, "servarr-jellyseerr-config", "250Mi", "truenas-iscsi"); err != nil {
		log.Fatal(err)
	}

	// 4. Manual steps
	fmt.Println("Please perform an ArgoCD sync of the application once restored.")
	fmt.Println("Please restore radarr, sonarr and prowlarr yourself by uploading the exports in the GUI.")
}
