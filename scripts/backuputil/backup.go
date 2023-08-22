package main

import (
	"context"
	v1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	snapshotter "github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"time"
)

func Backup() {
	client, snapclient, err := InitClients()
	if err != nil {
		log.Fatal(err)
	}

	// 1. Get namespace
	namespaces, err := GetNamespaces(client)
	if err != nil {
		log.Fatal(err)
	}

	namespace := NamespaceSelectionPrompt(namespaces)

	// 2. Get deployment
	deployments, err := client.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	if len(deployments.Items) == 0 {
		log.Fatal("No deployments found")
	}

	selectedDeployments := DeploymentSelectionPrompt(deployments.Items)

	// 3. Get PVCs
	pvcNames := GetPvcsOfDeployments(client, namespace, selectedDeployments)
	selectedPvcNames := PvcSelectionPrompt(pvcNames)
	selectedPvcs, err := GetPvcsByName(client, namespace, selectedPvcNames)
	if err != nil {
		log.Fatal(err)
	}

	// 4. Create Snapshots
	for _, pvc := range selectedPvcs {
		CreateSnapshot(snapclient, namespace, pvc)
	}
}

func CreateSnapshot(snapclient *snapshotter.Clientset, namespace string, pvc corev1.PersistentVolumeClaim) error {
	timeString := time.Now().Format("200601021504")
	pvcName := pvc.ObjectMeta.Name
	snapshot := v1.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName + "-manual-" + timeString,
			Namespace: namespace,
		},
		Spec: v1.VolumeSnapshotSpec{
			VolumeSnapshotClassName: pvc.Spec.StorageClassName,
			Source: v1.VolumeSnapshotSource{
				PersistentVolumeClaimName: &pvcName,
			},
		},
	}

	_, err := snapclient.SnapshotV1().VolumeSnapshots(namespace).Create(context.Background(), &snapshot, metav1.CreateOptions{})
	return err
}
