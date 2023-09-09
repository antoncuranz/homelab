package common

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/thoas/go-funk"
	"os"
	"os/exec"
	"strings"
)

type ResticSnapshot struct {
	id   string
	date string
	time string
}
type NamespacedSnapshotMap = map[string][]ResticSnapshot
type ResticSnapshotMap = map[string]NamespacedSnapshotMap

func CreateResticSnapshotMap() (ResticSnapshotMap, error) {
	out, err := exec.Command("restic", "-r", "s3:"+S3Endpoint+"/"+S3BucketK8up, "snapshots").Output()
	if err != nil {
		return nil, err
	}

	snapshotMap := map[string]map[string][]ResticSnapshot{}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		snap := ResticSnapshot{
			id:   fields[0],
			date: fields[1],
			time: fields[2],
		}
		namespace := fields[3]
		path := fields[4]
		if snapshotMap[namespace] == nil {
			snapshotMap[namespace] = map[string][]ResticSnapshot{}
		}
		snapshotMap[namespace][path] = append(snapshotMap[namespace][path], snap)
	}

	return snapshotMap, nil
}

func ResticSnapshotSelectionPrompt(snapshotMap map[string][]ResticSnapshot, path string) string {
	snapshots := funk.Reverse(snapshotMap[path]).([]ResticSnapshot)

	snapshotDisplays := funk.Map(snapshots, func(snapshot ResticSnapshot) string {
		return snapshot.id
	}).([]string)

	var selected string
	err := survey.AskOne(&survey.Select{
		Message: "Choose snapshot for " + path + ":",
		Options: snapshotDisplays,
		Description: func(value string, index int) string {
			return snapshots[index].date + " " + snapshots[index].time
		},
		VimMode: true,
	}, &selected)

	if err != nil {
		os.Exit(1)
	}

	return selected
}

func RestoreResticSnapshot(namespace string, path string, snapshot string, target string) error {
	return exec.Command("restic", "-r", "s3:"+S3Endpoint+"/"+S3BucketK8up, "restore", snapshot, "--path", path, "--target", target).Run()
}
