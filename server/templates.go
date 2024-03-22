package server

import (
	"bytes"
	"fmt"
	"html/template"
)

func (s *Server) UpdateAssignmentWithTemplate(assignmentID int, templateFile string, data any) error {
	description, err := RunTemplate(templateFile, data)
	if err != nil {
		return err
	}

	s.hive.UpdateAssignment(assignmentID, description)

	return nil
}

func RunTemplate(templateFile string, data any) (string, error) {
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return "", fmt.Errorf("failed to parse template file: %w", err)
	}

	buffer := &bytes.Buffer{}

	err = tmpl.Execute(buffer, data)
	if err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}

	return buffer.String(), nil
}
