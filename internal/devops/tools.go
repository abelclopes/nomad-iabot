package devops

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/abelclopes/nomad-iabot/internal/llm"
)

// Tool represents an Azure DevOps tool for the LLM
type Tool struct {
	client *Client
}

// NewTool creates a new DevOps tool
func NewTool(client *Client) *Tool {
	return &Tool{client: client}
}

// GetToolDefinitions returns the tool definitions for the LLM
func (t *Tool) GetToolDefinitions() []llm.Tool {
	return []llm.Tool{
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "devops_list_my_workitems",
				Description: "List Azure DevOps work items assigned to me (current user). Returns tasks, bugs, and stories that are not closed.",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
					"required":   []string{},
				},
			},
		},
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "devops_get_workitem",
				Description: "Get details of a specific Azure DevOps work item by ID",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":        "integer",
							"description": "The work item ID",
						},
					},
					"required": []string{"id"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "devops_create_workitem",
				Description: "Create a new Azure DevOps work item (Task, Bug, User Story, etc.)",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"type": map[string]interface{}{
							"type":        "string",
							"description": "Work item type: Task, Bug, User Story, Feature, Epic",
							"enum":        []string{"Task", "Bug", "User Story", "Feature", "Epic"},
						},
						"title": map[string]interface{}{
							"type":        "string",
							"description": "Title of the work item",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Description of the work item (HTML supported)",
						},
						"assigned_to": map[string]interface{}{
							"type":        "string",
							"description": "Email or display name of the assignee",
						},
						"priority": map[string]interface{}{
							"type":        "integer",
							"description": "Priority (1=highest, 4=lowest)",
							"enum":        []int{1, 2, 3, 4},
						},
						"tags": map[string]interface{}{
							"type":        "array",
							"items":       map[string]interface{}{"type": "string"},
							"description": "Tags to add to the work item",
						},
						"parent_id": map[string]interface{}{
							"type":        "integer",
							"description": "Parent work item ID (for hierarchy)",
						},
					},
					"required": []string{"type", "title"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "devops_update_workitem",
				Description: "Update an existing Azure DevOps work item",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":        "integer",
							"description": "The work item ID to update",
						},
						"title": map[string]interface{}{
							"type":        "string",
							"description": "New title",
						},
						"state": map[string]interface{}{
							"type":        "string",
							"description": "New state: New, Active, Resolved, Closed, etc.",
						},
						"assigned_to": map[string]interface{}{
							"type":        "string",
							"description": "New assignee email or display name",
						},
						"priority": map[string]interface{}{
							"type":        "integer",
							"description": "New priority (1-4)",
						},
					},
					"required": []string{"id"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "devops_query_workitems",
				Description: "Query Azure DevOps work items using WIQL (Work Item Query Language)",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "WIQL query string. Example: SELECT [System.Id], [System.Title] FROM WorkItems WHERE [System.State] = 'Active'",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "devops_list_pipelines",
				Description: "List all Azure DevOps pipelines in the project",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
					"required":   []string{},
				},
			},
		},
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "devops_run_pipeline",
				Description: "Trigger an Azure DevOps pipeline run",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"pipeline_id": map[string]interface{}{
							"type":        "integer",
							"description": "The pipeline ID to run",
						},
						"branch": map[string]interface{}{
							"type":        "string",
							"description": "Git branch to run the pipeline on (e.g., refs/heads/main)",
							"default":     "refs/heads/main",
						},
						"variables": map[string]interface{}{
							"type":        "object",
							"description": "Pipeline variables as key-value pairs",
						},
					},
					"required": []string{"pipeline_id"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "devops_list_repos",
				Description: "List all Git repositories in the Azure DevOps project",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
					"required":   []string{},
				},
			},
		},
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "devops_list_boards",
				Description: "List all boards (Kanban) in the Azure DevOps project",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"team": map[string]interface{}{
							"type":        "string",
							"description": "Team name (optional, defaults to project default team)",
						},
					},
					"required": []string{},
				},
			},
		},
	}
}

// Execute executes a DevOps tool call - returns (result, handled, error)
func (t *Tool) Execute(ctx context.Context, name string, args map[string]interface{}) (string, bool, error) {
	switch name {
	case "devops_list_my_workitems":
		result, err := t.listMyWorkItems(ctx)
		return result, true, err
	case "devops_get_workitem":
		result, err := t.getWorkItem(ctx, args)
		return result, true, err
	case "devops_create_workitem":
		result, err := t.createWorkItem(ctx, args)
		return result, true, err
	case "devops_update_workitem":
		result, err := t.updateWorkItem(ctx, args)
		return result, true, err
	case "devops_query_workitems":
		result, err := t.queryWorkItems(ctx, args)
		return result, true, err
	case "devops_list_pipelines":
		result, err := t.listPipelines(ctx)
		return result, true, err
	case "devops_run_pipeline":
		result, err := t.runPipeline(ctx, args)
		return result, true, err
	case "devops_list_repos":
		result, err := t.listRepos(ctx)
		return result, true, err
	case "devops_list_boards":
		result, err := t.listBoards(ctx, args)
		return result, true, err
	default:
		return "", false, nil
	}
}

// ExecuteTool executes a DevOps tool call (legacy)
func (t *Tool) ExecuteTool(ctx context.Context, name string, arguments string) (string, error) {
	var args map[string]interface{}
	if arguments != "" {
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return "", fmt.Errorf("failed to parse arguments: %w", err)
		}
	}

	switch name {
	case "devops_list_my_workitems":
		return t.listMyWorkItems(ctx)
	case "devops_get_workitem":
		return t.getWorkItem(ctx, args)
	case "devops_create_workitem":
		return t.createWorkItem(ctx, args)
	case "devops_update_workitem":
		return t.updateWorkItem(ctx, args)
	case "devops_query_workitems":
		return t.queryWorkItems(ctx, args)
	case "devops_list_pipelines":
		return t.listPipelines(ctx)
	case "devops_run_pipeline":
		return t.runPipeline(ctx, args)
	case "devops_list_repos":
		return t.listRepos(ctx)
	case "devops_list_boards":
		return t.listBoards(ctx, args)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (t *Tool) listMyWorkItems(ctx context.Context) (string, error) {
	items, err := t.client.GetMyWorkItems(ctx)
	if err != nil {
		return "", err
	}
	return formatWorkItems(items), nil
}

func (t *Tool) getWorkItem(ctx context.Context, args map[string]interface{}) (string, error) {
	id, ok := args["id"].(float64)
	if !ok {
		return "", fmt.Errorf("id is required")
	}

	item, err := t.client.GetWorkItem(ctx, int(id))
	if err != nil {
		return "", err
	}
	return formatWorkItem(item), nil
}

func (t *Tool) createWorkItem(ctx context.Context, args map[string]interface{}) (string, error) {
	req := WorkItemCreateRequest{
		Type:  getString(args, "type"),
		Title: getString(args, "title"),
	}

	if desc := getString(args, "description"); desc != "" {
		req.Description = desc
	}
	if assigned := getString(args, "assigned_to"); assigned != "" {
		req.AssignedTo = assigned
	}
	if priority, ok := args["priority"].(float64); ok {
		req.Priority = int(priority)
	}
	if parentID, ok := args["parent_id"].(float64); ok {
		req.ParentID = int(parentID)
	}
	if tags, ok := args["tags"].([]interface{}); ok {
		for _, tag := range tags {
			if s, ok := tag.(string); ok {
				req.Tags = append(req.Tags, s)
			}
		}
	}

	item, err := t.client.CreateWorkItem(ctx, req)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Created work item #%d: %s", item.ID, item.Fields["System.Title"]), nil
}

func (t *Tool) updateWorkItem(ctx context.Context, args map[string]interface{}) (string, error) {
	id, ok := args["id"].(float64)
	if !ok {
		return "", fmt.Errorf("id is required")
	}

	req := WorkItemUpdateRequest{}

	if title := getString(args, "title"); title != "" {
		req.Title = &title
	}
	if state := getString(args, "state"); state != "" {
		req.State = &state
	}
	if assigned := getString(args, "assigned_to"); assigned != "" {
		req.AssignedTo = &assigned
	}
	if priority, ok := args["priority"].(float64); ok {
		p := int(priority)
		req.Priority = &p
	}

	item, err := t.client.UpdateWorkItem(ctx, int(id), req)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Updated work item #%d: %s", item.ID, item.Fields["System.Title"]), nil
}

func (t *Tool) queryWorkItems(ctx context.Context, args map[string]interface{}) (string, error) {
	query := getString(args, "query")
	if query == "" {
		return "", fmt.Errorf("query is required")
	}

	items, err := t.client.QueryWorkItems(ctx, query)
	if err != nil {
		return "", err
	}
	return formatWorkItems(items), nil
}

func (t *Tool) listPipelines(ctx context.Context) (string, error) {
	pipelines, err := t.client.ListPipelines(ctx)
	if err != nil {
		return "", err
	}
	return formatPipelines(pipelines), nil
}

func (t *Tool) runPipeline(ctx context.Context, args map[string]interface{}) (string, error) {
	pipelineID, ok := args["pipeline_id"].(float64)
	if !ok {
		return "", fmt.Errorf("pipeline_id is required")
	}

	branch := getString(args, "branch")
	if branch == "" {
		branch = "refs/heads/main"
	}

	var variables map[string]string
	if vars, ok := args["variables"].(map[string]interface{}); ok {
		variables = make(map[string]string)
		for k, v := range vars {
			if s, ok := v.(string); ok {
				variables[k] = s
			}
		}
	}

	run, err := t.client.RunPipeline(ctx, int(pipelineID), branch, variables)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Started pipeline run #%d: %s (state: %s)", run.ID, run.Name, run.State), nil
}

func (t *Tool) listRepos(ctx context.Context) (string, error) {
	repos, err := t.client.ListRepositories(ctx)
	if err != nil {
		return "", err
	}
	return formatRepos(repos), nil
}

func (t *Tool) listBoards(ctx context.Context, args map[string]interface{}) (string, error) {
	team := getString(args, "team")
	boards, err := t.client.ListBoards(ctx, team)
	if err != nil {
		return "", err
	}
	return formatBoards(boards), nil
}

// Helper functions
func getString(args map[string]interface{}, key string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return ""
}

func formatWorkItems(items []WorkItem) string {
	if len(items) == 0 {
		return "No work items found."
	}

	result := fmt.Sprintf("Found %d work items:\n\n", len(items))
	for _, item := range items {
		result += fmt.Sprintf("- #%d [%s] %s (State: %s)\n",
			item.ID,
			item.Fields["System.WorkItemType"],
			item.Fields["System.Title"],
			item.Fields["System.State"],
		)
	}
	return result
}

func formatWorkItem(item *WorkItem) string {
	result := fmt.Sprintf("Work Item #%d\n", item.ID)
	result += fmt.Sprintf("Type: %s\n", item.Fields["System.WorkItemType"])
	result += fmt.Sprintf("Title: %s\n", item.Fields["System.Title"])
	result += fmt.Sprintf("State: %s\n", item.Fields["System.State"])
	
	if assigned, ok := item.Fields["System.AssignedTo"].(map[string]interface{}); ok {
		result += fmt.Sprintf("Assigned To: %s\n", assigned["displayName"])
	}
	
	if desc, ok := item.Fields["System.Description"].(string); ok && desc != "" {
		result += fmt.Sprintf("Description: %s\n", desc)
	}
	
	if tags, ok := item.Fields["System.Tags"].(string); ok && tags != "" {
		result += fmt.Sprintf("Tags: %s\n", tags)
	}

	return result
}

func formatPipelines(pipelines []Pipeline) string {
	if len(pipelines) == 0 {
		return "No pipelines found."
	}

	result := fmt.Sprintf("Found %d pipelines:\n\n", len(pipelines))
	for _, p := range pipelines {
		result += fmt.Sprintf("- [%d] %s (folder: %s)\n", p.ID, p.Name, p.Folder)
	}
	return result
}

func formatRepos(repos []Repository) string {
	if len(repos) == 0 {
		return "No repositories found."
	}

	result := fmt.Sprintf("Found %d repositories:\n\n", len(repos))
	for _, r := range repos {
		result += fmt.Sprintf("- %s (default branch: %s)\n", r.Name, r.DefaultBranch)
	}
	return result
}

func formatBoards(boards []Board) string {
	if len(boards) == 0 {
		return "No boards found."
	}

	result := fmt.Sprintf("Found %d boards:\n\n", len(boards))
	for _, b := range boards {
		result += fmt.Sprintf("- %s\n", b.Name)
	}
	return result
}
