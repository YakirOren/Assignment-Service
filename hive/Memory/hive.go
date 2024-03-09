package Memory

import "gitlab-service/hive"

type InMemoryHive struct {
}

func (h *InMemoryHive) Token() string {
	return ""
}

func (h *InMemoryHive) SetToken(token string) {
}

func New() hive.Hive {
	return &InMemoryHive{}
}

func (*InMemoryHive) UpdateAssignment(assignmentID int, description string) error {
	return nil
}
