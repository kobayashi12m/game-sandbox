package game

import (
	"game-sandbox/server/utils"
	"math/rand"
	"time"
)

// NPCを追加
func (g *Game) AddNPC(count int) {
	names := []string{"Bot Alpha", "Bot Beta", "Bot Gamma", "Bot Delta", "Bot Epsilon",
		"Bot Zeta", "Bot Eta", "Bot Theta", "Bot Iota", "Bot Kappa"}

	for i := range count {
		npcID := utils.GenerateID()
		npcName := names[i%len(names)]
		if i >= len(names) {
			npcName = "Bot " + string(rune('A'+i))
		}

		// 既存のAddPlayer関数を使ってNPCを追加（WebSocket接続はnil）
		g.AddPlayer(npcID, npcName, nil)

		utils.LogConnectionEvent("npc_joined", npcID, npcName, true)
	}

	utils.Info("NPCs added to game", map[string]interface{}{
		"event":         "npc_batch_add",
		"game_id":       g.ID,
		"npc_count":     count,
		"total_players": len(g.Players),
	})
}

// NPCのAIを更新 - 毎フレーム加速度を設定
func (g *Game) UpdateNPCAI() {
	for _, player := range g.Players {
		if !player.IsNPC || !player.Celestial.Alive {
			continue
		}

		// 初回または1秒経過で方向を新しく選択
		if player.TargetDirection == nil || time.Since(player.LastDirectionChange) > time.Second {
			// 4方向のシンプルな移動
			directions := []struct{ X, Y float64 }{
				{1, 0},  // 右
				{-1, 0}, // 左
				{0, 1},  // 下
				{0, -1}, // 上
			}

			dir := directions[rand.Intn(len(directions))]
			player.TargetDirection = &dir
			player.LastDirectionChange = time.Now()
		}

		// 毎フレーム、目標方向に加速し続ける（プレイヤーと同じ）
		player.Celestial.SetAcceleration(player.TargetDirection.X, player.TargetDirection.Y)
	}
}
