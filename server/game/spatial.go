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
	playerSpheres map[*models.Player][]*models.Sphere // プレイヤー別の球体ポインタ
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
				playerSpheres: make(map[*models.Player][]*models.Sphere),
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
			sg.cells[i][j].playerSpheres = make(map[*models.Player][]*models.Sphere)
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

// AddPlayerSpheres はプレイヤーの全球体をグリッドに追加する
func (sg *SpatialGrid) AddPlayerSpheres(player *models.Player, spheres []*models.Sphere) {
	for _, sphere := range spheres {
		cellX, cellY := sg.GetCellCoords(sphere.Position.X, sphere.Position.Y)

		// 安全性チェック
		if cellY >= 0 && cellY < sg.height && cellX >= 0 && cellX < sg.width {
			cell := sg.cells[cellY][cellX]
			cell.playerSpheres[player] = append(cell.playerSpheres[player], sphere)
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

// CheckCollisionAt は指定した位置で衝突しているプレイヤーと球体を返す
func (sg *SpatialGrid) CheckCollisionAt(position models.Position, radius float64, excludePlayer *models.Player) (*models.Player, *models.Sphere) {
	centerX, centerY := sg.GetCellCoords(position.X, position.Y)

	var resultPlayer *models.Player
	var resultSphere *models.Sphere
	sg.iterateNearbyCells(centerX, centerY, 1, func(cell *GridCell) {
		if resultPlayer != nil {
			return
		}
		for player, spheres := range cell.playerSpheres {
			if player == excludePlayer || !player.Celestial.Alive {
				continue
			}
			for _, sphere := range spheres {
				dx := position.X - sphere.Position.X
				dy := position.Y - sphere.Position.Y
				dist := dx*dx + dy*dy
				collisionDist := (radius + sphere.Radius) * (radius + sphere.Radius)
				if dist < collisionDist {
					resultPlayer = player
					resultSphere = sphere
					return
				}
			}
		}
	})

	return resultPlayer, resultSphere
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

				// プレイヤーの球体をチェック（死んだプレイヤーも含む）
				for player := range cell.playerSpheres {
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
