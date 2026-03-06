package classic

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/N3moAhead/bombahead/server/internal/game"
	"github.com/N3moAhead/bombahead/server/internal/message"
	"github.com/N3moAhead/bombahead/server/pkg/logger"
	"github.com/N3moAhead/bombahead/server/pkg/types"
)

var log = logger.New("[Classic]")

type Classic struct {
	gameFinisher game.GameFinisher
	stopChan     chan bool

	gameID          string
	players         map[string]*Player     // ClientID -> ClassicPlayer
	playerMap       map[string]game.Player // ClientID -> game.Player
	playerMux       sync.RWMutex
	history         *History
	historyFilePath string

	field      *Field
	bombs      map[string]*Bomb      // Bomb.Pos -> Bomb
	explosions map[string]types.Vec2 // Pos -> Vec2(Pos of the Bomb)

	isRunning  bool
	minPlayers int
	maxPLayers int
	isTimeOut  bool

	ticker       *time.Ticker
	lastTickTime time.Time // for delta time
}

func NewClassic(finisher game.GameFinisher, id string, historyFilePath string) *Classic {
	return &Classic{
		gameFinisher: finisher,
		stopChan:     make(chan bool),

		gameID:          id,
		players:         make(map[string]*Player),
		playerMap:       make(map[string]game.Player),
		historyFilePath: historyFilePath,

		field:      NewField(),
		bombs:      make(map[string]*Bomb),
		explosions: make(map[string]types.Vec2),

		isRunning:  false,
		minPlayers: MIN_PLAYERS,
		maxPLayers: MAX_PLAYERS,
		isTimeOut:  false,
	}
}

func (c *Classic) GetID() string {
	return c.gameID
}

func (c *Classic) AddPlayer(player game.Player) error {
	c.playerMux.Lock()
	defer c.playerMux.Unlock()

	if len(c.players) >= c.maxPLayers {
		return fmt.Errorf(
			"[Game %s] Already full, can't add more players (%d/%d)!\n",
			c.gameID,
			len(c.players),
			c.maxPLayers,
		)
	}

	playerID := player.GetID()
	if _, exists := c.players[playerID]; exists {
		return fmt.Errorf("[Game %s] Player %s already exists.\n", c.gameID, playerID)
	}

	// We define the spawn points for the players in the corners of the map
	spawnPoints := []types.Vec2{
		types.NewVec2(1, 1),                          // Top-Left
		types.NewVec2(field_width-2, field_height-2), // Bottom-Right
		types.NewVec2(field_width-2, 1),              // Top-Right
		types.NewVec2(1, field_height-2),             // Bottom-Left
	}

	// Assign a spawn point based on the number of players already in the game.
	playerIndex := len(c.players)
	spawnPos := spawnPoints[playerIndex]

	newPlayer := &Player{
		ID:        playerID,
		Pos:       spawnPos,
		Score:     0,
		Health:    initial_health,
		NextMove:  NO_INPUT_DEFINED,
		AuthToken: player.GetAuthToken(),
	}
	c.players[playerID] = newPlayer
	c.playerMap[playerID] = player

	log.Success("Player %s added. (Game %s)\n", playerID, c.gameID)
	return nil
}

func (c *Classic) RemovePlayer(player game.Player) {
	c.playerMux.Lock()
	defer c.playerMux.Unlock()

	playerID := player.GetID()
	if _, ok := c.players[playerID]; ok {
		delete(c.players, playerID)
		delete(c.playerMap, playerID)
		log.Info("[Game %s] Player %s removed.\n", c.gameID, playerID)

		if len(c.players) < c.minPlayers && c.isRunning {
			log.Warn(
				"[Game %s] Not enough players remaining (%d/%d). Stopping game.\n",
				c.gameID,
				len(c.players),
				c.minPlayers,
			)
			go c.Stop()
		}
	}
}

func (c *Classic) Start() {
	c.playerMux.Lock()
	if len(c.players) < c.minPlayers {
		c.playerMux.Unlock()
		log.Error("[Game %s] Cannot start, not enough players (%d/%d).", c.gameID, len(c.players), c.minPlayers)
		go c.Stop()
		return
	}

	c.isRunning = true
	// Initialize history recording at the start of the game
	c.history = NewHistory(c.getGameState().Field)
	c.lastTickTime = time.Now()
	c.ticker = time.NewTicker(TICK_RATE)
	c.playerMux.Unlock()

	log.Info("[Game %s] Starting game loop.", c.gameID)
	defer func() {
		if c.ticker != nil {
			c.ticker.Stop()
		}
		log.Info("[Game %s] Game loop stopped.", c.gameID)
	}()

	maxGameTimer := time.NewTimer(MAX_GAME_TIME)
	defer maxGameTimer.Stop()

	for {
		select {
		case <-c.ticker.C:
			// A snapshot of players is created to
			// avoid holding the lock during network I/O
			c.playerMux.RLock()
			playersToMessage := make([]game.Player, 0, len(c.playerMap))
			for _, p := range c.playerMap {
				playersToMessage = append(playersToMessage, p)
			}
			c.playerMux.RUnlock()

			// Lock the mutex to ensure exclusive access to
			// the game state during the update.
			c.playerMux.Lock()
			destroyedBoxes := c.update()
			c.history.RecordTick(c.players, c.bombs, c.explosions, destroyedBoxes)
			gameState := c.getGameState()
			gameOver := c.isGameOver()
			c.resetPlayerInputs()
			c.playerMux.Unlock()

			// Now, with the mutex released, we can safely send the new state to all players.
			for _, p := range playersToMessage {
				if err := p.SendMessage(message.ClassicState, gameState); err != nil {
					log.Error(
						"[Game %s] Error sending state to player %s: %v",
						c.gameID,
						p.GetID(),
						err,
					)
				}
			}
			if gameOver {
				go c.Stop()
			}
		case <-maxGameTimer.C:
			c.playerMux.Lock()
			c.isTimeOut = true
			c.playerMux.Unlock()
			go c.Stop()

		case <-c.stopChan:
			// When receiving a stop signal we stop the goroutine
			return
		}
	}
}

func (c *Classic) Stop() {
	c.playerMux.Lock()
	if !c.isRunning {
		c.playerMux.Unlock()
		return
	}
	c.isRunning = false

	if c.ticker != nil {
		c.ticker.Stop()
		c.ticker = nil
	}

	// closing the stop channel
	select {
	case <-c.stopChan:
	default:
		close(c.stopChan)
	}

	result := game.GameResult{
		Winner: "",
		Scores: make(map[string]int),
	}

	if c.isTimeOut {
		var healthiestPlayer *Player = nil
		isUnique := true
		// The time ran out so we will check if one of the
		// players has more health left then the others
		for _, player := range c.players {
			if healthiestPlayer == nil || player.Health > healthiestPlayer.Health {
				healthiestPlayer = player
			} else if player.Health == healthiestPlayer.Health {
				isUnique = false
			}
		}

		if isUnique {
			result.Winner = healthiestPlayer.ID
		}
	} else {
		// There can only be one winner
		// The field will be left empty, if
		// it's a draw
		for pId, player := range c.players {
			if player.Health > 0 {
				if result.Winner != "" {
					result.Winner = ""
					break
				}
				result.Winner = pId
			}
		}
	}

	if result.Winner != "" {
		result.Scores[result.Winner] = WIN_SCORE_POINTS
	}

	if c.history != nil {
		var winnerAuthToken string = ""
		if player, ok := c.players[result.Winner]; ok {
			winnerAuthToken = player.AuthToken
		}
		gameHistoryForSerialization := c.history.ToGameHistory(winnerAuthToken)
		b, err := json.Marshal(gameHistoryForSerialization)
		if err != nil {
			log.Error("Failed to marshal game history: %v", err)
		} else {
			if fileInfo, statErr := os.Stat(c.historyFilePath); statErr != nil {
				if os.IsNotExist(statErr) {
					log.Warn(
						"Game history file does not exist. Skipping history write. path='%s'",
						c.historyFilePath,
					)
				} else {
					log.Warn(
						"Could not access game history file. Skipping history write. path='%s', err=%v",
						c.historyFilePath,
						statErr,
					)
				}
			} else if fileInfo.IsDir() {
				log.Warn(
					"Game history path is a directory, not a file. Skipping history write. path='%s'",
					c.historyFilePath,
				)
			} else if writeErr := os.WriteFile(c.historyFilePath, b, 0); writeErr != nil {
				log.Error("Failed to write game history to '%s': %v", c.historyFilePath, writeErr)
			} else {
				log.Success("Game history written to '%s'", c.historyFilePath)
			}
		}
	}

	c.playerMux.Unlock()

	log.Info("Stopping game. (Game %s)", c.gameID)

	// Inform the hub that the game is finished and retrieve all
	// players back to the lobby
	c.gameFinisher.GameFinished(c.gameID, result)
}

func (c *Classic) HandleMessage(player game.Player, msg message.Message) {
	playerID := player.GetID()

	switch msg.Type {
	case message.ClassicInput:
		var payload ClassicInputPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Error("[Game %s] Error unmarshalling ClassInput from %s: %v\n", c.gameID, playerID, err)
			return
		}

		c.playerMux.Lock()
		defer c.playerMux.Unlock()
		pState, ok := c.players[playerID]
		if ok {
			pState.HandleInput(payload)
		} else {
			log.Warn(
				"[Game %s] Received input from player %s who is not in the internal state map.",
				c.gameID,
				playerID,
			)
		}
	// The message is normally not handled here...
	// The normal hub will handle this message. But the oneShot hub will
	// just pass this message down to the game so we can handle it here...
	// In that case we care about the auth token...
	case message.PlayerStatusUpdate:
		var payload message.PlayerStatusUpdatePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Error("[Game %s] Error while unmarshalling PlayerStatusUpdate %v\n", c.gameID, err)
			return
		}

		c.playerMux.Lock()
		defer c.playerMux.Unlock()

		pState, ok := c.players[playerID]
		if ok {
			pState.AuthToken = payload.AuthToken
			log.Success("Player %s received the auth token %s", playerID, payload.AuthToken)
		} else {
			log.Warn(
				"[Game %s] Received input from player %s who is not in the internal state map.",
				c.gameID,
				playerID,
			)
		}

	default:
		log.Warn(
			"[Game %s] Received unhandled message type '%s' from player %s",
			c.gameID,
			msg.Type,
			playerID,
		)
	}
}

var _ game.Game = (*Classic)(nil)
