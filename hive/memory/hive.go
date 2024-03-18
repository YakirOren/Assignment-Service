package memory

import "gitlab-service/hive"

type InMemoryHive struct{}

func (h *InMemoryHive) Token() string {
	return ""
}

func (h *InMemoryHive) SetToken(_ string) {
}

func New() hive.Hive {
	return &InMemoryHive{}
}

func (*InMemoryHive) UpdateAssignment(assignmentID int, description string) {
	return
}
