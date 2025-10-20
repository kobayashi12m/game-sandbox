package game

import (
	"game-sandbox/server/models"
	"game-sandbox/server/utils"
	"log"
	"math"
	"math/rand"
	"time"
)

// NPCを追加
func (g *Game) AddNPC(count int) {
	log.Printf("Adding %d NPCs to game", count)

	names := []string{"Bot Alpha", "Bot Beta", "Bot Gamma", "Bot Delta", "Bot Epsilon",
		"Bot Zeta", "Bot Eta", "Bot Theta", "Bot Iota", "Bot Kappa"}

	for i := range count {
		npcID := utils.GenerateID()
		npcName := names[i%len(names)]
		if i >= len(names) {
			npcName = "Bot " + string(rune('A'+i))
		}

		log.Printf("Creating NPC: %s (%s)", npcName, npcID)

		// 既存のAddPlayer関数を使ってNPCを追加（WebSocket接続はnil）
		g.AddPlayer(npcID, npcName, nil)

		// NPCフラグを設定
		if player, exists := g.Players[npcID]; exists {
			player.IsNPC = true
			player.LastDirectionChange = time.Now()
			log.Printf("NPC %s added successfully", npcName)
		} else {
			log.Printf("Failed to add NPC %s", npcName)
		}
	}

	log.Printf("Total players after adding NPCs: %d", len(g.Players))
}

// NPCの方向をより自然に更新
func (g *Game) updateNPCDirections() {
	now := time.Now()

	for _, player := range g.Players {
		if !player.IsNPC || !player.Organism.Alive {
			continue
		}

		// 最低1秒は同じ方向に進む
		if now.Sub(player.LastDirectionChange) < time.Second {
			continue
		}

		// 食べ物に向かう行動を優先
		targetFood := g.findNearestFood(player.Organism.Core.Position)

		var newDirection *utils.Direction

		if targetFood != nil && rand.Float64() < 0.7 { // 70%の確率で食べ物に向かう
			newDirection = g.calculateDirectionToTarget(player.Organism.Core.Position, *targetFood)
		} else if rand.Float64() < 0.3 { // 30%の確率でランダムに方向変更
			directions := []string{"UP", "DOWN", "LEFT", "RIGHT"}
			randomDir := directions[rand.Intn(len(directions))]
			if dir, ok := utils.DIRECTIONS[randomDir]; ok {
				newDirection = &dir
			}
		}

		// 新しい方向が決まった場合のみ変更
		if newDirection != nil {
			player.Organism.SetAcceleration(newDirection.X, newDirection.Y)
			player.LastDirectionChange = now
		}
	}
}

// 最も近い食べ物を探す
func (g *Game) findNearestFood(head models.Position) *models.Position {
	if len(g.Food) == 0 {
		return nil
	}

	var nearestFood *models.Position
	minDistance := math.MaxFloat64

	for _, food := range g.Food {
		distance := g.calculateDistance(head, food.Position)
		if distance < minDistance {
			minDistance = distance
			nearestFood = &food.Position
		}
	}

	return nearestFood
}

// 2点間の距離を計算
func (g *Game) calculateDistance(p1 models.Position, p2 models.Position) float64 {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// 目標位置への方向を計算
func (g *Game) calculateDirectionToTarget(from, to models.Position) *utils.Direction {
	dx := to.X - from.X
	dy := to.Y - from.Y

	// より大きい成分の方向を選択
	if math.Abs(dx) > math.Abs(dy) {
		if dx > 0 {
			dir := utils.DIRECTIONS["RIGHT"]
			return &dir
		} else {
			dir := utils.DIRECTIONS["LEFT"]
			return &dir
		}
	} else {
		if dy > 0 {
			dir := utils.DIRECTIONS["DOWN"]
			return &dir
		} else {
			dir := utils.DIRECTIONS["UP"]
			return &dir
		}
	}
}
