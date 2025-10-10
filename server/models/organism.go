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
	Core        *PhysicsNode    `json:"core"`        // 中心球
	Nodes       []*PhysicsNode  `json:"nodes"`       // 周辺球
	Connections []*Connection   `json:"connections"` // 制約
	Color       string          `json:"color"`
	Alive       bool            `json:"alive"`
	Growing     int             `json:"-"`
	Respawning  bool            `json:"-"`
	DeathTime   time.Time       `json:"-"`
	Speed       float64         `json:"-"` // 移動速度（コアの基準速度）
	Direction   utils.Direction `json:"-"` // 現在の移動方向
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
		Radius:   utils.SNAKE_RADIUS,
		Mass:     1.0,
	}

	// 初期ノード（2個）
	o.Nodes = []*PhysicsNode{
		{
			Position: Position{X: startX - utils.SNAKE_RADIUS*3, Y: startY},
			Velocity: Position{X: 0, Y: 0},
			Radius:   utils.SNAKE_RADIUS * 0.8,
			Mass:     0.8,
		},
		{
			Position: Position{X: startX - utils.SNAKE_RADIUS*6, Y: startY},
			Velocity: Position{X: 0, Y: 0},
			Radius:   utils.SNAKE_RADIUS * 0.6,
			Mass:     0.6,
		},
	}

	// 接続（制約）
	o.Connections = []*Connection{
		{
			NodeA:      o.Core,
			NodeB:      o.Nodes[0],
			RestLength: utils.SNAKE_RADIUS * 2.5,
			Stiffness:  0.8,
			Damping:    0.3,
		},
		{
			NodeA:      o.Nodes[0],
			NodeB:      o.Nodes[1],
			RestLength: utils.SNAKE_RADIUS * 2.5,
			Stiffness:  0.7,
			Damping:    0.3,
		},
	}

	o.Direction = utils.DIRECTIONS["RIGHT"]
	o.Growing = 0
	o.Alive = true
	o.Respawning = false
	o.Speed = utils.SNAKE_SPEED
}

// Move は球体構造を物理シミュレーションで移動させる
func (o *OrganismBody) Move(deltaTime float64) {
	if !o.Alive {
		return
	}

	// コアを目標方向に移動（プレイヤー入力による駆動力）
	targetVel := Position{
		X: o.Direction.X * o.Speed,
		Y: o.Direction.Y * o.Speed,
	}

	// コアの速度を徐々に目標速度に近づける
	damping := 0.1
	o.Core.Velocity.X += (targetVel.X - o.Core.Velocity.X) * damping
	o.Core.Velocity.Y += (targetVel.Y - o.Core.Velocity.Y) * damping

	// 全ノードの物理シミュレーション
	o.updatePhysics(deltaTime)

	// フィールド境界でのラップアラウンド
	o.applyBoundaryWrapping()
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

			nodeA.Velocity.X += forceX * deltaTime / nodeA.Mass
			nodeA.Velocity.Y += forceY * deltaTime / nodeA.Mass
			nodeB.Velocity.X -= forceX * deltaTime / nodeB.Mass
			nodeB.Velocity.Y -= forceY * deltaTime / nodeB.Mass
		}
	}

	// 位置の更新
	o.Core.Position.X += o.Core.Velocity.X * deltaTime
	o.Core.Position.Y += o.Core.Velocity.Y * deltaTime

	for i := range o.Nodes {
		o.Nodes[i].Position.X += o.Nodes[i].Velocity.X * deltaTime
		o.Nodes[i].Position.Y += o.Nodes[i].Velocity.Y * deltaTime
	}

	// 空気抵抗
	airResistance := 0.95
	o.Core.Velocity.X *= airResistance
	o.Core.Velocity.Y *= airResistance
	for i := range o.Nodes {
		o.Nodes[i].Velocity.X *= airResistance
		o.Nodes[i].Velocity.Y *= airResistance
	}
}

// applyBoundaryWrapping はフィールド境界でのラップアラウンドを適用
func (o *OrganismBody) applyBoundaryWrapping() {
	// コアのラップアラウンド
	if o.Core.Position.X < 0 {
		o.Core.Position.X += utils.FIELD_WIDTH
	} else if o.Core.Position.X >= utils.FIELD_WIDTH {
		o.Core.Position.X -= utils.FIELD_WIDTH
	}
	if o.Core.Position.Y < 0 {
		o.Core.Position.Y += utils.FIELD_HEIGHT
	} else if o.Core.Position.Y >= utils.FIELD_HEIGHT {
		o.Core.Position.Y -= utils.FIELD_HEIGHT
	}

	// ノードのラップアラウンド
	for i := range o.Nodes {
		if o.Nodes[i].Position.X < 0 {
			o.Nodes[i].Position.X += utils.FIELD_WIDTH
		} else if o.Nodes[i].Position.X >= utils.FIELD_WIDTH {
			o.Nodes[i].Position.X -= utils.FIELD_WIDTH
		}
		if o.Nodes[i].Position.Y < 0 {
			o.Nodes[i].Position.Y += utils.FIELD_HEIGHT
		} else if o.Nodes[i].Position.Y >= utils.FIELD_HEIGHT {
			o.Nodes[i].Position.Y -= utils.FIELD_HEIGHT
		}
	}
}

// ChangeDirection は移動方向を変更する
func (o *OrganismBody) ChangeDirection(newDir utils.Direction) {
	o.Direction = newDir
}

// AddNode は新しいノードを追加する（成長時）
func (o *OrganismBody) AddNode() {
	if len(o.Nodes) == 0 {
		return
	}

	// 最後のノードから更に離れた位置に新ノードを追加
	lastNode := o.Nodes[len(o.Nodes)-1]

	newNode := &PhysicsNode{
		Position: Position{
			X: lastNode.Position.X - utils.SNAKE_RADIUS*2,
			Y: lastNode.Position.Y,
		},
		Velocity: Position{X: 0, Y: 0},
		Radius:   utils.SNAKE_RADIUS * 0.5,
		Mass:     0.5,
	}

	// 新しい接続
	newConnection := &Connection{
		NodeA:      lastNode,
		NodeB:      newNode,
		RestLength: utils.SNAKE_RADIUS * 2.0,
		Stiffness:  0.6,
		Damping:    0.3,
	}

	o.Nodes = append(o.Nodes, newNode)
	o.Connections = append(o.Connections, newConnection)
}
