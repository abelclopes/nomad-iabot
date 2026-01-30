package trello

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
	"encoding/json"
)

// Client is a Trello REST API client
type Client struct {
	apiKey      string
	token       string
	httpClient  *http.Client
	baseURL     string
}

// NewClient creates a new Trello client
func NewClient(apiKey, token string) *Client {
	return &Client{
		apiKey: apiKey,
		token:  token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.trello.com/1",
	}
}

// ========================================
// Boards
// ========================================

// Board represents a Trello board
type Board struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Desc        string `json:"desc"`
	URL         string `json:"url"`
	ShortURL    string `json:"shortUrl"`
	Closed      bool   `json:"closed"`
	IDOrganization string `json:"idOrganization,omitempty"`
}

// ListBoards lists all boards for the authenticated user
func (c *Client) ListBoards(ctx context.Context) ([]Board, error) {
	endpoint := fmt.Sprintf("%s/members/me/boards", c.baseURL)
	
	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var boards []Board
	if err := json.NewDecoder(resp.Body).Decode(&boards); err != nil {
		return nil, fmt.Errorf("failed to decode boards: %w", err)
	}

	return boards, nil
}

// GetBoard retrieves a specific board by ID
func (c *Client) GetBoard(ctx context.Context, boardID string) (*Board, error) {
	endpoint := fmt.Sprintf("%s/boards/%s", c.baseURL, boardID)
	
	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var board Board
	if err := json.NewDecoder(resp.Body).Decode(&board); err != nil {
		return nil, fmt.Errorf("failed to decode board: %w", err)
	}

	return &board, nil
}

// ========================================
// Lists
// ========================================

// List represents a Trello list
type List struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Closed  bool   `json:"closed"`
	IDBoard string `json:"idBoard"`
	Pos     float64 `json:"pos"`
}

// GetLists retrieves all lists from a board
func (c *Client) GetLists(ctx context.Context, boardID string) ([]List, error) {
	endpoint := fmt.Sprintf("%s/boards/%s/lists", c.baseURL, boardID)
	
	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var lists []List
	if err := json.NewDecoder(resp.Body).Decode(&lists); err != nil {
		return nil, fmt.Errorf("failed to decode lists: %w", err)
	}

	return lists, nil
}

// CreateList creates a new list on a board
func (c *Client) CreateList(ctx context.Context, boardID, name string) (*List, error) {
	endpoint := fmt.Sprintf("%s/lists", c.baseURL)
	
	params := url.Values{}
	params.Set("name", name)
	params.Set("idBoard", boardID)
	
	resp, err := c.doRequestWithParams(ctx, "POST", endpoint, params, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var list List
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("failed to decode list: %w", err)
	}

	return &list, nil
}

// ========================================
// Cards
// ========================================

// Card represents a Trello card
type Card struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Desc        string   `json:"desc"`
	Closed      bool     `json:"closed"`
	IDList      string   `json:"idList"`
	IDBoard     string   `json:"idBoard"`
	IDMembers   []string `json:"idMembers"`
	IDLabels    []string `json:"idLabels"`
	URL         string   `json:"url"`
	ShortURL    string   `json:"shortUrl"`
	Due         string   `json:"due,omitempty"`
	Labels      []Label  `json:"labels,omitempty"`
}

// Label represents a Trello label
type Label struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// CreateCardRequest represents a card creation request
type CreateCardRequest struct {
	Name      string
	Desc      string
	ListID    string
	Position  string   // "top", "bottom", or a number
	DueDate   string   // ISO 8601 date format
	MemberIDs []string
	LabelIDs  []string
}

// CreateCard creates a new card on a list
func (c *Client) CreateCard(ctx context.Context, req CreateCardRequest) (*Card, error) {
	endpoint := fmt.Sprintf("%s/cards", c.baseURL)
	
	params := url.Values{}
	params.Set("name", req.Name)
	params.Set("idList", req.ListID)
	
	if req.Desc != "" {
		params.Set("desc", req.Desc)
	}
	if req.Position != "" {
		params.Set("pos", req.Position)
	}
	if req.DueDate != "" {
		params.Set("due", req.DueDate)
	}
	if len(req.MemberIDs) > 0 {
		for _, memberID := range req.MemberIDs {
			params.Add("idMembers", memberID)
		}
	}
	if len(req.LabelIDs) > 0 {
		for _, labelID := range req.LabelIDs {
			params.Add("idLabels", labelID)
		}
	}
	
	resp, err := c.doRequestWithParams(ctx, "POST", endpoint, params, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var card Card
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, fmt.Errorf("failed to decode card: %w", err)
	}

	return &card, nil
}

// GetCard retrieves a specific card by ID
func (c *Client) GetCard(ctx context.Context, cardID string) (*Card, error) {
	endpoint := fmt.Sprintf("%s/cards/%s", c.baseURL, cardID)
	
	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var card Card
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, fmt.Errorf("failed to decode card: %w", err)
	}

	return &card, nil
}

// GetCardsOnList retrieves all cards from a list
func (c *Client) GetCardsOnList(ctx context.Context, listID string) ([]Card, error) {
	endpoint := fmt.Sprintf("%s/lists/%s/cards", c.baseURL, listID)
	
	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var cards []Card
	if err := json.NewDecoder(resp.Body).Decode(&cards); err != nil {
		return nil, fmt.Errorf("failed to decode cards: %w", err)
	}

	return cards, nil
}

// GetCardsOnBoard retrieves all cards from a board
func (c *Client) GetCardsOnBoard(ctx context.Context, boardID string) ([]Card, error) {
	endpoint := fmt.Sprintf("%s/boards/%s/cards", c.baseURL, boardID)
	
	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var cards []Card
	if err := json.NewDecoder(resp.Body).Decode(&cards); err != nil {
		return nil, fmt.Errorf("failed to decode cards: %w", err)
	}

	return cards, nil
}

// UpdateCardRequest represents a card update request
type UpdateCardRequest struct {
	Name      *string
	Desc      *string
	Closed    *bool
	IDList    *string
	IDMembers []string
	Due       *string
}

// UpdateCard updates an existing card
func (c *Client) UpdateCard(ctx context.Context, cardID string, req UpdateCardRequest) (*Card, error) {
	endpoint := fmt.Sprintf("%s/cards/%s", c.baseURL, cardID)
	
	params := url.Values{}
	
	if req.Name != nil {
		params.Set("name", *req.Name)
	}
	if req.Desc != nil {
		params.Set("desc", *req.Desc)
	}
	if req.Closed != nil {
		params.Set("closed", fmt.Sprintf("%t", *req.Closed))
	}
	if req.IDList != nil {
		params.Set("idList", *req.IDList)
	}
	if req.Due != nil {
		params.Set("due", *req.Due)
	}
	if len(req.IDMembers) > 0 {
		for _, memberID := range req.IDMembers {
			params.Add("idMembers", memberID)
		}
	}
	
	resp, err := c.doRequestWithParams(ctx, "PUT", endpoint, params, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var card Card
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, fmt.Errorf("failed to decode card: %w", err)
	}

	return &card, nil
}

// ========================================
// Members
// ========================================

// Member represents a Trello member
type Member struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	FullName string `json:"fullName"`
	Initials string `json:"initials"`
	AvatarURL string `json:"avatarUrl,omitempty"`
}

// GetBoardMembers retrieves all members of a board
func (c *Client) GetBoardMembers(ctx context.Context, boardID string) ([]Member, error) {
	endpoint := fmt.Sprintf("%s/boards/%s/members", c.baseURL, boardID)
	
	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var members []Member
	if err := json.NewDecoder(resp.Body).Decode(&members); err != nil {
		return nil, fmt.Errorf("failed to decode members: %w", err)
	}

	return members, nil
}

// ========================================
// Comments
// ========================================

// Comment represents a comment on a card
type Comment struct {
	ID              string `json:"id"`
	IDMemberCreator string `json:"idMemberCreator"`
	Data            struct {
		Text string `json:"text"`
	} `json:"data"`
	Date string `json:"date"`
}

// AddComment adds a comment to a card
func (c *Client) AddComment(ctx context.Context, cardID, text string) (*Comment, error) {
	endpoint := fmt.Sprintf("%s/cards/%s/actions/comments", c.baseURL, cardID)
	
	params := url.Values{}
	params.Set("text", text)
	
	resp, err := c.doRequestWithParams(ctx, "POST", endpoint, params, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var comment Comment
	if err := json.NewDecoder(resp.Body).Decode(&comment); err != nil {
		return nil, fmt.Errorf("failed to decode comment: %w", err)
	}

	return &comment, nil
}

// ========================================
// Helpers
// ========================================

func (c *Client) doRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Response, error) {
	return c.doRequestWithParams(ctx, method, endpoint, nil, body)
}

func (c *Client) doRequestWithParams(ctx context.Context, method, endpoint string, params url.Values, body io.Reader) (*http.Response, error) {
	// Add auth parameters
	if params == nil {
		params = url.Values{}
	}
	params.Set("key", c.apiKey)
	params.Set("token", c.token)
	
	// Build URL with query parameters
	fullURL := endpoint
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}
	
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

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
