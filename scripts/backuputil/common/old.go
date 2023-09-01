package common

import (
	"context"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/thoas/go-funk"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
)

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
