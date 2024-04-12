package server

import (
	"crypto/tls"
	"fmt"
	"gitlab-service/config"
	"gitlab-service/hive"
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/xanzy/go-gitlab"
)

type Server struct {
	gitlab        *gitlab.Client
	logger        *slog.Logger
	hive          hive.Hive
	templatesPath string
	retries       int
}

func NewServer(config *config.Config, logger *slog.Logger, hive hive.Hive) (*Server, error) {
	httpClient := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	client, err := gitlab.NewClient(config.Gitlab.GitlabToken, gitlab.WithBaseURL(config.Gitlab.GitlabAPIURL), gitlab.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &Server{
		gitlab:        client,
		logger:        logger,
		hive:          hive,
		retries:       config.Retries,
		templatesPath: config.TemplatesPath,
	}, nil
}

func requestID(ctx *fiber.Ctx) interface{} {
	return ctx.Context().Value("requestid")
}

func (s *Server) auth(ctx *fiber.Ctx) {
	authToken, ok := ctx.GetReqHeaders()["Authorization"]
	if !ok {
		return
	}
	if authToken[0] != "" && authToken[0] != s.hive.Token() {
		s.hive.SetToken(authToken[0])
	}
}

func (s *Server) requestLogger(ctx *fiber.Ctx, body Request) *slog.Logger {
	return s.logger.With("username", body.UserName).With("assignment_id", body.AssignmentID).With("requestid", ctx.Context().Value("requestid"))
}
