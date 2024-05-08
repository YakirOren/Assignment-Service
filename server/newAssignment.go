package server

import (
	"encoding/json"
	"errors"
	"fmt"
	gitlabWrapper "gitlab-service/gitlab"
	"log/slog"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/xanzy/go-gitlab"
)

type Request struct {
	UserName       string `json:"user_name"`
	AssignmentID   int    `json:"assignment_id"`
	ExerciseID     int    `json:"exercise_id"`
	OnCreationData struct {
		Gitlab      *gitlabWrapper.Data `json:"gitlab"`
		Description *Description        `json:"description"`
	} `json:"data"`
}

type Description struct {
	Instructions      []string `json:"instructions"`
	CustomInstruction *string  `json:"custom_instructions"`
}

const RepoCreationFailed = "# ❌ Failed to create a Gitlab repository\nplease ask for help"

func (s *Server) OnNewAssignment(ctx *fiber.Ctx) error {
	var body Request
	if err := ctx.BodyParser(&body); err != nil {
		return fmt.Errorf("failed to parse the request body: %w", err)
	}

	log := s.requestLogger(ctx, body)

	s.auth(ctx)

	gitlabData := body.OnCreationData.Gitlab

	if gitlabData != nil {
		go s.hive.UpdateAssignment(body.AssignmentID, "Your repo is getting ready please wait....")
		project, err := s.processGitlab(gitlabData, body.UserName, log)
		if err != nil {
			log.Error(err.Error())
			res2B, _ := json.Marshal(body)
			log.Info(string(res2B))
			go s.hive.UpdateAssignment(body.AssignmentID, fmt.Sprintf("%s\nrequestID=%s", RepoCreationFailed, requestID(ctx)))
			return err
		}
		log.Info("creating description")

		gitTemplateFile := "git.md"
		detailedInstructions := body.OnCreationData.Gitlab.DetailedInstructions
		if detailedInstructions != nil && *detailedInstructions == false {
			gitTemplateFile = "short-git.md"
		}

		a := struct {
			UserName       string
			BaseBranchName string
			WorkBranchName string
			Project        *gitlab.Project
		}{body.UserName, body.OnCreationData.Gitlab.BaseBranchName, body.OnCreationData.Gitlab.WorkBranchName, project}

		if err := s.UpdateAssignmentWithTemplate(body.AssignmentID, filepath.Join(s.config.TemplatesPath, gitTemplateFile), a); err != nil {
			log.Error(err.Error())
		}

		return nil
	}

	if body.OnCreationData.Description != nil {
		desc := body.OnCreationData.Description

		for _, t := range desc.Instructions {
			if err := s.UpdateAssignmentWithTemplate(body.AssignmentID, filepath.Join(s.config.TemplatesPath, t), body); err != nil {
				log.Error(err.Error())
			}
		}
	}

	return ctx.SendString("DONE")
}

func (s *Server) processGitlab(data *gitlabWrapper.Data, username string, log *slog.Logger) (*gitlab.Project, error) {
	wrapper := gitlabWrapper.New(s.gitlab,
		username, data, s.config.Retries, s.config.AccessLevel, log)
	user, exists := wrapper.GetUser()
	if !exists {
		return nil, fiber.NewError(fiber.StatusBadRequest, "user doesn't exist on gitlab")
	}

	log.Info("checking if the users subgroup exists")
	usersGroup, exists := wrapper.GetUsersGroup()
	if !exists {
		log.Info("creating subgroup for the user")
		var err error
		usersGroup, err = wrapper.CreateUsersSubGroup()
		if err != nil {
			var e *gitlab.ErrorResponse
			if errors.As(err, &e) {
				log.Error(e.Message)
			}

			return nil, fiber.NewError(fiber.StatusInternalServerError, "failed to create subgroup")
		}
	}

	log.Info("check if target repo exists")
	project, exists := wrapper.GetProject()
	if exists {
		log.Info("project already exists")
		return project, nil
	}

	project, err := wrapper.CreateRepoInGroup(usersGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to create new repo inside the users group: %w", err)
	}

	log.Info("created repo", slog.String("path", project.Path))

	if err = wrapper.AddUserToProject(user, project); err != nil {
		return nil, fmt.Errorf("failed to add user to project: %w", err)
	}

	if data.WorkBranchName == data.BaseBranchName {
		if err := wrapper.RemoveBranchProtection(project); err != nil {
			log.Error("failed to remove branch protection: %w", err)
		}
	}

	return project, nil
}
