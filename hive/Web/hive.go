package Web

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"gitlab-service/hive"
	"net/http"
)

type Hive struct {
	hiveURL string
	token   string
	client  *retryablehttp.Client
}

func (h *Hive) Token() string {
	return h.token
}

func (h *Hive) SetToken(token string) {
	h.token = token
}

func New(hiveURL string) hive.Hive {
	// Configure the HTTP client.
	client := retryablehttp.NewClient()
	client.RetryMax = 10
	client.HTTPClient.Transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	return &Hive{
		hiveURL: hiveURL,
		client:  client,
		token:   "",
	}
}

type UpdateAssignmentRequest struct {
	Description string `json:"description"`
}

func (h *Hive) UpdateAssignment(assignmentID int, description string) error {
	marshal, _ := json.Marshal(UpdateAssignmentRequest{
		Description: description,
	})

	payload := bytes.NewReader(marshal)

	method := "PATCH"
	url := fmt.Sprintf("%s/api/core/assignments/%d/", h.hiveURL, assignmentID)

	request, err := retryablehttp.NewRequest(method, url, payload)
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", h.token)

	res, err := h.client.Do(request)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.New("failed to update description")
	}

	return nil
}
