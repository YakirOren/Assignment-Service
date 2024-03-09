package server

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/xanzy/go-gitlab"
	"log/slog"
)

type Request struct {
	UserName       string `json:"user_name"`
	AssignmentID   int    `json:"assignment_id"`
	ExerciseID     int    `json:"exercise_id"`
	OnCreationData struct {
		Gitlab *GitlabData `json:"gitlab"`
	} `json:"data"`
}

type GitlabData struct {
	Namespace        string `json:"namespace"`
	SourceRepo       string `json:"source_repo"`
	NewRepoName      string `json:"new_repo_name"`
	HiveInstructions bool   `json:"hive_instructions"`
}

const RepoCreationFailed = "# ❌ Failed to create a Gitlab repository\nplease ask for help"

func (s *Server) OnNewAssignment(ctx *fiber.Ctx) error {
	var body Request
	if err := ctx.BodyParser(&body); err != nil {
		return fmt.Errorf("failed to parse the request body: %w", err)
	}

	log := s.requestLogger(ctx, body.UserName)

	s.auth(ctx, log)

	if body.OnCreationData.Gitlab != nil {
		s.hive.UpdateAssignment(body.AssignmentID, "Your repo is getting ready please wait....")
		err := s.processGitlab(body, log)
		if err != nil {
			log.Error(err.Error())
			s.hive.UpdateAssignment(body.AssignmentID, fmt.Sprintf("%s\nrequestID=%s", RepoCreationFailed, ctx.Context().Value("requestid")))
			return err
		}
	}

	return ctx.SendString("DONE")
}

func (s *Server) auth(ctx *fiber.Ctx, log *slog.Logger) {
	authToken, ok := ctx.GetReqHeaders()["Authorization"]
	if !ok {
		return
	}
	if authToken[0] != "" && authToken[0] != s.hive.Token() {
		s.hive.SetToken(authToken[0])
		log.Info("auth token has changed")
	}
}

func (s *Server) requestLogger(ctx *fiber.Ctx, username string) *slog.Logger {
	return s.logger.With("username", username).With("requestid", ctx.Context().Value("requestid"))
}

func (s *Server) processGitlab(body Request, log *slog.Logger) error {
	data := *body.OnCreationData.Gitlab

	user, exists := s.listUsersByName(body.UserName)
	if !exists {
		return fiber.NewError(fiber.StatusBadRequest, "user doesn't exist on gitlab")
	}

	var usersGroup *gitlab.Group

	log.Info("checking if the users subgroup exists")
	usersGroup, exists = s.listGroupsByName(body.UserName, data.Namespace)
	if !exists {
		log.Info("creating subgroup for the user")
		var err error
		usersGroup, err = s.createUsersSubGroup(body.UserName, data.Namespace)
		if err != nil {
			var e *gitlab.ErrorResponse
			if errors.As(err, &e) {
				log.Error(e.Message)
			}

			return fiber.NewError(fiber.StatusInternalServerError, "failed to create subgroup")
		}
	}

	log.Info("creating new repo")
	project, err := s.CreateRepoInGroup(usersGroup, data)
	if err != nil {
		return fmt.Errorf("failed to create new repo inside the users group: %w", err)
	}

	log.Info("created repo", slog.String("path", project.Path))

	log.Info("adding the user to the new group")
	_, _, err = s.addUserToProject(user, project)
	if err != nil {
		return fmt.Errorf("failed to add user to project: %w", err)
	}

	log.Info("creating description")
	err = s.hive.UpdateAssignment(body.AssignmentID, "git description")
	if err != nil {
		return fmt.Errorf("failed to update description")
	}

	return nil
}
