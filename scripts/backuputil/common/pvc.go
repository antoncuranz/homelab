package common

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreatePvc(client *kubernetes.Clientset, namespace string, name string, storage string, storageClass string) error {
	pvc := v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					"storage": resource.MustParse(storage),
				},
			},
			StorageClassName: &storageClass,
		},
	}

	_, err := client.CoreV1().PersistentVolumeClaims(namespace).Create(context.Background(), &pvc, metav1.CreateOptions{})
	return err
}

func DeletePvcsInNamespace(client *kubernetes.Clientset, namespace string) error {
	pvcs, err := FindPvcsByNamespace(client, namespace)
	if err != nil {
		return err
	}

	for _, pvc := range pvcs.Items {
		if err := DeletePvc(client, namespace, pvc.Name); err != nil {
			return err
		}
	}

	return nil
}

func DeletePvc(client *kubernetes.Clientset, namespace string, pvcName string) error {
	err := client.CoreV1().PersistentVolumeClaims(namespace).Delete(context.Background(), pvcName, metav1.DeleteOptions{})
	return err
}

func FindPvcsByNamespace(client *kubernetes.Clientset, namespace string) (*v1.PersistentVolumeClaimList, error) {
	return client.CoreV1().PersistentVolumeClaims(namespace).List(context.Background(), metav1.ListOptions{})
}
