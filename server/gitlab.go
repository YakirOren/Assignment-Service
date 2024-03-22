package server

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/xanzy/go-gitlab"
)

func (s *Server) createRepoInGroup(group *gitlab.Group, gitlabData GitlabData) (*gitlab.Project, error) {
	path := strings.Replace(gitlabData.NewRepoName, " ", "_", -1)

	opt := &gitlab.ForkProjectOptions{
		MergeRequestDefaultTargetSelf: gitlab.Ptr(true),
		Name:                          gitlab.Ptr(gitlabData.NewRepoName),
		NamespaceID:                   gitlab.Ptr(group.ID),
		Path:                          gitlab.Ptr(path),
		Visibility:                    gitlab.Ptr(gitlab.PrivateVisibility),
	}

	project, _, err := s.gitlab.Projects.ForkProject(gitlabData.SourceRepo, opt)
	if err != nil {
		return nil, err
	}

	_, err = s.gitlab.Projects.DeleteProjectForkRelation(project.ID)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (s *Server) createUsersSubGroup(username string, namespace string) (*gitlab.Group, error) {
	group, _, err := s.gitlab.Groups.GetGroup(namespace, nil)
	if err != nil {
		return nil, err
	}

	opt := &gitlab.CreateGroupOptions{
		Name:     gitlab.Ptr(username),
		Path:     gitlab.Ptr(username),
		ParentID: gitlab.Ptr(group.ID),
	}

	subgroup, _, err := s.gitlab.Groups.CreateGroup(opt)
	if err != nil {
		return nil, err
	}

	return subgroup, nil
}

func (s *Server) listUsersByName(username string) (*gitlab.User, bool) {
	users, r, _ := s.gitlab.Users.ListUsers(&gitlab.ListUsersOptions{
		Username: gitlab.Ptr(username),
	})

	if r.StatusCode == http.StatusNotFound {
		return nil, false
	}

	if len(users) == 0 {
		return nil, false
	}

	return users[0], true
}

func (s *Server) listGroupsByName(groupName string, namespace string) (*gitlab.Group, bool) {
	groups, r, _ := s.gitlab.Groups.ListSubGroups(namespace, &gitlab.ListSubGroupsOptions{
		Search: gitlab.Ptr(groupName),
	})

	if r.StatusCode == http.StatusNotFound {
		return nil, false
	}

	if len(groups) == 0 {
		return nil, false
	}

	return groups[0], true
}

func (s *Server) addUserToProject(user *gitlab.User, group *gitlab.Project) error {
	opt := &gitlab.AddProjectMemberOptions{
		UserID:      user.ID,
		AccessLevel: gitlab.Ptr(gitlab.DeveloperPermissions),
	}

	_, _, err := s.gitlab.ProjectMembers.AddProjectMember(group.ID, opt)
	return err
}

func (s *Server) removeBranchProtection(project *gitlab.Project) error {
	branches, _, err := s.gitlab.ProtectedBranches.ListProtectedBranches(project.ID, nil)
	if err != nil {
		return err
	}

	for _, b := range branches {
		_, _ = s.gitlab.ProtectedBranches.UnprotectRepositoryBranches(project.ID, b.Name)
	}

	return nil
}

func (s *Server) buildRepoPath(body Request) string {
	data := body.OnCreationData.Gitlab
	repo := strings.Replace(data.NewRepoName, " ", "_", -1)
	return filepath.Join(data.Namespace, body.UserName, repo)
}

func (s *Server) projectExists(body Request) (*gitlab.Project, bool) {
	project, r, err := s.gitlab.Projects.GetProject(s.buildRepoPath(body), nil)
	if err != nil {
		return nil, false
	}

	if r.StatusCode == http.StatusNotFound {
		return nil, false
	}
	return project, true
}
