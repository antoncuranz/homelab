package common

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func FindDeploymentsByNamespace(client *kubernetes.Clientset, namespace string) (*appsv1.DeploymentList, error) {
	return client.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
}

func ScaleDownDeploymentsInNamespace(client *kubernetes.Clientset, namespace string) error {
	deployments, err := FindDeploymentsByNamespace(client, namespace)
	if err != nil {
		return err
	}

	return ScaleDeployments(client, namespace, deployments.Items, 0)
}

func ScaleDeployments(client *kubernetes.Clientset, namespace string, deployments []appsv1.Deployment, replicas int32) error {
	for _, deployment := range deployments {
		if err := ScaleDeployment(client, namespace, deployment, replicas); err != nil {
			return err
		}
	}

	return nil
}

func ScaleDeployment(client *kubernetes.Clientset, namespace string, deployment appsv1.Deployment, replicas int32) error {
	*deployment.Spec.Replicas = replicas
	_, err := client.AppsV1().Deployments(namespace).Update(context.Background(), &deployment, metav1.UpdateOptions{})
	return err
}

func FindStatefulsetsByNamespace(client *kubernetes.Clientset, namespace string) (*appsv1.StatefulSetList, error) {
	return client.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{})
}

func ScaleDownStatefulSetsInNamespace(client *kubernetes.Clientset, namespace string) error {
	statefulSets, err := FindStatefulsetsByNamespace(client, namespace)
	if err != nil {
		return err
	}

	return ScaleStatefulSets(client, namespace, statefulSets.Items, 0)
}

func ScaleStatefulSets(client *kubernetes.Clientset, namespace string, statefulSets []appsv1.StatefulSet, replicas int32) error {
	for _, statefulSet := range statefulSets {
		if err := ScaleStatefulSet(client, namespace, statefulSet, replicas); err != nil {
			return err
		}
	}

	return nil
}

func ScaleStatefulSet(client *kubernetes.Clientset, namespace string, statefulSet appsv1.StatefulSet, replicas int32) error {
	*statefulSet.Spec.Replicas = replicas
	_, err := client.AppsV1().StatefulSets(namespace).Update(context.Background(), &statefulSet, metav1.UpdateOptions{})
	return err
}

func ScaleDownDeploymentsAndStatefulSets(client *kubernetes.Clientset, namespace string) error {
	if err := ScaleDownDeploymentsInNamespace(client, namespace); err != nil {
		return err
	}
	if err := ScaleDownStatefulSetsInNamespace(client, namespace); err != nil {
		return err
	}

	return WaitUntilPodsAreDeleted(client, namespace)
}
