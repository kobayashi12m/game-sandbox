package game

import (
	"encoding/json"
	"log"
	"sort"

	"game-sandbox/server/models"
	"game-sandbox/server/utils"

	"github.com/gorilla/websocket"
)

// GetState はクライアント用の現在のゲーム状態を返す
func (g *Game) GetState() models.GameState {
	g.mu.RLock()
	defer g.mu.RUnlock()

	players := make([]models.PlayerState, 0, len(g.Players))
	for _, p := range g.Players {
		// 元のデータを変更しないよう球体構造のコピーを作成
		organismCopy := *p.Organism

		players = append(players, models.PlayerState{
			ID:       p.ID,
			Name:     p.Name,
			Organism: &organismCopy,
			Score:    p.Score,
		})
	}
	// Food を Position に変換
	foodPositions := make([]models.Position, len(g.Food))
	for i, food := range g.Food {
		foodPositions[i] = food.Position
	}

	return models.GameState{
		Players: players,
		Food:    foodPositions,
	}
}

// GetOptimizedState はクライアント専用の最適化されたゲーム状態を返す
func (g *Game) GetOptimizedState(clientPlayerID string, clientX, clientY, viewWidth, viewHeight float64) models.GameState {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// クライアントの画面範囲計算
	minX := clientX - viewWidth/2
	maxX := clientX + viewWidth/2
	minY := clientY - viewHeight/2
	maxY := clientY + viewHeight/2

	// Spatial Gridで画面範囲内のプレイヤーと食べ物を同時に取得
	areaResult := g.spatialGrid.GetObjectsInArea(minX, maxX, minY, maxY)

	players := make([]models.PlayerState, 0, len(areaResult.Players))
	for _, p := range areaResult.Players {
		if p.Organism.Core != nil {
			// 元のデータを変更しないよう球体構造のコピーを作成
			organismCopy := *p.Organism

			players = append(players, models.PlayerState{
				ID:       p.ID,
				Name:     p.Name,
				Organism: &organismCopy,
				Score:    p.Score,
			})
		}
	}

	food := make([]models.Position, 0, len(areaResult.Food))
	for _, f := range areaResult.Food {
		food = append(food, f.Position)
	}

	return models.GameState{
		Players: players,
		Food:    food,
	}
}

// GetScoreboard は全プレイヤーのスコア情報をソート済みで返す
func (g *Game) GetScoreboard() []models.ScoreInfo {
	g.mu.RLock()
	defer g.mu.RUnlock()

	scores := make([]models.ScoreInfo, 0, len(g.Players))
	for _, p := range g.Players {
		scores = append(scores, models.ScoreInfo{
			ID:    p.ID,
			Name:  p.Name,
			Score: p.Score,
			Alive: p.Organism.Alive,
			Color: p.Organism.Color,
		})
	}

	// サーバー側でスコア順にソート（高い順）
	// 同スコアの場合はIDでソート
	sort.Slice(scores, func(i, j int) bool {
		if scores[i].Score != scores[j].Score {
			return scores[i].Score > scores[j].Score // スコア高い順
		}
		return scores[i].ID < scores[j].ID // 同スコアならID昇順
	})

	return scores
}

// BroadcastScoreboard はスコアボード情報を全クライアントに送信
func (g *Game) BroadcastScoreboard() {
	g.mu.RLock()
	playerList := make([]*models.Player, 0, len(g.humanPlayers))
	for _, player := range g.humanPlayers {
		if player.Conn != nil {
			playerList = append(playerList, player)
		}
	}
	g.mu.RUnlock()

	if len(playerList) == 0 {
		return
	}

	// スコアボード情報を取得
	scoreboard := g.GetScoreboard()
	scoreUpdate := models.ScoreUpdate{Players: scoreboard}

	message := map[string]interface{}{
		"type":       "scoreboard",
		"scoreboard": scoreUpdate,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling scoreboard: %v", err)
		return
	}

	// 全クライアントに送信
	for _, player := range playerList {
		func() {
			player.ConnMu.Lock()
			defer player.ConnMu.Unlock()
			if err := player.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("\033[31m❌ WS_ERROR: Error broadcasting scoreboard to player %s: %v\033[0m", player.ID, err)
			}
		}()
	}
}

// BroadcastOptimized は各クライアントに最適化されたデータを個別送信
func (g *Game) BroadcastOptimized() {
	// キャッシュされた人間プレイヤーリストを取得
	g.mu.RLock()
	playerList := make([]*models.Player, 0, len(g.humanPlayers))
	for _, player := range g.humanPlayers {
		// 接続が切断されていないかチェック
		if player.Conn == nil {
			continue
		}

		// プレイヤーの位置を取得（死んでいても送信を続ける）
		if player.Organism.Core == nil {
			continue
		}

		playerList = append(playerList, player)
	}
	g.mu.RUnlock()

	// スナップショットを使って各プレイヤーに送信（デッドロック回避）
	for _, player := range playerList {
		head := player.Organism.Core.Position
		// constants.goからカリング範囲を取得
		viewWidth := utils.CULLING_WIDTH
		viewHeight := utils.CULLING_HEIGHT

		// このクライアント専用の最適化されたゲーム状態を取得
		optimizedState := g.GetOptimizedState(player.ID, head.X, head.Y, viewWidth, viewHeight)

		message := map[string]interface{}{
			"type":  "gameState",
			"state": optimizedState,
		}

		data, err := json.Marshal(message)
		if err != nil {
			log.Printf("Error marshaling optimized state for player %s: %v", player.ID, err)
			continue
		}

		// WebSocket書き込みを同期化
		func() {
			player.ConnMu.Lock()
			defer player.ConnMu.Unlock()
			if err := player.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("\033[31m❌ WS_ERROR: Error broadcasting optimized state to player %s: %v\033[0m", player.ID, err)
			}
			// デバッグ：データサイズをログ出力（10秒に1回）
			if g.frameCount%600 == 0 {
				log.Printf("\033[34m📊 WS_DATA: size for player %s: %d bytes\033[0m", player.Name, len(data))
			}
		}()
	}
}
