package common

import (
	"context"
	"encoding/json"
	k8upv1a1 "github.com/vshn/k8up/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const backendUrl = "http://backup.backup.svc.cluster.local:50001/"

func CreateRestore(client *kubernetes.Clientset, namespace string, restoreName string, pvcName string, snapshot string) error {
	restore := k8upv1a1.Restore{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Restore",
			APIVersion: "k8up.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      restoreName,
			Namespace: namespace,
		},
		Spec: k8upv1a1.RestoreSpec{
			RestoreMethod: &k8upv1a1.RestoreMethod{
				Folder: &k8upv1a1.FolderRestore{
					&v1.PersistentVolumeClaimVolumeSource{
						ClaimName: pvcName,
					},
				},
			},
			Snapshot: snapshot,
			RunnableSpec: k8upv1a1.RunnableSpec{
				Backend: &k8upv1a1.Backend{
					Rest: &k8upv1a1.RestServerSpec{
						URL: backendUrl + namespace,
					},
				},
			},
		},
	}

	body, err := json.Marshal(restore)
	if err != nil {
		return err
	}

	rsp := k8upv1a1.Restore{}
	absPath := "/apis/k8up.io/v1/namespaces/" + namespace + "/restores"
	return client.RESTClient().Post().AbsPath(absPath).Body(body).Do(context.Background()).Into(&rsp)
}

func CreateBackup(client *kubernetes.Clientset, namespace string, backupName string) error {
	failedJobsHistoryLimit := 1
	successfulJobHistoryLimit := 0

	restore := k8upv1a1.Backup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Backup",
			APIVersion: "k8up.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      backupName,
			Namespace: namespace,
		},
		Spec: k8upv1a1.BackupSpec{
			FailedJobsHistoryLimit:     &failedJobsHistoryLimit,
			SuccessfulJobsHistoryLimit: &successfulJobHistoryLimit,
			RunnableSpec: k8upv1a1.RunnableSpec{
				Backend: &k8upv1a1.Backend{
					Rest: &k8upv1a1.RestServerSpec{
						URL: backendUrl + namespace,
					},
				},
			},
		},
	}

	body, err := json.Marshal(restore)
	if err != nil {
		return err
	}

	rsp := k8upv1a1.Restore{}
	absPath := "/apis/k8up.io/v1/namespaces/" + namespace + "/backups"
	return client.RESTClient().Post().AbsPath(absPath).Body(body).Do(context.Background()).Into(&rsp)
}
