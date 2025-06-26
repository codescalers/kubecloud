package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
)

const (
	BaseURL = "http://localhost:8080/api/v1"
)

type Client struct {
	httpClient  *http.Client
	accessToken string
	baseURL     string
}

// Request/Response structs
type RegisterRequest struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         struct {
		ID    uint   `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user"`
}

type DeploymentRequest struct {
	BlueprintID string                 `json:"blueprint_id"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type DeploymentResponse struct {
	TaskID    string    `json:"task_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type DeployResponse struct {
	TaskID    string    `json:"task_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type TaskStatusResponse struct {
	TaskID      string                 `json:"task_id"`
	Status      string                 `json:"status"`
	Message     string                 `json:"message"`
	BlueprintID string                 `json:"blueprint_id"`
	Parameters  map[string]interface{} `json:"parameters"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

type Blueprint struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    BaseURL,
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
	req := RegisterRequest{
		Name:            name,
		Email:           email,
		Password:        password,
		ConfirmPassword: confirmPassword,
	}

	resp, err := c.makeRequest("POST", "/register", req, false)
	if err != nil {
		return fmt.Errorf("registration request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Println("✅ Registration successful")
	return nil
}

func (c *Client) Login(email, password string) error {
	req := LoginRequest{
		Email:    email,
		Password: password,
	}

	resp, err := c.makeRequest("POST", "/login", req, false)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	c.accessToken = loginResp.AccessToken
	fmt.Printf("✅ Login successful. Welcome, %s!\n", loginResp.User.Name)
	return nil
}

func (c *Client) DeployCluster(clusterName string) error {
	cluster := workloads.K8sCluster{
		Master: &workloads.K8sNode{
			VM: &workloads.VM{
				Name:        "k8s_master_1",
				NodeID:      1,
				CPU:         2,
				MemoryMB:    2048,
				NetworkName: "test",
				IP:          "10.1.0.2",
				Flist:       "https://hub.grid.tf/tf-official-apps/base:latest.flist",
				Entrypoint:  "/sbin/zinit init",
				EnvVars: map[string]string{
					"K8S_TOKEN": "abc123def456",
				},
			},
		},
		Workers: []workloads.K8sNode{
			{
				VM: &workloads.VM{
					NodeID:      2,
					CPU:         2,
					MemoryMB:    2048,
					NetworkName: "test",
					IP:          "10.1.0.3",
					Flist:       "https://hub.grid.tf/tf-official-apps/base:latest.flist",
					Entrypoint:  "/sbin/zinit init",
					EnvVars: map[string]string{
						"K8S_TOKEN": "abc123def456",
					},
				},
			},
			{
				VM: &workloads.VM{
					NodeID:      3,
					CPU:         2,
					MemoryMB:    2048,
					NetworkName: "test",
					IP:          "10.1.0.4",
					Flist:       "https://hub.grid.tf/tf-official-apps/base:latest.flist",
					Entrypoint:  "/sbin/zinit init",
				},
			},
		},
		Token:       "abc123def456",
		NetworkName: "test",
	}

	resp, err := c.makeRequest("POST", "/deploy", cluster, true)
	if err != nil {
		return fmt.Errorf("deploy request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("deploy request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var deployResp DeployResponse
	if err := json.NewDecoder(resp.Body).Decode(&deployResp); err != nil {
		return fmt.Errorf("failed to decode deploy response: %w", err)
	}

	fmt.Printf("✅ Cluster deployment initiated successfully!\n")
	fmt.Printf("   Task ID: %s\n", deployResp.TaskID)
	fmt.Printf("   Cluster: %s\n", clusterName)
	fmt.Printf("   Master Node: %s (Node ID: %d)\n", cluster.Master.Name, cluster.Master.NodeID)
	fmt.Printf("   Worker Nodes: %d\n", len(cluster.Workers))
	for _, worker := range cluster.Workers {
		fmt.Printf("     - %s (Node ID: %d)\n", worker.Name, worker.NodeID)
	}
	fmt.Printf("   Network: %s\n", cluster.NetworkName)

	return nil
}

func (c *Client) GetTaskStatus(taskID string) (*TaskStatusResponse, error) {
	resp, err := c.makeRequest("GET", "/tasks/"+taskID, nil, true)
	if err != nil {
		return nil, fmt.Errorf("get task status request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get task status failed with status %d: %s", resp.StatusCode, string(body))
	}

	var taskResp TaskStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		return nil, fmt.Errorf("failed to decode task status response: %w", err)
	}

	return &taskResp, nil
}

func (c *Client) ListTasks() ([]TaskStatusResponse, error) {
	resp, err := c.makeRequest("GET", "/tasks", nil, true)
	if err != nil {
		return nil, fmt.Errorf("list tasks request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list tasks failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Tasks []TaskStatusResponse `json:"tasks"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode tasks response: %w", err)
	}

	return response.Tasks, nil
}

func (c *Client) ListenToSSE(taskID string) error {
	req, err := http.NewRequest("GET", c.baseURL+"/events", nil)
	if err != nil {
		return fmt.Errorf("failed to create SSE request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("SSE request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SSE connection failed with status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("🔄 Listening for real-time updates for task %s...\n", taskID)
	fmt.Println("Press Ctrl+C to stop listening")

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data != "" {
				fmt.Printf("📡 SSE Update: %s\n", data)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading SSE stream: %w", err)
	}

	return nil
}

func main() {
	client := NewClient()

	fmt.Println("🚀 KubeCloud Async Deployment Test Client")
	fmt.Println("=========================================")

	testEmail := "omar@example.com"
	testPassword := "password"

	fmt.Println("\n2. Logging in...")
	if err := client.Login(testEmail, testPassword); err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	fmt.Println("\n4. Deploying k8s cluster...")
	clusterName := "my_k8s_cluster"
	if err := client.DeployCluster(clusterName); err != nil {
		log.Fatalf("Deployment failed: %v", err)
	}

	fmt.Println("\n🏁 Test client finished!")
}
