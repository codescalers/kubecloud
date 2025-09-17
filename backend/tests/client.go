//go:build example

package tests

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"kubecloud/app"
	"kubecloud/kubedeployer"
)

type Client struct {
	httpClient  *http.Client
	accessToken string
	baseURL     string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "http://localhost:8080/api/v1",
	}
}

func (c *Client) makeRequest(method, endpoint string, body interface{}, needsAuth bool) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if needsAuth && c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	return c.httpClient.Do(req)
}

func (c *Client) Register(name, email, password, confirmPassword string) error {
	req := app.RegisterInput{
		Name:            name,
		Email:           email,
		Password:        password,
		ConfirmPassword: confirmPassword,
	}

	resp, err := c.makeRequest("POST", "/user/register", req, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Client) Login(email, password string) error {
	req := app.LoginInput{
		Email:    email,
		Password: password,
	}

	resp, err := c.makeRequest("POST", "/user/login", req, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
		Data    struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	if loginResp.Data.AccessToken == "" {
		return fmt.Errorf("no access token received in login response")
	}

	c.accessToken = loginResp.Data.AccessToken
	return nil
}

func (c *Client) DeployCluster(cluster kubedeployer.Cluster) (string, error) {
	resp, err := c.makeRequest("POST", "/deployments", cluster, true)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("deploy request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var deployResp app.Response
	if err := json.NewDecoder(resp.Body).Decode(&deployResp); err != nil {
		return "", err
	}

	return deployResp.WorkflowID, nil
}

func (c *Client) ListenToSSE(taskID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/events", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	sseClient := &http.Client{Timeout: 0}

	resp, err := sseClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SSE connection failed with status %d: %s", resp.StatusCode, string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					if ctx.Err() != nil {
						return nil
					}
					return err
				}
				return nil
			}

			line := scanner.Text()
			if strings.HasPrefix(line, "data:") {
				data := strings.TrimPrefix(line, "data:")
				if data != "" {
					fmt.Printf("SSE Update: %s\n", data)
				}
			}
		}
	}
}

func (c *Client) ListDeployments() ([]interface{}, error) {
	resp, err := c.makeRequest("GET", "/deployments", nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list deployments failed with status %d: %s", resp.StatusCode, string(body))
	}

	var listResp struct {
		Deployments []interface{} `json:"deployments"`
		Count       int           `json:"count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("failed to decode list response: %w", err)
	}

	return listResp.Deployments, nil
}

func (c *Client) GetDeployment(name string) (interface{}, error) {
	resp, err := c.makeRequest("GET", "/deployments/"+name, nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get deployment failed with status %d: %s", resp.StatusCode, string(body))
	}

	var deployment interface{}
	if err := json.NewDecoder(resp.Body).Decode(&deployment); err != nil {
		return nil, fmt.Errorf("failed to decode deployment response: %w", err)
	}

	return deployment, nil
}

func (c *Client) GetKubeconfig(name string) (string, error) {
	resp, err := c.makeRequest("GET", "/deployments/"+name+"/kubeconfig", nil, true)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("get kubeconfig failed with status %d: %s", resp.StatusCode, string(body))
	}

	kubeconfig, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read kubeconfig response: %w", err)
	}

	return string(kubeconfig), nil
}

func (c *Client) DeleteCluster(name string) error {
	resp, err := c.makeRequest("DELETE", "/deployments/"+name, nil, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete deployment failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Client) DeleteAllDeployments() (string, error) {
	resp, err := c.makeRequest("DELETE", "/deployments", nil, true)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("delete all deployments failed with status %d: %s", resp.StatusCode, string(body))
	}

	var deleteResp app.Response
	if err := json.NewDecoder(resp.Body).Decode(&deleteResp); err != nil {
		return "", fmt.Errorf("failed to decode delete all response: %w", err)
	}

	return deleteResp.WorkflowID, nil
}

func (c *Client) AddNode(deploymentName string, node kubedeployer.Node) (string, error) {
	cluster := kubedeployer.Cluster{
		Name:  deploymentName,
		Nodes: []kubedeployer.Node{node},
	}

	resp, err := c.makeRequest("POST", "/deployments/"+deploymentName+"/nodes", cluster, true)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("add node request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var addNodeResp struct {
		TaskID string `json:"task_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&addNodeResp); err != nil {
		return "", err
	}

	return addNodeResp.TaskID, nil
}

func (c *Client) RemoveNode(deploymentName, nodeName string) error {
	resp, err := c.makeRequest("DELETE", "/deployments/"+deploymentName+"/nodes/"+nodeName, nil, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("remove node request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// NotificationsResponse represents the response structure for notification lists
type NotificationsResponse struct {
	Notifications []app.NotificationResponse `json:"notifications"`
	Limit         int                        `json:"limit"`
	Offset        int                        `json:"offset"`
	Count         int                        `json:"count"`
}

// APIResponseWrapper represents the standard API response structure
type APIResponseWrapper struct {
	Status  int                   `json:"status"`
	Message string                `json:"message"`
	Data    NotificationsResponse `json:"data"`
}

// GetAllNotifications retrieves all notifications with optional pagination
func (c *Client) GetAllNotifications(limit, offset int) (*NotificationsResponse, error) {
	endpoint := "/notifications"
	if limit > 0 || offset > 0 {
		endpoint += fmt.Sprintf("?limit=%d&offset=%d", limit, offset)
	}

	resp, err := c.makeRequest("GET", endpoint, nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get all notifications failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Decode the API response wrapper first
	var apiResp APIResponseWrapper

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	return &apiResp.Data, nil
}

// GetUnreadNotifications retrieves only unread notifications with optional pagination
func (c *Client) GetUnreadNotifications(limit, offset int) (*NotificationsResponse, error) {
	endpoint := fmt.Sprintf("/notifications/unread?limit=%d&offset=%d", limit, offset)
	resp, err := c.makeRequest("GET", endpoint, nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get unread notifications failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Decode the API response wrapper first
	var apiResp APIResponseWrapper

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	return &apiResp.Data, nil
}

// MarkNotificationRead marks a specific notification as read
func (c *Client) MarkNotificationRead(notificationID string) error {
	endpoint := fmt.Sprintf("/notifications/%s/read", notificationID)

	resp, err := c.makeRequest("PUT", endpoint, nil, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mark notification read failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// MarkNotificationUnread marks a specific notification as unread
func (c *Client) MarkNotificationUnread(notificationID string) error {
	endpoint := fmt.Sprintf("/notifications/%s/unread", notificationID)

	resp, err := c.makeRequest("PUT", endpoint, nil, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mark notification unread failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// MarkAllNotificationsRead marks all notifications as read for the authenticated user
func (c *Client) MarkAllNotificationsRead() error {
	resp, err := c.makeRequest("PUT", "/notifications/read-all", nil, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mark all notifications read failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteNotification deletes a specific notification
func (c *Client) DeleteNotification(notificationID string) error {
	endpoint := fmt.Sprintf("/notifications/%s", notificationID)

	resp, err := c.makeRequest("DELETE", endpoint, nil, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete notification failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteAllNotifications deletes all notifications for the authenticated user
func (c *Client) DeleteAllNotifications() error {
	resp, err := c.makeRequest("DELETE", "/notifications", nil, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete all notifications failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
