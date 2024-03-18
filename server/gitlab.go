package server

import (
	"net/http"

	"github.com/xanzy/go-gitlab"
)

func (s *Server) CreateRepoInGroup(group *gitlab.Group, gitlabData GitlabData) (*gitlab.Project, error) {
	opt := &gitlab.ForkProjectOptions{
		MergeRequestDefaultTargetSelf: gitlab.Ptr(true),
		Name:                          gitlab.Ptr(gitlabData.NewRepoName),
		NamespaceID:                   gitlab.Ptr(group.ID),
		Path:                          gitlab.Ptr(gitlabData.NewRepoName),
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
