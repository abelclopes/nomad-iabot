# Trello Integration Testing Guide

This document describes how to test the Trello integration in Nomad Agent.

## Prerequisites

1. A Trello account
2. A Trello API Key and Token

## Getting Trello Credentials

1. Go to https://trello.com/app-key
2. Copy your API Key
3. Click the "Token" link to generate a token
4. Copy the generated token

## Configuration

Add the following to your `.env` file:

```env
TRELLO_ENABLED=true
TRELLO_API_KEY=your-api-key-here
TRELLO_TOKEN=your-token-here
```

## Testing via WebChat

1. Start the Nomad Agent:
   ```bash
   ./nomad
   ```

2. Open your browser and navigate to `http://localhost:8080`

3. Test the following commands:

### List Your Boards
```
Liste meus boards do Trello
```

### Get Board Details
```
Me mostre os detalhes do board [board-id]
```

### List Lists on a Board
```
Quais são as listas do board [board-id]?
```

### Create a Card
```
Crie um card chamado "Nova Tarefa" na lista [list-id] com descrição "Esta é uma tarefa de teste"
```

### List Cards
```
Liste todos os cards da lista [list-id]
```

### Update a Card
```
Atualize o card [card-id] para mover para a lista [new-list-id]
```

### Add Comment
```
Adicione um comentário no card [card-id] dizendo "Comentário de teste"
```

## Expected Behavior

The agent should:
1. Recognize Trello-related requests
2. Use the appropriate Trello tool
3. Return formatted results with board/card information
4. Handle errors gracefully (e.g., invalid IDs, missing permissions)

## Troubleshooting

### "Trello integration not available"
- Verify TRELLO_ENABLED=true in your .env
- Verify TRELLO_API_KEY and TRELLO_TOKEN are set correctly
- Check logs for "Trello integration enabled" message on startup

### "API error (status 401)"
- Your token may be invalid or expired
- Regenerate your token at https://trello.com/app-key

### "API error (status 404)"
- The board/card/list ID may be incorrect
- Verify you have access to the resource
- Board IDs can be found in the URL: https://trello.com/b/[BOARD-ID]/board-name

## Available Tools

The following Trello tools are available to the agent:

1. **trello_list_boards** - List all accessible boards
2. **trello_get_board** - Get board details
3. **trello_get_lists** - Get all lists from a board
4. **trello_create_list** - Create a new list
5. **trello_create_card** - Create a new card
6. **trello_get_card** - Get card details
7. **trello_get_cards_on_list** - List cards on a list
8. **trello_get_cards_on_board** - List all cards on a board
9. **trello_update_card** - Update card properties
10. **trello_add_comment** - Add a comment to a card
11. **trello_get_board_members** - List board members
