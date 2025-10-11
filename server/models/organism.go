package models

import (
	"chess-mmo/server/utils"
	"math"
	"math/rand/v2"
	"time"
)

// PhysicsNode は物理演算される球体ノードを表す
type PhysicsNode struct {
	Position Position `json:"position"`
	Velocity Position `json:"velocity"`
	Radius   float64  `json:"radius"`
	Mass     float64  `json:"mass"`
}

// Connection は2つのノード間の制約（線）を表す
type Connection struct {
	NodeA      *PhysicsNode `json:"-"`          // ノードAへの参照
	NodeB      *PhysicsNode `json:"-"`          // ノードBへの参照
	RestLength float64      `json:"restLength"` // 自然長
	Stiffness  float64      `json:"stiffness"`  // ばね定数
	Damping    float64      `json:"damping"`    // ダンピング係数
}

// OrganismBody は新しい球体＋線構造のエンティティを表す
type OrganismBody struct {
	Core        *PhysicsNode   `json:"core"`        // 中心球
	Nodes       []*PhysicsNode `json:"nodes"`       // 周辺球
	Connections []*Connection  `json:"connections"` // 制約
	Color       string         `json:"color"`
	Alive       bool           `json:"alive"`
	Growing     int            `json:"-"`
	Respawning  bool           `json:"-"`
	DeathTime   time.Time      `json:"-"`

	// 新しい移動システム用
	Acceleration Position `json:"-"` // 現在の加速度
	MaxSpeed     float64  `json:"-"` // 最大速度
	AccelForce   float64  `json:"-"` // 加速力
	InputActive  bool     `json:"-"` // 入力が有効かどうか
}

// Reset は球体構造を初期状態に初期化する
func (o *OrganismBody) Reset() {
	// フィールド内のランダムな位置にスポーン
	startX := rand.Float64()*(utils.FIELD_WIDTH-100) + 50
	startY := rand.Float64()*(utils.FIELD_HEIGHT-100) + 50

	// コア（中心球）を初期化
	o.Core = &PhysicsNode{
		Position: Position{X: startX, Y: startY},
		Velocity: Position{X: 0, Y: 0},
		Radius:   utils.ORGANISM_RADIUS,
		Mass:     1.0,
	}

	// 初期ノード（原子構造：コアの周りに4個を円形配置）
	o.Nodes = []*PhysicsNode{}
	o.Connections = []*Connection{}

	// コアの周りに4個のノードを円形に配置（原子の電子軌道のように）
	nodeCount := 4
	orbitalRadius := utils.ORGANISM_RADIUS * utils.CONNECTION_NATURAL_RATIO // 自然長

	for i := 0; i < nodeCount; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(nodeCount) // 等間隔で配置

		nodeX := startX + orbitalRadius*math.Cos(angle)
		nodeY := startY + orbitalRadius*math.Sin(angle)

		node := &PhysicsNode{
			Position: Position{X: nodeX, Y: nodeY},
			Velocity: Position{X: 0, Y: 0},
			Radius:   utils.ORGANISM_RADIUS,
			Mass:     1.0,
		}

		// ノードをコアに接続
		connection := &Connection{
			NodeA:      o.Core,
			NodeB:      node,
			RestLength: orbitalRadius,
			Stiffness:  utils.CONNECTION_STIFFNESS,
			Damping:    utils.CONNECTION_DAMPING,
		}

		o.Nodes = append(o.Nodes, node)
		o.Connections = append(o.Connections, connection)
	}

	// ノード同士を環状に接続（隣接するノード同士を接続）
	// 正方形配置での隣接ノード間の理想距離を計算
	idealRingDistance := orbitalRadius * math.Sqrt(2) // 正方形の一辺の長さ（45度間隔の場合）

	for i := 0; i < len(o.Nodes); i++ {
		nextIndex := (i + 1) % len(o.Nodes) // 次のノード（最後は最初に戻る）

		ringConnection := &Connection{
			NodeA:      o.Nodes[i],
			NodeB:      o.Nodes[nextIndex],
			RestLength: idealRingDistance,          // 理想的な正方形配置での距離を自然長とする
			Stiffness:  utils.CONNECTION_STIFFNESS, // Core接続と同じ強さ
			Damping:    utils.CONNECTION_DAMPING,
		}

		o.Connections = append(o.Connections, ringConnection)
	}

	o.Growing = 0
	o.Alive = true
	o.Respawning = false

	// 新しい移動システムのパラメータ
	o.Acceleration = Position{X: 0, Y: 0}
	o.MaxSpeed = utils.ORGANISM_SPEED
	o.AccelForce = utils.ORGANISM_ACCEL_FORCE // 加速力
	o.InputActive = false
}

// Move は球体構造を物理シミュレーションで移動させる
func (o *OrganismBody) Move(deltaTime float64) {
	if !o.Alive {
		return
	}

	// 加速度ベースの移動システム
	if o.InputActive {
		// キーが押されている間は加速度を適用
		o.Core.Velocity.X += o.Acceleration.X * deltaTime
		o.Core.Velocity.Y += o.Acceleration.Y * deltaTime
	}

	// 最大速度制限
	speed := math.Sqrt(o.Core.Velocity.X*o.Core.Velocity.X + o.Core.Velocity.Y*o.Core.Velocity.Y)
	if speed > o.MaxSpeed {
		o.Core.Velocity.X = (o.Core.Velocity.X / speed) * o.MaxSpeed
		o.Core.Velocity.Y = (o.Core.Velocity.Y / speed) * o.MaxSpeed
	}

	// 全ノードの物理シミュレーション
	o.updatePhysics(deltaTime)

	// フィールド境界での衝突処理
	o.applyBoundaryCollision()
}

// updatePhysics は物理シミュレーションを実行する
func (o *OrganismBody) updatePhysics(deltaTime float64) {
	// 制約力を計算してノードに適用
	for _, conn := range o.Connections {
		nodeA := conn.NodeA
		nodeB := conn.NodeB

		if nodeA == nil || nodeB == nil {
			continue
		}

		// ばね力の計算
		dx := nodeB.Position.X - nodeA.Position.X
		dy := nodeB.Position.Y - nodeA.Position.Y
		currentLength := math.Sqrt(dx*dx + dy*dy)

		if currentLength > 0 {
			// 正規化された方向ベクトル
			nx := dx / currentLength
			ny := dy / currentLength

			// ばね力
			springForce := conn.Stiffness * (currentLength - conn.RestLength)

			// ダンピング力
			relVelX := nodeB.Velocity.X - nodeA.Velocity.X
			relVelY := nodeB.Velocity.Y - nodeA.Velocity.Y
			dampingForce := conn.Damping * (relVelX*nx + relVelY*ny)

			// 総力
			totalForce := springForce + dampingForce

			// 力をノードに適用
			forceX := totalForce * nx
			forceY := totalForce * ny

			// コア（中心球）には物理制約を適用しない（加速度のみで制御）
			if nodeA != o.Core {
				nodeA.Velocity.X += forceX * deltaTime / nodeA.Mass
				nodeA.Velocity.Y += forceY * deltaTime / nodeA.Mass
			}
			if nodeB != o.Core {
				nodeB.Velocity.X -= forceX * deltaTime / nodeB.Mass
				nodeB.Velocity.Y -= forceY * deltaTime / nodeB.Mass
			}
		}
	}

	// 位置の更新
	o.Core.Position.X += o.Core.Velocity.X * deltaTime
	o.Core.Position.Y += o.Core.Velocity.Y * deltaTime

	for i := range o.Nodes {
		o.Nodes[i].Position.X += o.Nodes[i].Velocity.X * deltaTime
		o.Nodes[i].Position.Y += o.Nodes[i].Velocity.Y * deltaTime
	}

	// 紐の長さ制限を強制適用
	o.enforceConnectionLimits()

	// ノード間の反発力を適用（重なり防止）
	o.applyNodeRepulsion(deltaTime)

	// 角度復元力を適用（絡まり防止）
	o.applyAngularRestoration(deltaTime)

	// 空気抵抗
	airResistance := utils.AIR_RESISTANCE
	o.Core.Velocity.X *= airResistance
	o.Core.Velocity.Y *= airResistance
	for i := range o.Nodes {
		o.Nodes[i].Velocity.X *= airResistance
		o.Nodes[i].Velocity.Y *= airResistance
	}

	// 低速時の停止判定（updatePhysics の最後で実行）
	if !o.InputActive {
		speed := math.Sqrt(o.Core.Velocity.X*o.Core.Velocity.X + o.Core.Velocity.Y*o.Core.Velocity.Y)
		stopThreshold := o.MaxSpeed * utils.STOP_THRESHOLD_RATIO

		if speed < stopThreshold {
			o.Core.Velocity.X = 0
			o.Core.Velocity.Y = 0
		}
	}
}

// applyBoundaryCollision はフィールド境界での衝突処理を適用
func (o *OrganismBody) applyBoundaryCollision() {
	// コアの境界衝突処理
	if o.Core.Position.X-o.Core.Radius < 0 {
		o.Core.Position.X = o.Core.Radius
		o.Core.Velocity.X = -o.Core.Velocity.X * 0.5 // 反発係数0.5
	} else if o.Core.Position.X+o.Core.Radius >= utils.FIELD_WIDTH {
		o.Core.Position.X = utils.FIELD_WIDTH - o.Core.Radius
		o.Core.Velocity.X = -o.Core.Velocity.X * 0.5
	}
	if o.Core.Position.Y-o.Core.Radius < 0 {
		o.Core.Position.Y = o.Core.Radius
		o.Core.Velocity.Y = -o.Core.Velocity.Y * 0.5
	} else if o.Core.Position.Y+o.Core.Radius >= utils.FIELD_HEIGHT {
		o.Core.Position.Y = utils.FIELD_HEIGHT - o.Core.Radius
		o.Core.Velocity.Y = -o.Core.Velocity.Y * 0.5
	}

	// ノードの境界衝突処理
	for i := range o.Nodes {
		node := o.Nodes[i]
		if node.Position.X-node.Radius < 0 {
			node.Position.X = node.Radius
			node.Velocity.X = -node.Velocity.X * 0.5
		} else if node.Position.X+node.Radius >= utils.FIELD_WIDTH {
			node.Position.X = utils.FIELD_WIDTH - node.Radius
			node.Velocity.X = -node.Velocity.X * 0.5
		}
		if node.Position.Y-node.Radius < 0 {
			node.Position.Y = node.Radius
			node.Velocity.Y = -node.Velocity.Y * 0.5
		} else if node.Position.Y+node.Radius >= utils.FIELD_HEIGHT {
			node.Position.Y = utils.FIELD_HEIGHT - node.Radius
			node.Velocity.Y = -node.Velocity.Y * 0.5
		}
	}
}

// SetAcceleration は加速度を直接設定する（360度自由移動用）
func (o *OrganismBody) SetAcceleration(x, y float64) {
	o.Acceleration.X = x * o.AccelForce
	o.Acceleration.Y = y * o.AccelForce
	o.InputActive = (x != 0 || y != 0)
}

// AddNode は新しいノードを追加する（成長時）- 原子構造を維持
func (o *OrganismBody) AddNode() {
	// 自然長を定数から取得
	orbitalRadius := utils.ORGANISM_RADIUS * utils.CONNECTION_NATURAL_RATIO

	// ランダムな角度で新しいノードを追加（既存ノードとの重複を避ける）
	angle := rand.Float64() * 2.0 * math.Pi

	coreX := o.Core.Position.X
	coreY := o.Core.Position.Y
	nodeX := coreX + orbitalRadius*math.Cos(angle)
	nodeY := coreY + orbitalRadius*math.Sin(angle)

	newNode := &PhysicsNode{
		Position: Position{X: nodeX, Y: nodeY},
		Velocity: Position{X: 0, Y: 0},
		Radius:   utils.ORGANISM_RADIUS,
		Mass:     1.0,
	}

	// 新しいノードをコアに直接接続（原子構造）
	newConnection := &Connection{
		NodeA:      o.Core,
		NodeB:      newNode,
		RestLength: orbitalRadius,
		Stiffness:  utils.CONNECTION_STIFFNESS,
		Damping:    utils.CONNECTION_DAMPING,
	}

	o.Nodes = append(o.Nodes, newNode)
	o.Connections = append(o.Connections, newConnection)
}

// enforceConnectionLimits は紐の長さを強制的に制限する
func (o *OrganismBody) enforceConnectionLimits() {
	maxLength := utils.ORGANISM_RADIUS * utils.CONNECTION_MAX_RATIO // 最大長を定数から取得

	for _, conn := range o.Connections {
		if conn.NodeA == nil || conn.NodeB == nil {
			continue
		}

		// 現在の距離を計算
		dx := conn.NodeB.Position.X - conn.NodeA.Position.X
		dy := conn.NodeB.Position.Y - conn.NodeA.Position.Y
		currentLength := math.Sqrt(dx*dx + dy*dy)

		// 最大長を超えている場合は強制的に短縮
		if currentLength > maxLength {
			// 正規化された方向ベクトル
			nx := dx / currentLength
			ny := dy / currentLength

			// 最大長の位置に強制移動（コアは動かさず、ノードのみ移動）
			if conn.NodeA == o.Core {
				// コアから見て最大長の位置にノードを配置
				conn.NodeB.Position.X = conn.NodeA.Position.X + nx*maxLength
				conn.NodeB.Position.Y = conn.NodeA.Position.Y + ny*maxLength
			} else if conn.NodeB == o.Core {
				// ノードからコアへの場合（通常はない）
				conn.NodeA.Position.X = conn.NodeB.Position.X - nx*maxLength
				conn.NodeA.Position.Y = conn.NodeB.Position.Y - ny*maxLength
			}
		}
	}
}

// applyNodeRepulsion はノード間の反発力を適用する（重なり防止）
func (o *OrganismBody) applyNodeRepulsion(deltaTime float64) {
	minDistance := utils.ORGANISM_RADIUS * 2.5 // 最小距離（半径の2.5倍）

	// 全ノードペアについて反発力をチェック
	for i := 0; i < len(o.Nodes); i++ {
		for j := i + 1; j < len(o.Nodes); j++ {
			nodeA := o.Nodes[i]
			nodeB := o.Nodes[j]

			// 距離を計算
			dx := nodeB.Position.X - nodeA.Position.X
			dy := nodeB.Position.Y - nodeA.Position.Y
			distance := math.Sqrt(dx*dx + dy*dy)

			// 最小距離より近い場合は反発力を適用
			if distance < minDistance && distance > 0 {
				// 正規化された方向ベクトル
				nx := dx / distance
				ny := dy / distance

				// 反発力の強さ（距離が近いほど強く）
				repulsionStrength := utils.NODE_REPULSION_FORCE * (minDistance - distance) / minDistance

				// 反発力を適用
				forceX := nx * repulsionStrength
				forceY := ny * repulsionStrength

				nodeA.Velocity.X -= forceX * deltaTime / nodeA.Mass
				nodeA.Velocity.Y -= forceY * deltaTime / nodeA.Mass
				nodeB.Velocity.X += forceX * deltaTime / nodeB.Mass
				nodeB.Velocity.Y += forceY * deltaTime / nodeB.Mass
			}
		}
	}
}

// applyAngularRestoration は各ノードを理想的な角度位置に復元する（絡まり防止）
func (o *OrganismBody) applyAngularRestoration(deltaTime float64) {
	if len(o.Nodes) != 4 {
		return // 4ノード構成のみ対応
	}

	orbitalRadius := utils.ORGANISM_RADIUS * utils.CONNECTION_NATURAL_RATIO

	for i := 0; i < 4; i++ {
		// 理想的な角度（90度間隔）
		idealAngle := float64(i) * 2.0 * math.Pi / 4.0

		// 理想的な位置
		idealX := o.Core.Position.X + orbitalRadius*math.Cos(idealAngle)
		idealY := o.Core.Position.Y + orbitalRadius*math.Sin(idealAngle)

		// 現在位置との差
		deltaX := idealX - o.Nodes[i].Position.X
		deltaY := idealY - o.Nodes[i].Position.Y

		// 角度復元力を適用
		forceX := deltaX * utils.ANGULAR_RESTORATION_FORCE
		forceY := deltaY * utils.ANGULAR_RESTORATION_FORCE

		o.Nodes[i].Velocity.X += forceX * deltaTime / o.Nodes[i].Mass
		o.Nodes[i].Velocity.Y += forceY * deltaTime / o.Nodes[i].Mass
	}
}
