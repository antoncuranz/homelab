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

func CreateResticSnapshotMap(namespace string) (map[string][]ResticSnapshot, error) {
	out, err := exec.Command("restic", "-r", "rclone:koofr:k8up/"+namespace, "snapshots").Output()
	if err != nil {
		return nil, err
	}

	snapshotMap := map[string][]ResticSnapshot{}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, namespace) {
			fields := strings.Fields(line)
			snap := ResticSnapshot{
				id:   fields[0],
				date: fields[1],
				time: fields[2],
			}
			path := fields[4]
			snapshotMap[path] = append(snapshotMap[path], snap)
		}
	}

	return snapshotMap, nil
}

func ResticSnapshotSelectionPrompt(snapshotMap map[string][]ResticSnapshot, path string) string {
	snapshots := snapshotMap[path]

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
	return exec.Command("restic", "-r", "rclone:koofr:k8up/"+namespace, "restore", snapshot, "--path", path, "--target", target).Run()
}
