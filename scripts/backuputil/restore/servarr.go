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
	const sonarrPath = "/data/servarr-sonarr-config"
	const radarrPath = "/data/servarr-radarr-config"
	const prowlarrPath = "/data/servarr-prowlarr-config"
	const sabnzbdPath = "/data/servarr-sabnzbd-config"

	// Input: snapshot id => download realm dump
	jellyfinSnapshot := ResticSnapshotSelectionPrompt(snapshotMap, jellyfinPath)
	jellyseerrSnapshot := ResticSnapshotSelectionPrompt(snapshotMap, jellyseerrPath)
	sonarrSnapshot := ResticSnapshotSelectionPrompt(snapshotMap, sonarrPath)
	radarrSnapshot := ResticSnapshotSelectionPrompt(snapshotMap, radarrPath)
	prowlarrSnapshot := ResticSnapshotSelectionPrompt(snapshotMap, prowlarrPath)
	sabnzbdSnapshot := ResticSnapshotSelectionPrompt(snapshotMap, sabnzbdPath)

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
	if err := RestorePvcKah(client, namespace, jellyfinSnapshot, "servarr-jellyfin-config", "50Gi", "local-path"); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Restoring jellyseerr PV...")
	if err := RestorePvcKah(client, namespace, jellyseerrSnapshot, "servarr-jellyseerr-config", "250Mi", "local-path"); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Restoring sonarr PV...")
	if err := RestorePvcKah(client, namespace, sonarrSnapshot, "servarr-sonarr-config", "250Mi", "local-path"); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Restoring radarr PV...")
	if err := RestorePvcKah(client, namespace, radarrSnapshot, "servarr-radarr-config", "250Mi", "local-path"); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Restoring prowlarr PV...")
	if err := RestorePvcKah(client, namespace, prowlarrSnapshot, "servarr-prowlarr-config", "250Mi", "local-path"); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Restoring sabnzbd PV...")
	if err := RestorePvcKah(client, namespace, sabnzbdSnapshot, "servarr-sabnzbd-config", "250Mi", "local-path"); err != nil {
		log.Fatal(err)
	}

	// 4. Manual steps
	fmt.Println("Please perform an ArgoCD sync of the application once restored.")
	fmt.Println("Please restore bazarr yourself by uploading the exports in the GUI.")
}
