package devops

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client is an Azure DevOps REST API client
type Client struct {
	organization string
	project      string
	pat          string
	apiVersion   string
	httpClient   *http.Client
	baseURL      string
}

// NewClient creates a new Azure DevOps client
func NewClient(organization, project, pat, apiVersion string) *Client {
	return &Client{
		organization: organization,
		project:      project,
		pat:          pat,
		apiVersion:   apiVersion,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: fmt.Sprintf("https://dev.azure.com/%s/%s", organization, project),
	}
}

// ========================================
// Work Items
// ========================================

// WorkItem represents an Azure DevOps work item
type WorkItem struct {
	ID     int                    `json:"id"`
	Rev    int                    `json:"rev"`
	Fields map[string]interface{} `json:"fields"`
	URL    string                 `json:"url"`
}

// WorkItemCreateRequest represents a work item creation request
type WorkItemCreateRequest struct {
	Type        string            // Bug, Task, User Story, etc.
	Title       string
	Description string
	AssignedTo  string
	State       string
	Priority    int
	Tags        []string
	ParentID    int // Optional parent work item ID
	CustomFields map[string]interface{}
}

// WorkItemUpdateRequest represents a work item update request
type WorkItemUpdateRequest struct {
	Title       *string
	Description *string
	State       *string
	AssignedTo  *string
	Priority    *int
	Tags        []string
	CustomFields map[string]interface{}
}

// WorkItemQueryResult represents WIQL query results
type WorkItemQueryResult struct {
	QueryType        string `json:"queryType"`
	QueryResultType  string `json:"queryResultType"`
	AsOf             string `json:"asOf"`
	WorkItems        []WorkItemRef `json:"workItems"`
}

// WorkItemRef is a reference to a work item
type WorkItemRef struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

// GetWorkItem retrieves a work item by ID
func (c *Client) GetWorkItem(ctx context.Context, id int) (*WorkItem, error) {
	endpoint := fmt.Sprintf("%s/_apis/wit/workitems/%d?api-version=%s&$expand=all",
		c.baseURL, id, c.apiVersion)

	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var wi WorkItem
	if err := json.NewDecoder(resp.Body).Decode(&wi); err != nil {
		return nil, fmt.Errorf("failed to decode work item: %w", err)
	}

	return &wi, nil
}

// QueryWorkItems executes a WIQL query
func (c *Client) QueryWorkItems(ctx context.Context, query string) ([]WorkItem, error) {
	endpoint := fmt.Sprintf("%s/_apis/wit/wiql?api-version=%s", c.baseURL, c.apiVersion)

	body := map[string]string{"query": query}
	jsonBody, _ := json.Marshal(body)

	resp, err := c.doRequest(ctx, "POST", endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result WorkItemQueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode query result: %w", err)
	}

	if len(result.WorkItems) == 0 {
		return []WorkItem{}, nil
	}

	// Get full work item details
	return c.GetWorkItemsBatch(ctx, result.WorkItems)
}

// GetWorkItemsBatch retrieves multiple work items by ID
func (c *Client) GetWorkItemsBatch(ctx context.Context, refs []WorkItemRef) ([]WorkItem, error) {
	ids := make([]int, len(refs))
	for i, ref := range refs {
		ids[i] = ref.ID
	}

	endpoint := fmt.Sprintf("%s/_apis/wit/workitemsbatch?api-version=%s", c.baseURL, c.apiVersion)

	body := map[string]interface{}{
		"ids":    ids,
		"fields": []string{
			"System.Id",
			"System.Title",
			"System.State",
			"System.AssignedTo",
			"System.WorkItemType",
			"System.Description",
			"System.CreatedDate",
			"System.ChangedDate",
			"Microsoft.VSTS.Common.Priority",
			"System.Tags",
		},
	}
	jsonBody, _ := json.Marshal(body)

	resp, err := c.doRequest(ctx, "POST", endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Count int        `json:"count"`
		Value []WorkItem `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode batch result: %w", err)
	}

	return result.Value, nil
}

// CreateWorkItem creates a new work item
func (c *Client) CreateWorkItem(ctx context.Context, req WorkItemCreateRequest) (*WorkItem, error) {
	endpoint := fmt.Sprintf("%s/_apis/wit/workitems/$%s?api-version=%s",
		c.baseURL, url.PathEscape(req.Type), c.apiVersion)

	// Build JSON Patch document
	ops := []map[string]interface{}{
		{"op": "add", "path": "/fields/System.Title", "value": req.Title},
	}

	if req.Description != "" {
		ops = append(ops, map[string]interface{}{
			"op": "add", "path": "/fields/System.Description", "value": req.Description,
		})
	}

	if req.AssignedTo != "" {
		ops = append(ops, map[string]interface{}{
			"op": "add", "path": "/fields/System.AssignedTo", "value": req.AssignedTo,
		})
	}

	if req.State != "" {
		ops = append(ops, map[string]interface{}{
			"op": "add", "path": "/fields/System.State", "value": req.State,
		})
	}

	if req.Priority > 0 {
		ops = append(ops, map[string]interface{}{
			"op": "add", "path": "/fields/Microsoft.VSTS.Common.Priority", "value": req.Priority,
		})
	}

	if len(req.Tags) > 0 {
		ops = append(ops, map[string]interface{}{
			"op": "add", "path": "/fields/System.Tags", "value": joinTags(req.Tags),
		})
	}

	if req.ParentID > 0 {
		ops = append(ops, map[string]interface{}{
			"op":    "add",
			"path":  "/relations/-",
			"value": map[string]interface{}{
				"rel": "System.LinkTypes.Hierarchy-Reverse",
				"url": fmt.Sprintf("%s/_apis/wit/workitems/%d", c.baseURL, req.ParentID),
			},
		})
	}

	for path, value := range req.CustomFields {
		ops = append(ops, map[string]interface{}{
			"op": "add", "path": "/fields/" + path, "value": value,
		})
	}

	jsonBody, _ := json.Marshal(ops)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json-patch+json")
	httpReq.Header.Set("Authorization", "Basic "+c.basicAuth())

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var wi WorkItem
	if err := json.NewDecoder(resp.Body).Decode(&wi); err != nil {
		return nil, fmt.Errorf("failed to decode work item: %w", err)
	}

	return &wi, nil
}

// UpdateWorkItem updates an existing work item
func (c *Client) UpdateWorkItem(ctx context.Context, id int, req WorkItemUpdateRequest) (*WorkItem, error) {
	endpoint := fmt.Sprintf("%s/_apis/wit/workitems/%d?api-version=%s", c.baseURL, id, c.apiVersion)

	var ops []map[string]interface{}

	if req.Title != nil {
		ops = append(ops, map[string]interface{}{
			"op": "add", "path": "/fields/System.Title", "value": *req.Title,
		})
	}

	if req.Description != nil {
		ops = append(ops, map[string]interface{}{
			"op": "add", "path": "/fields/System.Description", "value": *req.Description,
		})
	}

	if req.State != nil {
		ops = append(ops, map[string]interface{}{
			"op": "add", "path": "/fields/System.State", "value": *req.State,
		})
	}

	if req.AssignedTo != nil {
		ops = append(ops, map[string]interface{}{
			"op": "add", "path": "/fields/System.AssignedTo", "value": *req.AssignedTo,
		})
	}

	if req.Priority != nil {
		ops = append(ops, map[string]interface{}{
			"op": "add", "path": "/fields/Microsoft.VSTS.Common.Priority", "value": *req.Priority,
		})
	}

	if len(req.Tags) > 0 {
		ops = append(ops, map[string]interface{}{
			"op": "add", "path": "/fields/System.Tags", "value": joinTags(req.Tags),
		})
	}

	for path, value := range req.CustomFields {
		ops = append(ops, map[string]interface{}{
			"op": "add", "path": "/fields/" + path, "value": value,
		})
	}

	jsonBody, _ := json.Marshal(ops)

	httpReq, err := http.NewRequestWithContext(ctx, "PATCH", endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json-patch+json")
	httpReq.Header.Set("Authorization", "Basic "+c.basicAuth())

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var wi WorkItem
	if err := json.NewDecoder(resp.Body).Decode(&wi); err != nil {
		return nil, fmt.Errorf("failed to decode work item: %w", err)
	}

	return &wi, nil
}

// GetMyWorkItems returns work items assigned to the authenticated user
func (c *Client) GetMyWorkItems(ctx context.Context) ([]WorkItem, error) {
	query := `SELECT [System.Id], [System.Title], [System.State], [System.AssignedTo], [System.WorkItemType]
              FROM WorkItems 
              WHERE [System.AssignedTo] = @Me 
              AND [System.State] <> 'Closed'
              AND [System.State] <> 'Done'
              ORDER BY [System.ChangedDate] DESC`
	
	return c.QueryWorkItems(ctx, query)
}

// GetRecentWorkItems returns recently changed work items
func (c *Client) GetRecentWorkItems(ctx context.Context, days int) ([]WorkItem, error) {
	query := fmt.Sprintf(`SELECT [System.Id], [System.Title], [System.State], [System.AssignedTo], [System.WorkItemType]
              FROM WorkItems 
              WHERE [System.ChangedDate] >= @Today - %d
              ORDER BY [System.ChangedDate] DESC`, days)
	
	return c.QueryWorkItems(ctx, query)
}

// ========================================
// Pipelines
// ========================================

// Pipeline represents an Azure DevOps pipeline
type Pipeline struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Folder   string `json:"folder"`
	Revision int    `json:"revision"`
	URL      string `json:"url"`
}

// PipelineRun represents a pipeline run
type PipelineRun struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	State      string `json:"state"`
	Result     string `json:"result"`
	CreatedDate string `json:"createdDate"`
	FinishedDate string `json:"finishedDate,omitempty"`
	URL        string `json:"url"`
	Pipeline   struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"pipeline"`
}

// ListPipelines lists all pipelines
func (c *Client) ListPipelines(ctx context.Context) ([]Pipeline, error) {
	endpoint := fmt.Sprintf("%s/_apis/pipelines?api-version=%s", c.baseURL, c.apiVersion)

	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Count int        `json:"count"`
		Value []Pipeline `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode pipelines: %w", err)
	}

	return result.Value, nil
}

// RunPipeline triggers a pipeline run
func (c *Client) RunPipeline(ctx context.Context, pipelineID int, branch string, variables map[string]string) (*PipelineRun, error) {
	endpoint := fmt.Sprintf("%s/_apis/pipelines/%d/runs?api-version=%s", c.baseURL, pipelineID, c.apiVersion)

	body := map[string]interface{}{
		"resources": map[string]interface{}{
			"repositories": map[string]interface{}{
				"self": map[string]interface{}{
					"refName": branch,
				},
			},
		},
	}

	if len(variables) > 0 {
		vars := make(map[string]interface{})
		for k, v := range variables {
			vars[k] = map[string]interface{}{"value": v}
		}
		body["variables"] = vars
	}

	jsonBody, _ := json.Marshal(body)

	resp, err := c.doRequest(ctx, "POST", endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var run PipelineRun
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, fmt.Errorf("failed to decode pipeline run: %w", err)
	}

	return &run, nil
}

// GetPipelineRuns gets recent runs for a pipeline
func (c *Client) GetPipelineRuns(ctx context.Context, pipelineID int, top int) ([]PipelineRun, error) {
	endpoint := fmt.Sprintf("%s/_apis/pipelines/%d/runs?api-version=%s&$top=%d",
		c.baseURL, pipelineID, c.apiVersion, top)

	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Count int           `json:"count"`
		Value []PipelineRun `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode pipeline runs: %w", err)
	}

	return result.Value, nil
}

// ========================================
// Repositories
// ========================================

// Repository represents a Git repository
type Repository struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	URL           string `json:"url"`
	DefaultBranch string `json:"defaultBranch"`
	Size          int64  `json:"size"`
	WebURL        string `json:"webUrl"`
}

// ListRepositories lists all repositories
func (c *Client) ListRepositories(ctx context.Context) ([]Repository, error) {
	endpoint := fmt.Sprintf("%s/_apis/git/repositories?api-version=%s", c.baseURL, c.apiVersion)

	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Count int          `json:"count"`
		Value []Repository `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode repositories: %w", err)
	}

	return result.Value, nil
}

// ========================================
// Boards
// ========================================

// Board represents a board
type Board struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// BoardColumn represents a board column
type BoardColumn struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ItemLimit   int    `json:"itemLimit"`
	StateMappings map[string]string `json:"stateMappings"`
	IsSplit     bool   `json:"isSplit"`
	Description string `json:"description"`
	ColumnType  string `json:"columnType"`
}

// ListBoards lists all boards for the project
func (c *Client) ListBoards(ctx context.Context, team string) ([]Board, error) {
	if team == "" {
		team = c.project + " Team"
	}
	
	endpoint := fmt.Sprintf("https://dev.azure.com/%s/%s/%s/_apis/work/boards?api-version=%s",
		c.organization, c.project, url.PathEscape(team), c.apiVersion)

	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Count int     `json:"count"`
		Value []Board `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode boards: %w", err)
	}

	return result.Value, nil
}

// GetBoardColumns gets columns for a board
func (c *Client) GetBoardColumns(ctx context.Context, team, boardName string) ([]BoardColumn, error) {
	if team == "" {
		team = c.project + " Team"
	}
	
	endpoint := fmt.Sprintf("https://dev.azure.com/%s/%s/%s/_apis/work/boards/%s/columns?api-version=%s",
		c.organization, c.project, url.PathEscape(team), url.PathEscape(boardName), c.apiVersion)

	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Count int           `json:"count"`
		Value []BoardColumn `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode board columns: %w", err)
	}

	return result.Value, nil
}

// ========================================
// Helpers
// ========================================

func (c *Client) doRequest(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+c.basicAuth())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

func (c *Client) basicAuth() string {
	auth := ":" + c.pat
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func joinTags(tags []string) string {
	result := ""
	for i, tag := range tags {
		if i > 0 {
			result += "; "
		}
		result += tag
	}
	return result
}
