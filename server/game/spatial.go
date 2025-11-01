package game

import (
	"game-sandbox/server/models"
	"game-sandbox/server/utils"
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
}

// NewSpatialGrid は新しい空間分割グリッドを作成する
func NewSpatialGrid() *SpatialGrid {
	// セルサイズの設定
	cellSize := utils.SPHERE_RADIUS * 4.0
	width := int(utils.FIELD_WIDTH/cellSize) + 1
	height := int(utils.FIELD_HEIGHT/cellSize) + 1

	// セル配列を初期化
	cells := make([][]*GridCell, height)
	for i := range cells {
		cells[i] = make([]*GridCell, width)
		for j := range cells[i] {
			cells[i][j] = &GridCell{
				playerSegments: make(map[*models.Player][]*models.Position),
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
			// 完全に新しいマップを作成してメモリリークを防ぐ
			sg.cells[i][j].playerSegments = make(map[*models.Player][]*models.Position)
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


// iterateNearbyCells は指定位置周辺のセルに対してコールバック関数を実行する
func (sg *SpatialGrid) iterateNearbyCells(centerX, centerY, radius int, callback func(*GridCell)) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			cellX := centerX + dx
			cellY := centerY + dy

			// 境界チェック
			if cellX >= 0 && cellX < sg.width && cellY >= 0 && cellY < sg.height {
				callback(sg.cells[cellY][cellX])
			}
		}
	}
}

// CheckCollisionAt は指定した位置で衝突しているプレイヤーを返す
func (sg *SpatialGrid) CheckCollisionAt(position models.Position, excludePlayer *models.Player) *models.Player {
	centerX, centerY := sg.GetCellCoords(position.X, position.Y)

	var result *models.Player
	sg.iterateNearbyCells(centerX, centerY, 1, func(cell *GridCell) {
		if result != nil {
			return
		}
		for player, segments := range cell.playerSegments {
			if player == excludePlayer || !player.Celestial.Alive {
				continue
			}
			for _, segment := range segments {
				dx := position.X - segment.X
				dy := position.Y - segment.Y
				dist := dx*dx + dy*dy
				if dist < (utils.SPHERE_RADIUS*2)*(utils.SPHERE_RADIUS*2) {
					result = player
					return
				}
			}
		}
	})

	return result
}


// AreaResult はエリア内のプレイヤーをまとめて返す構造体
type AreaResult struct {
	Players []*models.Player
}

// GetObjectsInArea は指定した矩形エリア内のプレイヤーを取得する
func (sg *SpatialGrid) GetObjectsInArea(minX, maxX, minY, maxY float64) AreaResult {
	// エリアが含まれるセル範囲を計算
	startCellX, startCellY := sg.GetCellCoords(minX, minY)
	endCellX, endCellY := sg.GetCellCoords(maxX, maxY)

	playerSet := make(map[*models.Player]bool)

	// 指定したエリアのセルを一度だけスキャン
	for cellY := startCellY; cellY <= endCellY; cellY++ {
		for cellX := startCellX; cellX <= endCellX; cellX++ {
			// 境界チェック
			if cellX >= 0 && cellX < sg.width && cellY >= 0 && cellY < sg.height {
				cell := sg.cells[cellY][cellX]

				// プレイヤーのセグメントをチェック（死んだプレイヤーも含む）
				for player := range cell.playerSegments {
					playerSet[player] = true
				}
			}
		}
	}

	// プレイヤーセットをスライスに変換
	visiblePlayers := make([]*models.Player, 0, len(playerSet))
	for player := range playerSet {
		visiblePlayers = append(visiblePlayers, player)
	}

	return AreaResult{
		Players: visiblePlayers,
	}
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
