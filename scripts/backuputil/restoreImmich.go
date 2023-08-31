package main

import (
	"bufio"
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"os"
	"os/exec"
)

const namespace = "immich"
const postgresPod = "immich-postgresql-0"
const sqlDumpPath = "/immich-postgresql.sql.gz"
const dataPath = "/data/immich-data"
const backendUrl = "http://backup.backup.svc.cluster.local:50001/"
const tmpDir = "./tmp"

func RestoreImmich() {

	client, _, err := InitClients()
	if err != nil {
		log.Fatal(err)
	}

	snapshotMap, err := CreateResticSnapshotMap(namespace)
	if err != nil {
		log.Fatal(err)
	}

	// 0. Input: snapshot ids, sql backup path
	dataSnapshot := ResticSnapshotSelectionPrompt(snapshotMap, dataPath)
	psqlSnapshot := ResticSnapshotSelectionPrompt(snapshotMap, sqlDumpPath)
	if err := RestoreResticSnapshot(namespace, sqlDumpPath, psqlSnapshot, tmpDir); err != nil {
		log.Fatal(err)
	}

	// 1. Scale down Deployments and Statefulsets
	fmt.Println("Scaling down Deployments and StatefulSets...")
	if err := ScaleDownDeploymentsInNamespace(client, namespace); err != nil {
		log.Fatal(err)
	}
	if err := ScaleDownStatefulSetsInNamespace(client, namespace); err != nil {
		log.Fatal(err)
	}
	if err := WaitUntilPodsAreDeleted(client, namespace); err != nil {
		log.Fatal(err)
	}

	// 2. Delete PVCs
	fmt.Println("Deleting PVCs...")
	if err := DeletePvcsInNamespace(client, namespace); err != nil {
		log.Fatal(err)
	}
	if err := WaitUntilPvcsAreDeleted(client, namespace); err != nil {
		log.Fatal(err)
	}

	// 3. Create and restore NFS PVC using k8up Restore CRD
	fmt.Println("Restoring immich-data PVC...")
	if err := RestorePvc(client, namespace, dataSnapshot); err != nil {
		log.Fatal(err)
	}

	// 4. (Argo-)Sync Postgres and run psql Restore
	fmt.Println("Syncing ArgoCD immich postgres STS...")
	if err := ArgoSyncResource("immich", "apps:StatefulSet:immich-postgresql"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("IS IT DONE????") // TODO: use survey confirm or wait for pod to be ready
	bufio.NewScanner(os.Stdin).Scan()

	cmdStr := "gunzip < " + tmpDir + sqlDumpPath + " | kubectl exec -i " + postgresPod + " -n " + namespace +
		" -- sh -c 'PGPASSWORD=\"$POSTGRES_POSTGRES_PASSWORD\" psql -U postgres -d immich'"
	cmd := exec.Command("sh", "-c", cmdStr)
	if _, err = cmd.Output(); err != nil {
		log.Fatal(err)
	}

	// 5. Full Argo Sync
	fmt.Println("Syncing ArgoCD immich Application...")
	if err := ArgoSyncApplication("immich"); err != nil {
		log.Fatal(err)
	}
}

func ArgoSyncApplication(application string) error {
	return exec.Command("argocd", "app", "sync", application).Run()
}

func ArgoSyncResource(application string, resource string) error {
	return exec.Command("argocd", "app", "sync", application, "--resource", resource).Run()
}

func RestorePvc(client *kubernetes.Clientset, namespace string, snapshot string) error {
	pvcName := "immich-data"
	pvcStorage := "50Gi"
	storageClass := "truenas-nfs"
	restoreName := "restore-immich"

	if err := CreatePvc(client, namespace, pvcName, pvcStorage, storageClass); err != nil {
		return err
	}

	if err := CreateRestore(client, namespace, restoreName, pvcName, snapshot); err != nil {
		log.Fatal(err)
	}

	return nil
}

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

func FindDeploymentsByNamespace(client *kubernetes.Clientset, namespace string) (*appsv1.DeploymentList, error) {
	return client.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
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

func ScaleDownDeploymentsInNamespace(client *kubernetes.Clientset, namespace string) error {
	deployments, err := FindDeploymentsByNamespace(client, namespace)
	if err != nil {
		return err
	}

	return ScaleDeployments(client, namespace, deployments.Items, 0)
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

func FindPvcsByNamespace(client *kubernetes.Clientset, namespace string) (*v1.PersistentVolumeClaimList, error) {
	return client.CoreV1().PersistentVolumeClaims(namespace).List(context.Background(), metav1.ListOptions{})
}
