package gateway

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/abelclopes/nomad-iabot/internal/devops"
)

// Azure DevOps handlers

func (g *Gateway) handleListWorkItems(w http.ResponseWriter, r *http.Request) {
	if !g.cfg.AzureDevOps.Enabled {
		respondError(w, http.StatusNotFound, "Azure DevOps integration is not enabled")
		return
	}

	client := devops.NewClient(
		g.cfg.AzureDevOps.Organization,
		g.cfg.AzureDevOps.Project,
		g.cfg.AzureDevOps.PAT,
		g.cfg.AzureDevOps.APIVersion,
	)

	// Check for query parameter
	query := r.URL.Query().Get("query")
	
	var items []devops.WorkItem
	var err error
	
	if query != "" {
		items, err = client.QueryWorkItems(r.Context(), query)
	} else {
		// Default: get my work items
		items, err = client.GetMyWorkItems(r.Context())
	}

	if err != nil {
		g.logger.Error("failed to list work items", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to list work items")
		return
	}

	respondJSON(w, http.StatusOK, items)
}

func (g *Gateway) handleGetWorkItem(w http.ResponseWriter, r *http.Request) {
	if !g.cfg.AzureDevOps.Enabled {
		respondError(w, http.StatusNotFound, "Azure DevOps integration is not enabled")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid work item ID")
		return
	}

	client := devops.NewClient(
		g.cfg.AzureDevOps.Organization,
		g.cfg.AzureDevOps.Project,
		g.cfg.AzureDevOps.PAT,
		g.cfg.AzureDevOps.APIVersion,
	)

	item, err := client.GetWorkItem(r.Context(), id)
	if err != nil {
		g.logger.Error("failed to get work item", "error", err, "id", id)
		respondError(w, http.StatusInternalServerError, "failed to get work item")
		return
	}

	respondJSON(w, http.StatusOK, item)
}

func (g *Gateway) handleCreateWorkItem(w http.ResponseWriter, r *http.Request) {
	if !g.cfg.AzureDevOps.Enabled {
		respondError(w, http.StatusNotFound, "Azure DevOps integration is not enabled")
		return
	}

	var req struct {
		Type        string   `json:"type"`
		Title       string   `json:"title"`
		Description string   `json:"description,omitempty"`
		AssignedTo  string   `json:"assigned_to,omitempty"`
		Priority    int      `json:"priority,omitempty"`
		Tags        []string `json:"tags,omitempty"`
		ParentID    int      `json:"parent_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Type == "" || req.Title == "" {
		respondError(w, http.StatusBadRequest, "type and title are required")
		return
	}

	client := devops.NewClient(
		g.cfg.AzureDevOps.Organization,
		g.cfg.AzureDevOps.Project,
		g.cfg.AzureDevOps.PAT,
		g.cfg.AzureDevOps.APIVersion,
	)

	createReq := devops.WorkItemCreateRequest{
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		AssignedTo:  req.AssignedTo,
		Priority:    req.Priority,
		Tags:        req.Tags,
		ParentID:    req.ParentID,
	}

	item, err := client.CreateWorkItem(r.Context(), createReq)
	if err != nil {
		g.logger.Error("failed to create work item", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to create work item")
		return
	}

	respondJSON(w, http.StatusCreated, item)
}

func (g *Gateway) handleUpdateWorkItem(w http.ResponseWriter, r *http.Request) {
	if !g.cfg.AzureDevOps.Enabled {
		respondError(w, http.StatusNotFound, "Azure DevOps integration is not enabled")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid work item ID")
		return
	}

	var req struct {
		Title       *string `json:"title,omitempty"`
		Description *string `json:"description,omitempty"`
		State       *string `json:"state,omitempty"`
		AssignedTo  *string `json:"assigned_to,omitempty"`
		Priority    *int    `json:"priority,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	client := devops.NewClient(
		g.cfg.AzureDevOps.Organization,
		g.cfg.AzureDevOps.Project,
		g.cfg.AzureDevOps.PAT,
		g.cfg.AzureDevOps.APIVersion,
	)

	updateReq := devops.WorkItemUpdateRequest{
		Title:       req.Title,
		Description: req.Description,
		State:       req.State,
		AssignedTo:  req.AssignedTo,
		Priority:    req.Priority,
	}

	item, err := client.UpdateWorkItem(r.Context(), id, updateReq)
	if err != nil {
		g.logger.Error("failed to update work item", "error", err, "id", id)
		respondError(w, http.StatusInternalServerError, "failed to update work item")
		return
	}

	respondJSON(w, http.StatusOK, item)
}

func (g *Gateway) handleListPipelines(w http.ResponseWriter, r *http.Request) {
	if !g.cfg.AzureDevOps.Enabled {
		respondError(w, http.StatusNotFound, "Azure DevOps integration is not enabled")
		return
	}

	client := devops.NewClient(
		g.cfg.AzureDevOps.Organization,
		g.cfg.AzureDevOps.Project,
		g.cfg.AzureDevOps.PAT,
		g.cfg.AzureDevOps.APIVersion,
	)

	pipelines, err := client.ListPipelines(r.Context())
	if err != nil {
		g.logger.Error("failed to list pipelines", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to list pipelines")
		return
	}

	respondJSON(w, http.StatusOK, pipelines)
}

func (g *Gateway) handleRunPipeline(w http.ResponseWriter, r *http.Request) {
	if !g.cfg.AzureDevOps.Enabled {
		respondError(w, http.StatusNotFound, "Azure DevOps integration is not enabled")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid pipeline ID")
		return
	}

	var req struct {
		Branch    string            `json:"branch"`
		Variables map[string]string `json:"variables,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Body is optional
		req.Branch = "refs/heads/main"
	}

	if req.Branch == "" {
		req.Branch = "refs/heads/main"
	}

	client := devops.NewClient(
		g.cfg.AzureDevOps.Organization,
		g.cfg.AzureDevOps.Project,
		g.cfg.AzureDevOps.PAT,
		g.cfg.AzureDevOps.APIVersion,
	)

	run, err := client.RunPipeline(r.Context(), id, req.Branch, req.Variables)
	if err != nil {
		g.logger.Error("failed to run pipeline", "error", err, "id", id)
		respondError(w, http.StatusInternalServerError, "failed to run pipeline")
		return
	}

	respondJSON(w, http.StatusAccepted, run)
}

func (g *Gateway) handleListRepos(w http.ResponseWriter, r *http.Request) {
	if !g.cfg.AzureDevOps.Enabled {
		respondError(w, http.StatusNotFound, "Azure DevOps integration is not enabled")
		return
	}

	client := devops.NewClient(
		g.cfg.AzureDevOps.Organization,
		g.cfg.AzureDevOps.Project,
		g.cfg.AzureDevOps.PAT,
		g.cfg.AzureDevOps.APIVersion,
	)

	repos, err := client.ListRepositories(r.Context())
	if err != nil {
		g.logger.Error("failed to list repos", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to list repositories")
		return
	}

	respondJSON(w, http.StatusOK, repos)
}

func (g *Gateway) handleListBoards(w http.ResponseWriter, r *http.Request) {
	if !g.cfg.AzureDevOps.Enabled {
		respondError(w, http.StatusNotFound, "Azure DevOps integration is not enabled")
		return
	}

	team := r.URL.Query().Get("team")

	client := devops.NewClient(
		g.cfg.AzureDevOps.Organization,
		g.cfg.AzureDevOps.Project,
		g.cfg.AzureDevOps.PAT,
		g.cfg.AzureDevOps.APIVersion,
	)

	boards, err := client.ListBoards(r.Context(), team)
	if err != nil {
		g.logger.Error("failed to list boards", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to list boards")
		return
	}

	respondJSON(w, http.StatusOK, boards)
}
