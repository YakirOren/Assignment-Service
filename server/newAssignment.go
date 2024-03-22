package server

import (
	"encoding/json"
	"errors"
	"fmt"
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
		Gitlab      *GitlabData  `json:"gitlab"`
		Description *Description `json:"description"`
	} `json:"data"`
}

type Description struct {
	Instructions      []string `json:"instructions"`
	CustomInstruction *string  `json:"custom_instructions"`
}

type GitlabData struct {
	Namespace            string `json:"namespace"`
	SourceRepo           string `json:"source_repo"`
	NewRepoName          string `json:"new_repo_name"`
	BaseBranchName       string `json:"base_branch_name"`
	WorkBranchName       string `json:"work_branch_name"`
	DetailedInstructions *bool  `json:"detailed_instructions"`
}

const RepoCreationFailed = "# ❌ Failed to create a Gitlab repository\nplease ask for help"

func (s *Server) OnNewAssignment(ctx *fiber.Ctx) error {
	var body Request
	if err := ctx.BodyParser(&body); err != nil {
		return fmt.Errorf("failed to parse the request body: %w", err)
	}

	log := s.requestLogger(ctx, body)

	s.auth(ctx)

	if body.OnCreationData.Gitlab != nil {
		go s.hive.UpdateAssignment(body.AssignmentID, "Your repo is getting ready please wait....")
		project, err := s.processGitlab(body, log)
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

		if err := s.UpdateAssignmentWithTemplate(body.AssignmentID, filepath.Join(s.templatesPath, gitTemplateFile), a); err != nil {
			log.Error(err.Error())
		}

		return nil
	}

	if body.OnCreationData.Description != nil {
		desc := body.OnCreationData.Description

		for _, t := range desc.Instructions {
			if err := s.UpdateAssignmentWithTemplate(body.AssignmentID, filepath.Join(s.templatesPath, t), body); err != nil {
				log.Error(err.Error())
			}
		}
	}

	return ctx.SendString("DONE")
}

func (s *Server) processGitlab(body Request, log *slog.Logger) (*gitlab.Project, error) {
	data := *body.OnCreationData.Gitlab

	user, exists := s.listUsersByName(body.UserName)
	if !exists {
		return nil, fiber.NewError(fiber.StatusBadRequest, "user doesn't exist on gitlab")
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

			return nil, fiber.NewError(fiber.StatusInternalServerError, "failed to create subgroup")
		}
	}

	log.Info("check if target repo exists")
	project, exists := s.projectExists(body)
	if exists {
		log.Info("project already exists")
		return project, nil
	}

	log.Info("creating new repo")
	project, err := s.createRepoInGroup(usersGroup, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create new repo inside the users group: %w", err)
	}

	log.Info("created repo", slog.String("path", project.Path))

	log.Info("adding the user to the new group")
	if err = s.addUserToProject(user, project); err != nil {
		return nil, fmt.Errorf("failed to add user to project: %w", err)
	}

	if data.WorkBranchName == data.BaseBranchName {
		if err := s.removeBranchProtection(project); err != nil {
			log.Error("failed to remove branch protection: %w", err)
		}
	}

	return project, nil
}
