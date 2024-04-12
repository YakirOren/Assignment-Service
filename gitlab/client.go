package gitlab

import (
	"github.com/xanzy/go-gitlab"
)

type ProjectCreator struct {
	gitlab          *gitlab.Client
	userName        string
	gitlabData      *Data
	retries         int
	userAccessLevel gitlab.AccessLevelValue
}

func New(client *gitlab.Client, UserName string, data *Data, retries int, accessLevel gitlab.AccessLevelValue) *ProjectCreator {
	return &ProjectCreator{
		gitlab:          client,
		userName:        UserName,
		gitlabData:      data,
		retries:         retries,
		userAccessLevel: accessLevel,
	}
}
