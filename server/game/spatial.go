package game

import (
	"chess-mmo/server/models"
	"chess-mmo/server/utils"
)

// SpatialGrid はゲームフィールドを格子状に分割して効率的な当たり判定を行う
type SpatialGrid struct {
	cellSize float64       // 各セルのサイズ
	width    int           // グリッドの幅（セル数）
	height   int           // グリッドの高さ（セル数）
	cells    [][]*GridCell // セルの2次元配列
}

// GridCell は各グリッドセルに含まれるオブジェクト
type GridCell struct {
	playerSegments map[*models.Player][]*models.Position // プレイヤー別のセグメントポインタ
	food           []*models.Food                        // このセルに含まれる食べ物のポインタ
}

// NewSpatialGrid は新しい空間分割グリッドを作成する
func NewSpatialGrid() *SpatialGrid {
	// セルサイズを蛇の半径の4倍に設定（効率的な衝突判定のため）
	cellSize := utils.SNAKE_RADIUS * 4.0
	width := int(utils.FIELD_WIDTH/cellSize) + 1
	height := int(utils.FIELD_HEIGHT/cellSize) + 1

	// セル配列を初期化
	cells := make([][]*GridCell, height)
	for i := range cells {
		cells[i] = make([]*GridCell, width)
		for j := range cells[i] {
			cells[i][j] = &GridCell{
				playerSegments: make(map[*models.Player][]*models.Position),
				food:           make([]*models.Food, 0, 4),
			}
		}
	}

	return &SpatialGrid{
		cellSize: cellSize,
		width:    width,
		height:   height,
		cells:    cells,
	}
}

// Clear は全セルをクリアし、メモリを完全に解放する
func (sg *SpatialGrid) Clear() {
	for i := range sg.cells {
		for j := range sg.cells[i] {
			// 完全に新しいマップとスライスを作成してメモリリークを防ぐ
			sg.cells[i][j].playerSegments = make(map[*models.Player][]*models.Position)
			sg.cells[i][j].food = make([]*models.Food, 0, 4)
		}
	}
}

// GetCellCoords は座標からセルの座標を取得する
func (sg *SpatialGrid) GetCellCoords(x, y float64) (int, int) {
	cellX := int(x / sg.cellSize)
	cellY := int(y / sg.cellSize)

	// 境界チェック
	if cellX < 0 {
		cellX = 0
	} else if cellX >= sg.width {
		cellX = sg.width - 1
	}

	if cellY < 0 {
		cellY = 0
	} else if cellY >= sg.height {
		cellY = sg.height - 1
	}

	return cellX, cellY
}

// AddPlayerSegments はプレイヤーの全セグメントをグリッドに追加する
func (sg *SpatialGrid) AddPlayerSegments(player *models.Player, segments []models.Position) {
	for i := range segments {
		segment := &segments[i] // ポインタを取得
		cellX, cellY := sg.GetCellCoords(segment.X, segment.Y)

		// 安全性チェック
		if cellY >= 0 && cellY < sg.height && cellX >= 0 && cellX < sg.width {
			cell := sg.cells[cellY][cellX]
			cell.playerSegments[player] = append(cell.playerSegments[player], segment)
		}
	}
}

// AddFood は食べ物をグリッドに追加する
func (sg *SpatialGrid) AddFood(food *models.Food) {
	cellX, cellY := sg.GetCellCoords(food.Position.X, food.Position.Y)

	// 安全性チェック
	if cellY >= 0 && cellY < sg.height && cellX >= 0 && cellX < sg.width {
		sg.cells[cellY][cellX].food = append(sg.cells[cellY][cellX].food, food)
	}
}

// CheckCollisionAt は指定した位置で衝突しているプレイヤーを返す
func (sg *SpatialGrid) CheckCollisionAt(position models.Position, excludePlayer *models.Player) *models.Player {
	centerX, centerY := sg.GetCellCoords(position.X, position.Y)

	// 周囲9セル（3x3）をチェック
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			cellX := centerX + dx
			cellY := centerY + dy

			// 境界チェック
			if cellX >= 0 && cellX < sg.width && cellY >= 0 && cellY < sg.height {
				cell := sg.cells[cellY][cellX]

				// 各プレイヤーのセグメントをチェック
				for player, segments := range cell.playerSegments {
					if player == excludePlayer || !player.Snake.Alive {
						continue
					}

					// セグメントとの距離チェック
					for _, segment := range segments {
						dx := position.X - segment.X
						dy := position.Y - segment.Y
						dist := dx*dx + dy*dy
						if dist < utils.SNAKE_RADIUS*utils.SNAKE_RADIUS*4 { // 2*SNAKE_RADIUSの二乗
							return player
						}
					}
				}
			}
		}
	}

	return nil
}

// CheckFoodCollisionAt は指定した位置で衝突している食べ物を返す
func (sg *SpatialGrid) CheckFoodCollisionAt(position models.Position) *models.Food {
	centerX, centerY := sg.GetCellCoords(position.X, position.Y)

	// 周囲9セル（3x3）をチェック
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			cellX := centerX + dx
			cellY := centerY + dy

			// 境界チェック
			if cellX >= 0 && cellX < sg.width && cellY >= 0 && cellY < sg.height {
				cell := sg.cells[cellY][cellX]

				// 食べ物との距離チェック
				for _, food := range cell.food {
					dx := position.X - food.Position.X
					dy := position.Y - food.Position.Y
					dist := dx*dx + dy*dy
					if dist < (utils.SNAKE_RADIUS+utils.FOOD_RADIUS)*(utils.SNAKE_RADIUS+utils.FOOD_RADIUS) {
						return food
					}
				}
			}
		}
	}

	return nil
}

// GetNearbyFoodSafe は指定した位置の周囲の食べ物を安全に取得する
func (sg *SpatialGrid) GetNearbyFoodSafe(position models.Position) []*models.Food {
	centerX, centerY := sg.GetCellCoords(position.X, position.Y)

	nearbyFood := make([]*models.Food, 0, 10)

	// 周囲9セル（3x3）をチェック
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			cellX := centerX + dx
			cellY := centerY + dy

			// 境界チェック
			if cellX >= 0 && cellX < sg.width && cellY >= 0 && cellY < sg.height {
				cell := sg.cells[cellY][cellX]
				nearbyFood = append(nearbyFood, cell.food...)
			}
		}
	}

	return nearbyFood
}

// GetNearbyFoodInRadius は指定した位置の半径内の食べ物を取得する（NPC用）
func (sg *SpatialGrid) GetNearbyFoodInRadius(position models.Position, radius float64) []*models.Food {
	// 検索範囲を決定
	searchCells := int(radius/sg.cellSize) + 1
	centerX, centerY := sg.GetCellCoords(position.X, position.Y)

	nearbyFood := make([]*models.Food, 0, 20)

	// 指定した半径内のセルをチェック
	for dy := -searchCells; dy <= searchCells; dy++ {
		for dx := -searchCells; dx <= searchCells; dx++ {
			cellX := centerX + dx
			cellY := centerY + dy

			// 境界チェック
			if cellX >= 0 && cellX < sg.width && cellY >= 0 && cellY < sg.height {
				cell := sg.cells[cellY][cellX]
				nearbyFood = append(nearbyFood, cell.food...)
			}
		}
	}

	return nearbyFood
}

// IsPositionOccupiedOptimized は空間分割を使った効率的な占有チェック
func (sg *SpatialGrid) IsPositionOccupiedOptimized(pos models.Position) bool {
	centerX, centerY := sg.GetCellCoords(pos.X, pos.Y)

	// 周囲9セル（3x3）をチェック
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			cellX := centerX + dx
			cellY := centerY + dy

			// 境界チェック
			if cellX >= 0 && cellX < sg.width && cellY >= 0 && cellY < sg.height {
				cell := sg.cells[cellY][cellX]

				// 各プレイヤーのセグメントをチェック
				for _, segments := range cell.playerSegments {
					for _, segment := range segments {
						dx := segment.X - pos.X
						dy := segment.Y - pos.Y
						dist := dx*dx + dy*dy
						if dist < (utils.SNAKE_RADIUS+utils.FOOD_RADIUS)*(utils.SNAKE_RADIUS+utils.FOOD_RADIUS) {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// GetGridLines はSpatialGridの分割線を取得する
func (sg *SpatialGrid) GetGridLines() []models.GridLine {
	lines := make([]models.GridLine, 0, sg.width+sg.height)

	// 縦線（垂直線）
	for i := 1; i < sg.width; i++ {
		x := float64(i) * sg.cellSize
		lines = append(lines, models.GridLine{
			StartX: x,
			StartY: 0,
			EndX:   x,
			EndY:   utils.FIELD_HEIGHT,
		})
	}

	// 横線（水平線）
	for i := 1; i < sg.height; i++ {
		y := float64(i) * sg.cellSize
		lines = append(lines, models.GridLine{
			StartX: 0,
			StartY: y,
			EndX:   utils.FIELD_WIDTH,
			EndY:   y,
		})
	}

	return lines
}
