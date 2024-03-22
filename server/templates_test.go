package server

import (
	"fmt"
	"testing"

	"github.com/xanzy/go-gitlab"
)

func TestRunTemplate(t *testing.T) {
	a := struct {
		UserName string
		Project  *gitlab.Project
	}{"student-0", &gitlab.Project{
		ID:                0,
		HTTPURLToRepo:     "https://bislab/students/python/student-0/aaa_bbb",
		WebURL:            "",
		Name:              "aaa_bbb",
		NameWithNamespace: "",
		Path:              "",
		PathWithNamespace: "",
	}}
	got, err := RunTemplate("/Users/yakiroren/Documents/gitlab-service/templates/git.md", a)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(got)
}
