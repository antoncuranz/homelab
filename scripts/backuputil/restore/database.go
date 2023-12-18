package restore

import (
	. "backuputil/common"
	"context"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
)

func Database(client *kubernetes.Clientset) {
	const namespace = "database"
	const databaseName = "database-cluster"

	fmt.Println("Make sure to rename the database folder before continuing (rclone move -v b2:homelab-db/database-cluster b2:homelab-db/database-backup).")
	backupFolder := ""
	prompt := &survey.Input{
		Message: "Please enter the name of the backup folder in Backblaze:",
	}
	if err := survey.AskOne(prompt, &backupFolder); err != nil {
		log.Fatal(err)
	}
	if backupFolder == databaseName {
		log.Fatal("Error: backup folder cannot be the same as the database folder.")
	}

	fmt.Println("Deleting database CRD...")
	absPath := "/apis/postgresql.cnpg.io/v1/namespaces/database/clusters/" + databaseName
	if err := client.RESTClient().Delete().AbsPath(absPath).Do(context.Background()).Error(); err != nil {
		fmt.Println("Warning:", err)
	}

	fmt.Println("Deleting PVCs...")
	if err := DeletePvcsInNamespace(client, namespace); err != nil {
		log.Fatal(err)
	}
	if err := WaitUntilPvcsAreDeleted(client, namespace); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Restoring database...")
	if err := CreateDatabaseFromBackup(client, namespace, databaseName, backupFolder, S3Endpoint, S3BucketDb); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Please perform an ArgoCD sync of the application once restored.")
}

func CreateDatabaseFromBackup(client *kubernetes.Clientset, namespace string, databaseName string, backupFolder string, s3endpoint string, s3bucket string) error {
	storageClass := "truenas-iscsi"
	cluster := cnpgv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "postgresql.cnpg.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      databaseName,
			Namespace: namespace,
		},
		Spec: cnpgv1.ClusterSpec{
			Instances:             1,
			ImageName:             "ghcr.io/bo0tzz/cnpgvecto.rs:15.5-v0.1.11",
			PrimaryUpdateStrategy: "unsupervised",
			StorageConfiguration: cnpgv1.StorageConfiguration{
				Size:         "10Gi",
				StorageClass: &storageClass,
			},
			SuperuserSecret: &cnpgv1.LocalObjectReference{
				Name: "postgres-secrets",
			},
			Bootstrap: &cnpgv1.BootstrapConfiguration{
				Recovery: &cnpgv1.BootstrapRecovery{
					Source: "clusterBackup",
					//RecoveryTarget: &cnpgv1.RecoveryTarget{
					//	TargetTime: "2023-12-18 15:00:00.00000+02",
					//},
				},
			},
			ExternalClusters: []cnpgv1.ExternalCluster{
				{
					Name: "clusterBackup",
					BarmanObjectStore: &cnpgv1.BarmanObjectStoreConfiguration{
						ServerName:      backupFolder,
						DestinationPath: "s3://" + s3bucket + "/",
						EndpointURL:     "https://" + s3endpoint,
						BarmanCredentials: cnpgv1.BarmanCredentials{
							AWS: &cnpgv1.S3Credentials{
								AccessKeyIDReference: &cnpgv1.SecretKeySelector{
									LocalObjectReference: cnpgv1.LocalObjectReference{
										Name: "backblaze-db",
									},
									Key: "ACCESS_KEY_ID",
								},
								SecretAccessKeyReference: &cnpgv1.SecretKeySelector{
									LocalObjectReference: cnpgv1.LocalObjectReference{
										Name: "backblaze-db",
									},
									Key: "ACCESS_SECRET_KEY",
								},
							},
						},
					},
				},
			},
		},
	}

	body, err := json.Marshal(cluster)
	if err != nil {
		return err
	}

	absPath := "/apis/postgresql.cnpg.io/v1/namespaces/database/clusters"
	return client.RESTClient().Post().AbsPath(absPath).Body(body).Do(context.Background()).Error()
}
