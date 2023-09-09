package common

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

func WaitUntilPvcsAreDeleted(client *kubernetes.Clientset, namespace string) error {
	// TODO: race condition?
	pvcList, err := client.CoreV1().PersistentVolumeClaims(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	if len(pvcList.Items) == 0 {
		return nil
	}

	watcher, err := client.CoreV1().PersistentVolumeClaims(namespace).Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	added := 0
	deleted := 0

	for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Added:
			added++
		case watch.Deleted:
			deleted++
		}

		fmt.Printf("\r%d/%d pvcs deleted", deleted, added)

		if event.Type == watch.Deleted && added == deleted {
			return nil
		}
	}

	return nil
}

func WaitUntilPodsAreDeleted(client *kubernetes.Clientset, namespace string) error {
	podList, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	if len(podList.Items) == 0 {
		return nil
	}

	watcher, err := client.CoreV1().Pods(namespace).Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	added := 0
	deleted := 0

	for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Added:
			added++
		case watch.Deleted:
			deleted++
		}

		fmt.Printf("\r%d/%d pods deleted", deleted, added)

		if event.Type == watch.Deleted && added == deleted {
			return nil
		}
	}

	return nil
}

func WaitUntilPodIsReady(client *kubernetes.Clientset, namespace string, podName string) error {
	options := metav1.ListOptions{
		FieldSelector: "metadata.name=" + podName,
	}
	watcher, err := client.CoreV1().Pods(namespace).Watch(context.Background(), options)
	if err != nil {
		return err
	}

	for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Added:
		case watch.Modified:
			pod := event.Object.(*corev1.Pod)
			if pod.Status.Phase == "Running" && AreContainersReady(pod.Status.ContainerStatuses) {
				return nil
			}
		}
	}

	return nil
}

func AreContainersReady(containerStatuses []corev1.ContainerStatus) bool {
	for _, status := range containerStatuses {
		if !status.Ready {
			return false
		}
	}

	return true
}
