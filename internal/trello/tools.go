package trello

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/abelclopes/nomad-iabot/internal/llm"
)

// Tool represents a Trello tool for the LLM
type Tool struct {
	client *Client
}

// NewTool creates a new Trello tool
func NewTool(client *Client) *Tool {
	return &Tool{client: client}
}

// GetToolDefinitions returns the tool definitions for the LLM
func (t *Tool) GetToolDefinitions() []llm.Tool {
	return []llm.Tool{
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "trello_list_boards",
				Description: "List all Trello boards accessible to the authenticated user",
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
				Name:        "trello_get_board",
				Description: "Get details of a specific Trello board by ID",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"board_id": map[string]interface{}{
							"type":        "string",
							"description": "The board ID",
						},
					},
					"required": []string{"board_id"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "trello_get_lists",
				Description: "Get all lists from a Trello board",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"board_id": map[string]interface{}{
							"type":        "string",
							"description": "The board ID",
						},
					},
					"required": []string{"board_id"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "trello_create_list",
				Description: "Create a new list on a Trello board",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"board_id": map[string]interface{}{
							"type":        "string",
							"description": "The board ID",
						},
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Name of the new list",
						},
					},
					"required": []string{"board_id", "name"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "trello_create_card",
				Description: "Create a new card on a Trello list",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"list_id": map[string]interface{}{
							"type":        "string",
							"description": "The list ID where the card will be created",
						},
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Title of the card",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Description of the card (Markdown supported)",
						},
						"position": map[string]interface{}{
							"type":        "string",
							"description": "Position of the card: 'top' or 'bottom'",
							"enum":        []string{"top", "bottom"},
						},
						"due_date": map[string]interface{}{
							"type":        "string",
							"description": "Due date in ISO 8601 format (e.g., 2024-12-31T23:59:59Z)",
						},
					},
					"required": []string{"list_id", "name"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "trello_get_card",
				Description: "Get details of a specific Trello card by ID",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"card_id": map[string]interface{}{
							"type":        "string",
							"description": "The card ID",
						},
					},
					"required": []string{"card_id"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "trello_get_cards_on_list",
				Description: "Get all cards from a specific Trello list",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"list_id": map[string]interface{}{
							"type":        "string",
							"description": "The list ID",
						},
					},
					"required": []string{"list_id"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "trello_get_cards_on_board",
				Description: "Get all cards from a Trello board",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"board_id": map[string]interface{}{
							"type":        "string",
							"description": "The board ID",
						},
					},
					"required": []string{"board_id"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "trello_update_card",
				Description: "Update an existing Trello card",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"card_id": map[string]interface{}{
							"type":        "string",
							"description": "The card ID to update",
						},
						"name": map[string]interface{}{
							"type":        "string",
							"description": "New title for the card",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "New description for the card",
						},
						"closed": map[string]interface{}{
							"type":        "boolean",
							"description": "Whether the card is closed (archived)",
						},
						"list_id": map[string]interface{}{
							"type":        "string",
							"description": "Move card to a different list",
						},
						"due": map[string]interface{}{
							"type":        "string",
							"description": "Due date in ISO 8601 format",
						},
					},
					"required": []string{"card_id"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "trello_add_comment",
				Description: "Add a comment to a Trello card",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"card_id": map[string]interface{}{
							"type":        "string",
							"description": "The card ID",
						},
						"text": map[string]interface{}{
							"type":        "string",
							"description": "The comment text",
						},
					},
					"required": []string{"card_id", "text"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "trello_get_board_members",
				Description: "Get all members of a Trello board",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"board_id": map[string]interface{}{
							"type":        "string",
							"description": "The board ID",
						},
					},
					"required": []string{"board_id"},
				},
			},
		},
	}
}

// Execute executes a Trello tool call - returns (result, handled, error)
func (t *Tool) Execute(ctx context.Context, name string, args map[string]interface{}) (string, bool, error) {
	switch name {
	case "trello_list_boards":
		result, err := t.listBoards(ctx)
		return result, true, err
	case "trello_get_board":
		result, err := t.getBoard(ctx, args)
		return result, true, err
	case "trello_get_lists":
		result, err := t.getLists(ctx, args)
		return result, true, err
	case "trello_create_list":
		result, err := t.createList(ctx, args)
		return result, true, err
	case "trello_create_card":
		result, err := t.createCard(ctx, args)
		return result, true, err
	case "trello_get_card":
		result, err := t.getCard(ctx, args)
		return result, true, err
	case "trello_get_cards_on_list":
		result, err := t.getCardsOnList(ctx, args)
		return result, true, err
	case "trello_get_cards_on_board":
		result, err := t.getCardsOnBoard(ctx, args)
		return result, true, err
	case "trello_update_card":
		result, err := t.updateCard(ctx, args)
		return result, true, err
	case "trello_add_comment":
		result, err := t.addComment(ctx, args)
		return result, true, err
	case "trello_get_board_members":
		result, err := t.getBoardMembers(ctx, args)
		return result, true, err
	default:
		return "", false, nil
	}
}

// ExecuteTool executes a Trello tool call (legacy)
func (t *Tool) ExecuteTool(ctx context.Context, name string, arguments string) (string, error) {
	var args map[string]interface{}
	if arguments != "" {
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return "", fmt.Errorf("failed to parse arguments: %w", err)
		}
	}

	result, handled, err := t.Execute(ctx, name, args)
	if !handled {
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	return result, err
}

func (t *Tool) listBoards(ctx context.Context) (string, error) {
	boards, err := t.client.ListBoards(ctx)
	if err != nil {
		return "", err
	}
	return formatBoards(boards), nil
}

func (t *Tool) getBoard(ctx context.Context, args map[string]interface{}) (string, error) {
	boardID := getString(args, "board_id")
	if boardID == "" {
		return "", fmt.Errorf("board_id is required")
	}

	board, err := t.client.GetBoard(ctx, boardID)
	if err != nil {
		return "", err
	}
	return formatBoard(board), nil
}

func (t *Tool) getLists(ctx context.Context, args map[string]interface{}) (string, error) {
	boardID := getString(args, "board_id")
	if boardID == "" {
		return "", fmt.Errorf("board_id is required")
	}

	lists, err := t.client.GetLists(ctx, boardID)
	if err != nil {
		return "", err
	}
	return formatLists(lists), nil
}

func (t *Tool) createList(ctx context.Context, args map[string]interface{}) (string, error) {
	boardID := getString(args, "board_id")
	name := getString(args, "name")
	
	if boardID == "" || name == "" {
		return "", fmt.Errorf("board_id and name are required")
	}

	list, err := t.client.CreateList(ctx, boardID, name)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Created list '%s' (ID: %s)", list.Name, list.ID), nil
}

func (t *Tool) createCard(ctx context.Context, args map[string]interface{}) (string, error) {
	req := CreateCardRequest{
		ListID: getString(args, "list_id"),
		Name:   getString(args, "name"),
	}

	if req.ListID == "" || req.Name == "" {
		return "", fmt.Errorf("list_id and name are required")
	}

	if desc := getString(args, "description"); desc != "" {
		req.Desc = desc
	}
	if pos := getString(args, "position"); pos != "" {
		req.Position = pos
	}
	if due := getString(args, "due_date"); due != "" {
		req.DueDate = due
	}

	card, err := t.client.CreateCard(ctx, req)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Created card '%s' (ID: %s, URL: %s)", card.Name, card.ID, card.ShortURL), nil
}

func (t *Tool) getCard(ctx context.Context, args map[string]interface{}) (string, error) {
	cardID := getString(args, "card_id")
	if cardID == "" {
		return "", fmt.Errorf("card_id is required")
	}

	card, err := t.client.GetCard(ctx, cardID)
	if err != nil {
		return "", err
	}
	return formatCard(card), nil
}

func (t *Tool) getCardsOnList(ctx context.Context, args map[string]interface{}) (string, error) {
	listID := getString(args, "list_id")
	if listID == "" {
		return "", fmt.Errorf("list_id is required")
	}

	cards, err := t.client.GetCardsOnList(ctx, listID)
	if err != nil {
		return "", err
	}
	return formatCards(cards), nil
}

func (t *Tool) getCardsOnBoard(ctx context.Context, args map[string]interface{}) (string, error) {
	boardID := getString(args, "board_id")
	if boardID == "" {
		return "", fmt.Errorf("board_id is required")
	}

	cards, err := t.client.GetCardsOnBoard(ctx, boardID)
	if err != nil {
		return "", err
	}
	return formatCards(cards), nil
}

func (t *Tool) updateCard(ctx context.Context, args map[string]interface{}) (string, error) {
	cardID := getString(args, "card_id")
	if cardID == "" {
		return "", fmt.Errorf("card_id is required")
	}

	req := UpdateCardRequest{}

	if name := getString(args, "name"); name != "" {
		req.Name = &name
	}
	if desc := getString(args, "description"); desc != "" {
		req.Desc = &desc
	}
	if closed, ok := args["closed"].(bool); ok {
		req.Closed = &closed
	}
	if listID := getString(args, "list_id"); listID != "" {
		req.IDList = &listID
	}
	if due := getString(args, "due"); due != "" {
		req.Due = &due
	}

	card, err := t.client.UpdateCard(ctx, cardID, req)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Updated card '%s' (ID: %s)", card.Name, card.ID), nil
}

func (t *Tool) addComment(ctx context.Context, args map[string]interface{}) (string, error) {
	cardID := getString(args, "card_id")
	text := getString(args, "text")
	
	if cardID == "" || text == "" {
		return "", fmt.Errorf("card_id and text are required")
	}

	comment, err := t.client.AddComment(ctx, cardID, text)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Added comment to card (comment ID: %s)", comment.ID), nil
}

func (t *Tool) getBoardMembers(ctx context.Context, args map[string]interface{}) (string, error) {
	boardID := getString(args, "board_id")
	if boardID == "" {
		return "", fmt.Errorf("board_id is required")
	}

	members, err := t.client.GetBoardMembers(ctx, boardID)
	if err != nil {
		return "", err
	}
	return formatMembers(members), nil
}

// Helper functions
func getString(args map[string]interface{}, key string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return ""
}

func formatBoards(boards []Board) string {
	if len(boards) == 0 {
		return "No boards found."
	}

	result := fmt.Sprintf("Found %d boards:\n\n", len(boards))
	for _, board := range boards {
		status := "Open"
		if board.Closed {
			status = "Closed"
		}
		result += fmt.Sprintf("- [%s] %s (ID: %s, URL: %s)\n", status, board.Name, board.ID, board.ShortURL)
	}
	return result
}

func formatBoard(board *Board) string {
	status := "Open"
	if board.Closed {
		status = "Closed"
	}
	
	result := fmt.Sprintf("Board: %s\n", board.Name)
	result += fmt.Sprintf("ID: %s\n", board.ID)
	result += fmt.Sprintf("Status: %s\n", status)
	result += fmt.Sprintf("URL: %s\n", board.ShortURL)
	if board.Desc != "" {
		result += fmt.Sprintf("Description: %s\n", board.Desc)
	}
	return result
}

func formatLists(lists []List) string {
	if len(lists) == 0 {
		return "No lists found."
	}

	result := fmt.Sprintf("Found %d lists:\n\n", len(lists))
	for _, list := range lists {
		status := "Open"
		if list.Closed {
			status = "Closed"
		}
		result += fmt.Sprintf("- [%s] %s (ID: %s)\n", status, list.Name, list.ID)
	}
	return result
}

func formatCards(cards []Card) string {
	if len(cards) == 0 {
		return "No cards found."
	}

	result := fmt.Sprintf("Found %d cards:\n\n", len(cards))
	for _, card := range cards {
		status := "Open"
		if card.Closed {
			status = "Archived"
		}
		result += fmt.Sprintf("- [%s] %s (ID: %s, URL: %s)\n", status, card.Name, card.ID, card.ShortURL)
		if card.Desc != "" {
			result += fmt.Sprintf("  Description: %s\n", card.Desc)
		}
		if card.Due != "" {
			result += fmt.Sprintf("  Due: %s\n", card.Due)
		}
	}
	return result
}

func formatCard(card *Card) string {
	status := "Open"
	if card.Closed {
		status = "Archived"
	}
	
	result := fmt.Sprintf("Card: %s\n", card.Name)
	result += fmt.Sprintf("ID: %s\n", card.ID)
	result += fmt.Sprintf("Status: %s\n", status)
	result += fmt.Sprintf("URL: %s\n", card.ShortURL)
	if card.Desc != "" {
		result += fmt.Sprintf("Description: %s\n", card.Desc)
	}
	if card.Due != "" {
		result += fmt.Sprintf("Due: %s\n", card.Due)
	}
	if len(card.Labels) > 0 {
		result += "Labels: "
		for i, label := range card.Labels {
			if i > 0 {
				result += ", "
			}
			result += fmt.Sprintf("%s (%s)", label.Name, label.Color)
		}
		result += "\n"
	}
	return result
}

func formatMembers(members []Member) string {
	if len(members) == 0 {
		return "No members found."
	}

	result := fmt.Sprintf("Found %d members:\n\n", len(members))
	for _, member := range members {
		result += fmt.Sprintf("- %s (@%s, ID: %s)\n", member.FullName, member.Username, member.ID)
	}
	return result
}
