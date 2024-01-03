package common

import (
	"context"
	"encoding/json"
	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	k8upv1 "github.com/k8up-io/k8up/v2/api/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
)

func CreateRestore(client *kubernetes.Clientset, namespace string, restoreName string, pvcName string, snapshot string) error {
	fsGroup := int64(65532)
	fsGroupChangePolicy := v1.FSGroupChangeOnRootMismatch
	podSecurityContext := &v1.PodSecurityContext{
		FSGroup:             &fsGroup,
		FSGroupChangePolicy: &fsGroupChangePolicy,
	}

	restore := k8upv1.Restore{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Restore",
			APIVersion: "k8up.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      restoreName,
			Namespace: namespace,
		},
		Spec: k8upv1.RestoreSpec{
			RestoreMethod: &k8upv1.RestoreMethod{
				Folder: &k8upv1.FolderRestore{
					&v1.PersistentVolumeClaimVolumeSource{
						ClaimName: pvcName,
					},
				},
			},
			Snapshot: snapshot,
			RunnableSpec: k8upv1.RunnableSpec{
				PodSecurityContext: podSecurityContext,
			},
		},
	}

	body, err := json.Marshal(restore)
	if err != nil {
		return err
	}

	rsp := k8upv1.Restore{}
	absPath := "/apis/k8up.io/v1/namespaces/" + namespace + "/restores"
	return client.RESTClient().Post().AbsPath(absPath).Body(body).Do(context.Background()).Into(&rsp)
}

func CreateRestoreKah(client *kubernetes.Clientset, namespace string, restoreName string, pvcName string, snapshot string) error {
	uid := int64(568)
	fsGroupChangePolicy := v1.FSGroupChangeOnRootMismatch
	podSecurityContext := &v1.PodSecurityContext{
		RunAsUser:           &uid,
		FSGroup:             &uid,
		FSGroupChangePolicy: &fsGroupChangePolicy,
	}

	restore := k8upv1.Restore{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Restore",
			APIVersion: "k8up.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      restoreName,
			Namespace: namespace,
		},
		Spec: k8upv1.RestoreSpec{
			RestoreMethod: &k8upv1.RestoreMethod{
				Folder: &k8upv1.FolderRestore{
					&v1.PersistentVolumeClaimVolumeSource{
						ClaimName: pvcName,
					},
				},
			},
			Snapshot: snapshot,
			RunnableSpec: k8upv1.RunnableSpec{
				PodSecurityContext: podSecurityContext,
			},
		},
	}

	body, err := json.Marshal(restore)
	if err != nil {
		return err
	}

	rsp := k8upv1.Restore{}
	absPath := "/apis/k8up.io/v1/namespaces/" + namespace + "/restores"
	return client.RESTClient().Post().AbsPath(absPath).Body(body).Do(context.Background()).Into(&rsp)
}

func CreateBackup(client *kubernetes.Clientset, namespace string, backupName string, runAsUser int64) error {
	runnableSpec := k8upv1.RunnableSpec{}

	if runAsUser != -1 {
		runnableSpec = k8upv1.RunnableSpec{
			PodSecurityContext: &v1.PodSecurityContext{
				RunAsUser: &runAsUser,
			},
		}
	}

	backup := k8upv1.Backup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Backup",
			APIVersion: "k8up.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      backupName,
			Namespace: namespace,
		},
		Spec: k8upv1.BackupSpec{
			RunnableSpec: runnableSpec,
		},
	}

	body, err := json.Marshal(backup)
	if err != nil {
		return err
	}

	rsp := k8upv1.Backup{}
	absPath := "/apis/k8up.io/v1/namespaces/" + namespace + "/backups"
	return client.RESTClient().Post().AbsPath(absPath).Body(body).Do(context.Background()).Into(&rsp)
}

func CreateDatabaseBackup(client *kubernetes.Clientset, backupName string) error {
	backup := cnpgv1.Backup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Backup",
			APIVersion: "postgresql.cnpg.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      backupName,
			Namespace: "database",
		},
		Spec: cnpgv1.BackupSpec{
			Cluster: cnpgv1.LocalObjectReference{
				Name: "database-cluster",
			},
		},
	}

	body, err := json.Marshal(backup)
	if err != nil {
		return err
	}

	rsp := cnpgv1.Backup{}
	absPath := "/apis/postgresql.cnpg.io/v1/namespaces/database/backups"
	return client.RESTClient().Post().AbsPath(absPath).Body(body).Do(context.Background()).Into(&rsp)
}

func RestorePvc(client *kubernetes.Clientset, namespace string, snapshot string, pvcName string, pvcStorage string, storageClass string) error {
	if err := CreatePvc(client, namespace, pvcName, pvcStorage, storageClass); err != nil {
		return err
	}

	restoreName := "restore-" + pvcName
	if err := CreateRestore(client, namespace, restoreName, pvcName, snapshot); err != nil {
		log.Fatal(err)
	}

	return nil
}

func RestorePvcKah(client *kubernetes.Clientset, namespace string, snapshot string, pvcName string, pvcStorage string, storageClass string) error {
	if err := CreatePvc(client, namespace, pvcName, pvcStorage, storageClass); err != nil {
		return err
	}

	restoreName := "restore-" + pvcName
	if err := CreateRestoreKah(client, namespace, restoreName, pvcName, snapshot); err != nil {
		log.Fatal(err)
	}

	return nil
}
