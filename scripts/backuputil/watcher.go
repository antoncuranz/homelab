package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

func WaitUntilPvcsAreDeleted(client *kubernetes.Clientset, namespace string) error {
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

func WaitUntilPodIsReady(client *kubernetes.Clientset, namespace string, pod string) error {
	//watcher, err := client.CoreV1().Pods(namespace).Watch(context.Background(), metav1.ListOptions{})
	return nil
}
