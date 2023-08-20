package main

import (
	"context"
	"log"
	"os"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/AlecAivazis/survey/v2"
	v1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
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
