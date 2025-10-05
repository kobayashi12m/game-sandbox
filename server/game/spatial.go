package game

import (
	"chess-mmo/server/models"
	"chess-mmo/server/utils"
)

// SpatialGrid はゲームフィールドを格子状に分割して効率的な当たり判定を行う
type SpatialGrid struct {
	cellSize float64              // 各セルのサイズ
	width    int                  // グリッドの幅（セル数）
	height   int                  // グリッドの高さ（セル数）
	cells    [][]*GridCell        // セルの2次元配列
}

// GridCell は各グリッドセルに含まれるオブジェクト
type GridCell struct {
	players []string      // このセルに含まれるプレイヤーID
	food    []int         // このセルに含まれる食べ物のインデックス
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
				players: make([]string, 0, 4),
				food:    make([]int, 0, 4),
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
			// 完全に新しいスライスを作成してメモリリークを防ぐ
			sg.cells[i][j].players = make([]string, 0, 4)
			sg.cells[i][j].food = make([]int, 0, 4)
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

// AddPlayer はプレイヤーの全セグメントをグリッドに追加する
func (sg *SpatialGrid) AddPlayerSegments(playerID string, segments []models.Position) {
	for _, segment := range segments {
		cellX, cellY := sg.GetCellCoords(segment.X, segment.Y)
		
		// 安全性チェック
		if cellY >= 0 && cellY < sg.height && cellX >= 0 && cellX < sg.width {
			sg.cells[cellY][cellX].players = append(sg.cells[cellY][cellX].players, playerID)
		}
	}
}

// AddFood は食べ物をグリッドに追加する
func (sg *SpatialGrid) AddFood(foodIndex int, position models.Position) {
	cellX, cellY := sg.GetCellCoords(position.X, position.Y)
	
	// 安全性チェック
	if cellY >= 0 && cellY < sg.height && cellX >= 0 && cellX < sg.width {
		sg.cells[cellY][cellX].food = append(sg.cells[cellY][cellX].food, foodIndex)
	}
}

// GetNearbyPlayersUnique は指定した位置の周囲のプレイヤーIDを重複なしで取得する
func (sg *SpatialGrid) GetNearbyPlayersUnique(position models.Position) []string {
	centerX, centerY := sg.GetCellCoords(position.X, position.Y)
	
	playerSet := make(map[string]bool)
	
	// 周囲9セル（3x3）をチェック
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			cellX := centerX + dx
			cellY := centerY + dy
			
			// 境界チェック
			if cellX >= 0 && cellX < sg.width && cellY >= 0 && cellY < sg.height {
				cell := sg.cells[cellY][cellX]
				for _, playerID := range cell.players {
					playerSet[playerID] = true
				}
			}
		}
	}
	
	// セットをスライスに変換
	nearbyPlayers := make([]string, 0, len(playerSet))
	for playerID := range playerSet {
		nearbyPlayers = append(nearbyPlayers, playerID)
	}
	
	return nearbyPlayers
}

// GetNearbyFoodSafe は指定した位置の周囲の食べ物を安全に取得する
func (sg *SpatialGrid) GetNearbyFoodSafe(position models.Position, foodArray []models.Position) []models.Position {
	centerX, centerY := sg.GetCellCoords(position.X, position.Y)
	
	nearbyFood := make([]models.Position, 0, 10)
	
	// 周囲9セル（3x3）をチェック
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			cellX := centerX + dx
			cellY := centerY + dy
			
			// 境界チェック
			if cellX >= 0 && cellX < sg.width && cellY >= 0 && cellY < sg.height {
				cell := sg.cells[cellY][cellX]
				for _, foodIndex := range cell.food {
					if foodIndex >= 0 && foodIndex < len(foodArray) {
						nearbyFood = append(nearbyFood, foodArray[foodIndex])
					}
				}
			}
		}
	}
	
	return nearbyFood
}

// GetNearbyFoodInRadius は指定した位置の半径内の食べ物を取得する（NPC用）
func (sg *SpatialGrid) GetNearbyFoodInRadius(position models.Position, radius float64) []int {
	// 検索範囲を決定
	searchCells := int(radius/sg.cellSize) + 1
	centerX, centerY := sg.GetCellCoords(position.X, position.Y)
	
	nearbyFood := make([]int, 0, 20)
	
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
func (sg *SpatialGrid) IsPositionOccupiedOptimized(pos models.Position, players map[string]*models.Player) bool {
	nearbyPlayerIDs := sg.GetNearbyPlayersUnique(pos)
	
	for _, playerID := range nearbyPlayerIDs {
		if player, exists := players[playerID]; exists {
			for _, segment := range player.Snake.Body {
				dx := segment.X - pos.X
				dy := segment.Y - pos.Y
				dist := dx*dx + dy*dy
				if dist < (utils.SNAKE_RADIUS+utils.FOOD_RADIUS)*(utils.SNAKE_RADIUS+utils.FOOD_RADIUS) {
					return true
				}
			}
		}
	}
	
	return false
}