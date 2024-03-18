package web

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"gitlab-service/hive"

	"github.com/hashicorp/go-retryablehttp"
)

type Hive struct {
	hiveURL string
	token   string
	logger  *slog.Logger
	client  *retryablehttp.Client
}

func (h *Hive) Token() string {
	return h.token
}

func (h *Hive) SetToken(token string) {
	h.token = token
}

func New(hiveURL string, insecure bool, logger *slog.Logger) hive.Hive {
	client := retryablehttp.NewClient()
	client.Logger = logger

	client.RetryMax = 10
	client.HTTPClient.Transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: insecure}

	return &Hive{
		hiveURL: hiveURL,
		client:  client,
		token:   "",
		logger:  logger,
	}
}

type UpdateAssignmentRequest struct {
	Description string `json:"description"`
}

func (h *Hive) UpdateAssignment(assignmentID int, description string) {
	marshal, _ := json.Marshal(UpdateAssignmentRequest{
		Description: description,
	})

	payload := bytes.NewReader(marshal)

	method := "PATCH"
	url := fmt.Sprintf("%s/api/core/assignments/%d/", h.hiveURL, assignmentID)

	request, err := retryablehttp.NewRequest(method, url, payload)
	if err != nil {
		return
	}
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", h.token)

	res, err := h.client.Do(request)
	defer res.Body.Close()
	if err != nil {
		h.logger.Error(err.Error())
		return
	}

	if res.StatusCode != http.StatusOK {
		all, err := io.ReadAll(res.Body)
		if err != nil {
			return
		}
		h.logger.Error(string(all))
		return
	}

	return
}
