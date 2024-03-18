package hive

type Hive interface {
	SetToken(token string)
	Token() string
	UpdateAssignment(assignmentID int, description string)
}
