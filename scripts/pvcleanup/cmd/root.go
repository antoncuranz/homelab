package cmd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/thoas/go-funk"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"net/http"
	"net/url"
	"os"
)

const truenasApiUrl = "https://192.168.1.2/api/v2.0/pool/dataset/id/"
const nfsPathPrefix = "SSD/k8s/nfs"
const iscsiPathPrefix = "SSD/k8s/iscsi"

type TrueNASDatasetResponse struct {
	Id       string
	Children []TrueNASDataset
}

type TrueNASDataset struct {
	Id string
}

var rootCmd = &cobra.Command{
	Use:   "pvcleanup",
	Short: "Cleanup unused PVs",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := InitClient()
		if err != nil {
			log.Fatal(err)
		}

		allPvs, err := client.CoreV1().PersistentVolumes().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}

		releasedPvs := funk.Filter(allPvs.Items, func(pv v1.PersistentVolume) bool {
			return pv.Status.Phase == v1.VolumeReleased
		}).([]v1.PersistentVolume)

		nfsPvs := funk.Filter(releasedPvs, func(pv v1.PersistentVolume) bool {
			return pv.Spec.StorageClassName == "truenas-nfs"
		}).([]v1.PersistentVolume)

		iscsiPvs := funk.Filter(releasedPvs, func(pv v1.PersistentVolume) bool {
			return pv.Spec.StorageClassName == "truenas-iscsi"
		}).([]v1.PersistentVolume)

		rmOrphans, err := cmd.Flags().GetBool("rm-orphans")
		if err == nil && rmOrphans {
			if err := RemoveOrphans(allPvs.Items); err != nil {
				log.Fatal(err)
			}
			os.Exit(0)
		}

		// Input: PVs to delete
		nfsPvsToDelete := PvSelectionPrompt(nfsPvs, "Select nfs PVs to delete:")
		iscsiPvsToDelete := PvSelectionPrompt(iscsiPvs, "Select iscsi PVs to delete:")

		// 1. Delete K8s PV resource
		fmt.Println("Deleting PV resources...")
		for _, pvName := range append(nfsPvsToDelete, iscsiPvsToDelete...) {
			err := client.CoreV1().PersistentVolumes().Delete(context.Background(), pvName, metav1.DeleteOptions{})
			if err != nil {
				fmt.Println("Error deleting PV resource " + pvName)
				log.Fatal(err)
			}
		}

		// 2. Delete TrueNAS dataset
		fmt.Println("Deleting nfs datasets...")
		for _, pvName := range nfsPvsToDelete {
			status, err := TrueNasDeleteDataset(nfsPathPrefix + "/" + pvName)
			if err != nil {
				fmt.Println("Error deleting nfs dataset " + pvName)
				log.Fatal(err)
			}
			fmt.Println(status)
		}

		fmt.Println("Deleting iscsi datasets...")
		for _, pvName := range iscsiPvsToDelete {
			status, err := TrueNasDeleteDataset(iscsiPathPrefix + "/" + pvName)
			if err != nil {
				fmt.Println("Error deleting iscsi dataset " + pvName)
				log.Fatal(err)
			}
			fmt.Println(status)
		}
	},
}

func RemoveOrphans(allPvs []v1.PersistentVolume) error {
	// ISCSI
	iscsiPvs := funk.Filter(allPvs, func(pv v1.PersistentVolume) bool {
		return pv.Spec.StorageClassName == "truenas-iscsi"
	}).([]v1.PersistentVolume)

	iscsiPvNames := funk.Map(iscsiPvs, func(pv v1.PersistentVolume) string {
		return iscsiPathPrefix + "/" + pv.Name
	}).([]string)

	iscsiDatasets, err := TrueNasGetDatasets(iscsiPathPrefix)
	if err != nil {
		return err
	}

	var unusedIscsiDatasets []string
	for _, iscsiDataset := range iscsiDatasets {
		if !funk.Contains(iscsiPvNames, iscsiDataset) {
			unusedIscsiDatasets = append(unusedIscsiDatasets, iscsiDataset)
		}
	}
	fmt.Printf("%d of %d iscsi datasets are not linked.\n", len(unusedIscsiDatasets), len(iscsiDatasets))

	confirm := false
	survey.AskOne(&survey.Confirm{
		Message: "Delete unlinked datasets?",
	}, &confirm)
	if !confirm {
		os.Exit(0)
	}

	for _, pvName := range unusedIscsiDatasets {
		status, err := TrueNasDeleteDataset(pvName)
		if err != nil {
			fmt.Println("Error deleting iscsi dataset " + pvName)
			log.Fatal(err)
		}
		fmt.Println(status)
	}

	// NFS
	nfsPvs := funk.Filter(allPvs, func(pv v1.PersistentVolume) bool {
		return pv.Spec.StorageClassName == "truenas-nfs"
	}).([]v1.PersistentVolume)

	nfsPvNames := funk.Map(nfsPvs, func(pv v1.PersistentVolume) string {
		return nfsPathPrefix + "/" + pv.Name
	}).([]string)

	nfsDatasets, err := TrueNasGetDatasets(nfsPathPrefix)
	if err != nil {
		return err
	}

	var unusedNfsDatasets []string
	for _, nfsDataset := range nfsDatasets {
		if !funk.Contains(nfsPvNames, nfsDataset) {
			unusedNfsDatasets = append(unusedNfsDatasets, nfsDataset)
		}
	}
	fmt.Printf("%d of %d nfs datasets are not linked.\n", len(unusedNfsDatasets), len(nfsDatasets))

	survey.AskOne(&survey.Confirm{
		Message: "Delete unlinked datasets?",
	}, &confirm)
	if !confirm {
		os.Exit(0)
	}

	for _, pvName := range unusedNfsDatasets {
		status, err := TrueNasDeleteDataset(pvName)
		if err != nil {
			fmt.Println("Error deleting nfs dataset " + pvName)
			log.Fatal(err)
		}
		fmt.Println(status)
	}

	return nil
}

func TrueNasGetDatasets(prefix string) ([]string, error) {
	escapedPrefix := url.QueryEscape(prefix)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", truenasApiUrl+escapedPrefix, nil)
	if err != nil {
		return []string{}, err
	}
	req.Header.Add("Authorization", "Bearer "+os.Getenv("TRUENAS_API_KEY"))
	rsp, err := httpClient.Do(req)

	if err != nil {
		return []string{}, err
	}

	var parsed TrueNASDatasetResponse
	err = json.NewDecoder(rsp.Body).Decode(&parsed)
	if err != nil {
		log.Fatal(err)
	}
	children := funk.Map(parsed.Children, func(d TrueNASDataset) string { return d.Id }).([]string)
	return children, nil
}

func TrueNasDeleteDataset(datasetId string) (string, error) {
	escapedId := url.QueryEscape(datasetId)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}
	req, err := http.NewRequest("DELETE", truenasApiUrl+escapedId, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+os.Getenv("TRUENAS_API_KEY"))
	rsp, err := httpClient.Do(req)

	if err != nil {
		return "", err
	}

	return rsp.Status, nil
}

func PvSelectionPrompt(pvs []v1.PersistentVolume, message string) []string {
	if len(pvs) == 0 {
		return []string{}
	}

	pvNames := funk.Map(pvs, func(pv v1.PersistentVolume) string {
		return pv.Name
	}).([]string)

	var selectedNames []string
	err := survey.AskOne(&survey.MultiSelect{
		Message: message,
		Options: pvNames,
		Description: func(value string, index int) string {
			return pvs[index].Spec.ClaimRef.Name
		},
		VimMode: true,
	}, &selectedNames)

	if err != nil {
		os.Exit(1)
	}

	return selectedNames
}

func InitClient() (*kubernetes.Clientset, error) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(kubeConfig)
	return client, err
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("rm-orphans", "r", false, "remove datasets without pvs")
}
