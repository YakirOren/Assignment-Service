package server

import (
	"github.com/xanzy/go-gitlab"
	"net/http"
)

func (s *Server) CreateRepoInGroup(group *gitlab.Group, gitlabData GitlabData) (*gitlab.Project, error) {
	opt := &gitlab.ForkProjectOptions{
		MergeRequestDefaultTargetSelf: gitlab.Ptr(true),
		Name:                          gitlab.Ptr(gitlabData.NewRepoName),
		NamespaceID:                   gitlab.Ptr(group.ID),
		Visibility:                    gitlab.Ptr(gitlab.PrivateVisibility),
	}

	project, _, err := s.gitlab.Projects.ForkProject(gitlabData.SourceRepo, opt)
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

func (s *Server) ListGroups(groupName string, namespace string) ([]*gitlab.Group, bool) {
	groups, r, _ := s.gitlab.Groups.ListSubGroups(namespace, &gitlab.ListSubGroupsOptions{
		Search: gitlab.Ptr(groupName),
	})

	if r.StatusCode == http.StatusNotFound {
		return nil, false
	}

	if len(groups) == 0 {
		return nil, false
	}

	return groups, true
}

func (s *Server) CreateGroup(groupName string, path string) (*gitlab.Group, error) {
	opt := &gitlab.CreateGroupOptions{
		Name: gitlab.Ptr(groupName),
		Path: gitlab.Ptr(path),
	}

	group, _, err := s.gitlab.Groups.CreateGroup(opt)
	if err != nil {
		return nil, err
	}

	return group, nil
}
