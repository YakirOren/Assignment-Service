package server

import (
	"fmt"
	"github.com/xanzy/go-gitlab"
	"gitlab-service/config"
	"log/slog"
)

type Server struct {
	gitlab *gitlab.Client
	logger *slog.Logger
}

func NewServer(config *config.Config, logger *slog.Logger) (*Server, error) {
	client, err := gitlab.NewClient(config.Gitlab.GitlabToken, gitlab.WithBaseURL(config.Gitlab.GitlabAPIURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &Server{
		gitlab: client,
		logger: logger,
	}, nil
}
