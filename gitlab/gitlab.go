package gitlab

import (
	"errors"
	"log/slog"
	"math"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/xanzy/go-gitlab"
)

type Data struct {
	Namespace            string `json:"namespace"`
	SourceRepo           string `json:"source_repo"`
	NewRepoName          string `json:"new_repo_name"`
	BaseBranchName       string `json:"base_branch_name"`
	WorkBranchName       string `json:"work_branch_name"`
	DetailedInstructions *bool  `json:"detailed_instructions"`
}

func (s *ProjectCreator) GetUser() (*gitlab.User, bool) {
	users, r, _ := s.gitlab.Users.ListUsers(&gitlab.ListUsersOptions{
		Username: gitlab.Ptr(s.userName),
	})

	if r != nil && r.StatusCode == http.StatusNotFound {
		return nil, false
	}

	if len(users) == 0 {
		return nil, false
	}

	return users[0], true
}
func (s *ProjectCreator) GetUsersGroup() (*gitlab.Group, bool) {
	groups, r, _ := s.gitlab.Groups.ListSubGroups(s.gitlabData.Namespace, &gitlab.ListSubGroupsOptions{
		Search: gitlab.Ptr(s.userName),
	})

	if r.StatusCode == http.StatusNotFound {
		return nil, false
	}

	if len(groups) == 0 {
		return nil, false
	}

	return groups[0], true
}
func (s *ProjectCreator) GetProject() (*gitlab.Project, bool) {
	repo := strings.Replace(s.gitlabData.NewRepoName, " ", "_", -1)
	repoPath := filepath.Join(s.gitlabData.Namespace, s.userName, repo)
	project, r, err := s.gitlab.Projects.GetProject(repoPath, nil)
	if err != nil {
		return nil, false
	}

	if r.StatusCode == http.StatusNotFound {
		return nil, false
	}
	return project, true
}
func (s *ProjectCreator) AddUserToProject(user *gitlab.User, group *gitlab.Project) error {
	opt := &gitlab.AddProjectMemberOptions{
		UserID:      user.ID,
		AccessLevel: gitlab.Ptr(s.userAccessLevel),
	}
	s.log.Info("adding the user to the new group")
	_, _, err := s.gitlab.ProjectMembers.AddProjectMember(group.ID, opt)
	return err
}
func (s *ProjectCreator) CreateRepoInGroup(group *gitlab.Group) (*gitlab.Project, error) {
	path := strings.Replace(s.gitlabData.NewRepoName, " ", "_", -1)

	opt := &gitlab.ForkProjectOptions{
		MergeRequestDefaultTargetSelf: gitlab.Ptr(true),
		Name:                          gitlab.Ptr(s.gitlabData.NewRepoName),
		NamespaceID:                   gitlab.Ptr(group.ID),
		Path:                          gitlab.Ptr(path),
		Visibility:                    gitlab.Ptr(gitlab.PrivateVisibility),
	}

	s.log.Info("creating new repo", slog.String("path", path))

	project, _, err := s.gitlab.Projects.ForkProject(s.gitlabData.SourceRepo, opt)
	if err != nil {
		return nil, err
	}

	if err := s.waitForProjectCreation(); err != nil {
		return nil, err
	}

	_, err = s.gitlab.Projects.DeleteProjectForkRelation(project.ID)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectCreator) waitForProjectCreation() error {
	for i := 0; i < s.retries; i++ {
		p, exists := s.GetProject()
		if p.ImportStatus == "failed" {
			return errors.New("failed to create the new repo: " + p.ImportError)
		}

		if !exists || p.ImportStatus != "finished" {
			s.log.With("import_status", p.ImportStatus).Info("waiting for project to finish")

			timeout := time.Duration(math.Exp2(float64(i+2))) * time.Second
			time.Sleep(timeout)
		}
	}

	return nil
}
func (s *ProjectCreator) CreateUsersSubGroup() (*gitlab.Group, error) {
	group, _, err := s.gitlab.Groups.GetGroup(s.gitlabData.Namespace, nil)
	if err != nil {
		return nil, err
	}

	subgroup, _, err := s.gitlab.Groups.CreateGroup(&gitlab.CreateGroupOptions{
		Name:     gitlab.Ptr(s.userName),
		Path:     gitlab.Ptr(s.userName),
		ParentID: gitlab.Ptr(group.ID),
	})
	if err != nil {
		return nil, err
	}

	time.Sleep(5 * time.Second)

	return subgroup, nil
}
func (s *ProjectCreator) RemoveBranchProtection(project *gitlab.Project) error {
	branches, _, err := s.gitlab.ProtectedBranches.ListProtectedBranches(project.ID, nil)
	if err != nil {
		return err
	}

	for _, b := range branches {
		_, _ = s.gitlab.ProtectedBranches.UnprotectRepositoryBranches(project.ID, b.Name)
	}

	return nil
}
