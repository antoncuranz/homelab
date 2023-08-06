package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/AlecAivazis/survey/v2"
	snapshotter "github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned"
	"github.com/thoas/go-funk"
)

func main() {
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

func InitClients() (*kubernetes.Clientset, *snapshotter.Clientset, error) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		return nil, nil, err
	}

	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, nil, err
	}

	snapclient, err := snapshotter.NewForConfig(kubeConfig)
	if err != nil {
		return nil, nil, err
	}

	return client, snapclient, nil
}

func GetNamespaces(client *kubernetes.Clientset) ([]string, error) {
	namespaces, err := client.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	namespaceNames := funk.Map(namespaces.Items, func(namespace corev1.Namespace) string {
		return namespace.Name
	}).([]string)
	return namespaceNames, err
}

func GetDeploymentNames(client *kubernetes.Clientset, namespace string) ([]string, error) {
	deployments, err := client.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	deploymentNames := funk.Map(deployments.Items, func(deployment appsv1.Deployment) string {
		return deployment.Name
	}).([]string)

	return deploymentNames, nil
}

func GetPvcsOfDeployments(client *kubernetes.Clientset, namespace string, deployments []appsv1.Deployment) []string {
	var pvcNames []string
	for _, deployment := range deployments {
		for _, volume := range deployment.Spec.Template.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil {
				if !funk.Contains(pvcNames, volume.PersistentVolumeClaim.ClaimName) {
					pvcNames = append(pvcNames, volume.PersistentVolumeClaim.ClaimName)
				}
			}
		}
	}
	return pvcNames
}

func GetPvcsWithSnapshots(client *kubernetes.Clientset, namespace string, deployments []appsv1.Deployment, snapshotsByPvc map[string][]string) []string {
	pvcNames := GetPvcsOfDeployments(client, namespace, deployments)

	for _, pvcName := range pvcNames {
		if len(snapshotsByPvc[pvcName]) == 0 {
			fmt.Printf("Warning: PVC %s has no snapshots\n", pvcName)
		}
	}
	pvcNames = funk.Filter(pvcNames, func(pvcName string) bool {
		return len(snapshotsByPvc[pvcName]) > 0
	}).([]string)

	return pvcNames
}

func GetVolumeSnapshotMap(snapclient *snapshotter.Clientset, namespace string) (map[string][]string, error) {
	snaps, err := snapclient.SnapshotV1().VolumeSnapshots(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	snapshotsByPvc := make(map[string][]string)

	for _, snap := range snaps.Items {
		pvcName := *snap.Spec.Source.PersistentVolumeClaimName
		if !funk.Contains(snapshotsByPvc[pvcName], snap.Name) {
			snapshotsByPvc[pvcName] = append(snapshotsByPvc[pvcName], snap.Name)
		}
	}

	return snapshotsByPvc, err
}

func NamespaceSelectionPrompt(namespaces []string) string {
	var namespace string
	err := survey.AskOne(&survey.Select{
		Message: "Choose namespace:",
		Options: namespaces,
		VimMode: true,
	}, &namespace)
	if err != nil {
		log.Fatal(err)
	}

	return namespace
}

func DeploymentSelectionPrompt(deployments []appsv1.Deployment) []appsv1.Deployment {
	deploymentNames := funk.Map(deployments, func(deployment appsv1.Deployment) string {
		return deployment.Name
	}).([]string)

	var selectedNames []string
	err := survey.AskOne(&survey.MultiSelect{
		Message: "Choose deployments:",
		Options: deploymentNames,
		VimMode: true,
	}, &selectedNames)
	if err != nil {
		log.Fatal(err)
	}

	return funk.Filter(deployments, func(deployment appsv1.Deployment) bool {
		return funk.Contains(selectedNames, deployment.Name)
	}).([]appsv1.Deployment)
}

func PvcSelectionPrompt(pvcNames []string) []string {
	var selectedPvcs []string
	err := survey.AskOne(&survey.MultiSelect{
		Message: "Choose PVCs:",
		Options: pvcNames,
		VimMode: true,
	}, &selectedPvcs)
	if err != nil {
		log.Fatal(err)
	}

	return selectedPvcs
}

func SnapshotSelectionPrompt(pvcNames []string, snapshotsByPvc map[string][]string) map[string]string {
	selectedSnapshots := make(map[string]string, len(pvcNames))
	for _, pvcName := range pvcNames {
		var choice string
		survey.AskOne(
			&survey.Select{
				Message: fmt.Sprintf("Choose snapshot for PVC %s:", pvcName),
				Options: snapshotsByPvc[pvcName],
				VimMode: true,
			},
			&choice,
		)
		selectedSnapshots[pvcName] = choice
	}

	return selectedSnapshots
}

func ScaleDeployment(client *kubernetes.Clientset, namespace string, deployment appsv1.Deployment, replicas int32) error {
	*deployment.Spec.Replicas = replicas
	_, err := client.AppsV1().Deployments(namespace).Update(context.Background(), &deployment, metav1.UpdateOptions{})
	return err
}

func ScaleDeployments(client *kubernetes.Clientset, namespace string, deployments []appsv1.Deployment, replicas int32) error {
	for _, deployment := range deployments {
		if err := ScaleDeployment(client, namespace, deployment, replicas); err != nil {
			return err
		}
	}

	return nil
}

func DeletePvc(client *kubernetes.Clientset, namespace string, pvcName string) error {
	err := client.CoreV1().PersistentVolumeClaims(namespace).Delete(context.Background(), pvcName, metav1.DeleteOptions{})
	return err
}

func DeletePvcs(client *kubernetes.Clientset, namespace string, pvcNames []string) error {
	for _, pvcName := range pvcNames {
		if err := DeletePvc(client, namespace, pvcName); err != nil {
			return err
		}
	}

	return nil
}

func GetPvcsByName(client *kubernetes.Clientset, namespace string, pvcNames []string) ([]corev1.PersistentVolumeClaim, error) {
	pvcs := make([]corev1.PersistentVolumeClaim, len(pvcNames))
	for i, pvcName := range pvcNames {
		pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), pvcName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		pvcs[i] = *pvc
	}

	return pvcs, nil
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
