package gitlab

import (
	"github.com/xanzy/go-gitlab"
)

type ProjectCreator struct {
	gitlab     *gitlab.Client
	userName   string
	gitlabData *Data
	retries    int
}

func New(client *gitlab.Client, UserName string, data *Data, retries int) *ProjectCreator {
	return &ProjectCreator{
		gitlab:     client,
		userName:   UserName,
		gitlabData: data,
		retries:    retries,
	}
}
