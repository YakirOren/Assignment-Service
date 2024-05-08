package gitlab

import (
	"log/slog"

	"github.com/xanzy/go-gitlab"
)

type ProjectCreator struct {
	gitlab          *gitlab.Client
	userName        string
	gitlabData      *Data
	retries         int
	userAccessLevel gitlab.AccessLevelValue
	log             *slog.Logger
}

func New(client *gitlab.Client, UserName string, data *Data, retries int, accessLevel gitlab.AccessLevelValue, log *slog.Logger) *ProjectCreator {
	return &ProjectCreator{
		gitlab:          client,
		userName:        UserName,
		gitlabData:      data,
		retries:         retries,
		userAccessLevel: accessLevel,
		log:             log,
	}
}
