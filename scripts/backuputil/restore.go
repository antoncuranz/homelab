package main

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"time"
)

func Restore() {
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

	// 3. Get PVCs and VolumeSnapshots
	snapshotsByPvc, err := GetVolumeSnapshotMap(snapclient, namespace)
	if err != nil {
		log.Fatal(err)
	}

	pvcNames := GetPvcsWithSnapshots(client, namespace, selectedDeployments, snapshotsByPvc)
	selectedPvcNames := PvcSelectionPrompt(pvcNames)
	selectedPvcs, err := GetPvcsByName(client, namespace, selectedPvcNames)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: Check if selectedPvcs are used by other, non-selected deployments

	selectedSnapshots := SnapshotSelectionPrompt(selectedPvcNames, snapshotsByPvc)

	// 4. Scale down deployment to 0 replicas
	if err := ScaleDeployments(client, namespace, selectedDeployments, 0); err != nil {
		log.Fatal(err)
	}

	// TODO: Wait for deployment to scale down
	fmt.Println("Sleeping...")
	time.Sleep(5 * time.Second)

	// 5. Delete PVC
	if err := DeletePvcs(client, namespace, selectedPvcNames); err != nil {
		log.Fatal(err)
	}

	// TODO: Wait for PVC to be deleted
	fmt.Println("Sleeping...")
	time.Sleep(5 * time.Second)

	// 6. Recreate PVCs from Snapshots
	if err := RestorePvcsFromSnapshots(client, namespace, selectedPvcs, selectedSnapshots); err != nil {
		log.Fatal(err)
	}

	// Scale deployments back up (better: sync ArgoCD Application)
	// if err := ScaleDeployments(client, namespace, selectedDeployments, 0); err != nil {
	// 	log.Fatal(err)
	// }

	// TODO: Wait for deployment to scale up
}

func RestorePvcsFromSnapshots(client *kubernetes.Clientset, namespace string, selectedPvcs []corev1.PersistentVolumeClaim, selectedSnapshots map[string]string) error {
	apiGroup := "snapshot.storage.k8s.io"

	for _, pvc := range selectedPvcs {
		obj := corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pvc.Name,
				Namespace: namespace,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				DataSource: &corev1.TypedLocalObjectReference{
					Kind:     "VolumeSnapshot",
					Name:     selectedSnapshots[pvc.Name],
					APIGroup: &apiGroup,
				},
				AccessModes: pvc.Spec.AccessModes,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"storage": pvc.Spec.Resources.Requests["storage"],
					},
				},
			},
		}
		fmt.Println(obj)
		_, err := client.CoreV1().PersistentVolumeClaims(namespace).Create(context.Background(), &obj, metav1.CreateOptions{})

		if err != nil {
			return err
		}
	}

	return nil
}
