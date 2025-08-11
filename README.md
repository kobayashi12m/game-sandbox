# Chess MMO - ゲーム仕様書

## ゲームコンセプト

### 基本コンセプト
リアルタイムマルチプレイヤーチェスバトルロワイヤルゲーム。一つの大きな盤面で全プレイヤーが同時にプレイし、他プレイヤーの駒を取ることで成長していく永続世界ゲーム。

### ゲームの特徴
- **リアルタイム**: ターン制なし、全プレイヤーが同時に移動可能
- **永続世界**: 終了条件なし、常に変化し続ける戦場
- **成長システム**: 駒を取ることでクラスアップ
- **高リスク・ハイリターン**: 強くなるほど失うものが大きい

## ゲーム仕様

### 基本システム

#### プレイヤーライフサイクル
1. **参加時**: ポーンとしてランダム位置にスポーン
2. **陣営**: 白・黒にランダム割り当て（完全運任せ）
3. **成長**: 他プレイヤーを取って上位クラスに昇格
4. **死亡**: 取られたら即退場
5. **復帰**: 再参加時はポーンからリスタート

#### 盤面システム
- **サイズ**: 大きめの盤面（プロトタイプ: 50×50、最終: 200×200以上）
- **共有**: 全プレイヤーが同一盤面でプレイ
- **視野制限**: プレイヤー周辺のみ表示（大盤面対応）

#### 成長システム
```
駒の昇格ルート：
ポーン → ナイト → ビショップ → ルーク → クイーン

昇格条件：
- ポーン→ナイト: 3キル
- ナイト→ビショップ: 5キル  
- ビショップ→ルーク: 8キル
- ルーク→クイーン: 12キル
```

#### 戦闘システム
- **移動ルール**: 標準チェスルールに準拠
- **キル条件**: 敵駒の位置に移動することで撃破
- **リスポーン**: なし（即退場）
- **陣営バランス**: 運任せ（時には大きく偏ることも）

### ゲームフロー

#### 新規プレイヤー参加
1. ゲームに接続
2. 白・黒陣営にランダム割り当て
3. ポーンとして盤面にスポーン
4. 他プレイヤーとの戦闘開始

#### 通常プレイ
1. 盤面を移動して敵を探索
2. 敵駒を発見したら攻撃
3. 勝利すればキル数増加
4. 一定キル数で上位クラスに昇格
5. より強力な駒として継続プレイ

#### 敗北・復帰
1. 敵に取られて即退場
2. 再参加を選択
3. ポーンとして新たにスポーン
4. 再び成長を目指す

## 技術仕様

### アーキテクチャ概要

#### サーバーサイド
- **言語**: Go（高パフォーマンス、並行処理）
- **WebSocket**: Gorilla WebSocket（リアルタイム通信）
- **状態管理**: Redis（盤面状態、プレイヤー情報）
- **永続化**: PostgreSQL（統計、履歴）

#### クライアントサイド  
- **フレームワーク**: React + TypeScript
- **描画**: HTML5 Canvas（スムーズな盤面描画）
- **通信**: Socket.io-client（WebSocket通信）
- **状態管理**: React State + 楽観的アップデート

### データ構造

#### 駒データ
```go
type Piece struct {
    ID       string    `json:"id"`
    Type     PieceType `json:"type"`     // pawn, knight, bishop, rook, queen
    Owner    string    `json:"owner"`    // プレイヤーID
    Team     string    `json:"team"`     // white, black
    Position Position  `json:"position"` // x, y座標
    Kills    int       `json:"kills"`    // キル数
    Created  time.Time `json:"created"`  // 作成時刻
}
```

#### プレイヤーデータ
```go
type Player struct {
    ID       string    `json:"id"`
    Name     string    `json:"name"`
    Team     string    `json:"team"`     // white, black
    PieceID  string    `json:"pieceId"`  // 現在の駒ID
    IsAlive  bool      `json:"isAlive"`
    Stats    PlayerStats `json:"stats"`
}

type PlayerStats struct {
    TotalKills   int `json:"totalKills"`
    Deaths       int `json:"deaths"`
    MaxRank      PieceType `json:"maxRank"`
    PlayTime     int64 `json:"playTime"`
}
```

#### 盤面データ
```go
type Board struct {
    Size     int                    `json:"size"`      // 盤面サイズ
    Pieces   map[string]*Piece      `json:"pieces"`    // 駒情報
    Players  map[string]*Player     `json:"players"`   // プレイヤー情報
    LastMove time.Time             `json:"lastMove"`  // 最終移動時刻
}
```

### Redis データ設計

#### キー設計
```
# 盤面状態
board:pieces:{x}:{y} → piece_data
board:size → board_size

# プレイヤー情報
player:{id} → player_data
player:alive → set of alive player IDs

# ゲーム統計
game:stats:total_players → total player count
game:stats:active_players → current active players
game:stats:total_kills → total kills in game

# セッション管理
session:{id} → session_data
```

### 通信プロトコル

#### WebSocket メッセージ

##### クライアント → サーバー
```json
// プレイヤー参加
{
  "type": "join_game",
  "data": {
    "playerName": "Player1"
  }
}

// 駒移動
{
  "type": "move_piece", 
  "data": {
    "fromX": 10,
    "fromY": 10,
    "toX": 10,
    "toY": 9
  }
}

// 視野変更
{
  "type": "viewport_change",
  "data": {
    "centerX": 25,
    "centerY": 25,
    "width": 20,
    "height": 20
  }
}
```

##### サーバー → クライアント
```json
// 初期ゲーム状態
{
  "type": "game_state",
  "data": {
    "board": { /* board data */ },
    "player": { /* player data */ },
    "viewport": { /* visible area */ }
  }
}

// 駒移動通知
{
  "type": "piece_moved",
  "data": {
    "pieceId": "piece123",
    "fromX": 10,
    "fromY": 10, 
    "toX": 10,
    "toY": 9,
    "timestamp": 1642567890
  }
}

// 駒撃破通知
{
  "type": "piece_captured",
  "data": {
    "attackerId": "piece123",
    "victimId": "piece456",
    "position": {"x": 15, "y": 15},
    "timestamp": 1642567890
  }
}

// プレイヤー昇格通知
{
  "type": "player_promoted",
  "data": {
    "playerId": "player123",
    "oldType": "pawn",
    "newType": "knight",
    "kills": 3
  }
}

// プレイヤー退場通知
{
  "type": "player_eliminated",
  "data": {
    "playerId": "player123",
    "killedBy": "player456",
    "finalRank": "bishop",
    "totalKills": 7
  }
}
```

### パフォーマンス要件

#### 目標性能
- **同時接続数**: 1,000人以上
- **レスポンス時間**: <50ms（移動レスポンス）
- **フレームレート**: 60FPS（クライアント描画）
- **メモリ使用量**: <2GB（サーバー）

#### 最適化戦略
- **視野ベース配信**: プレイヤー周辺のみ配信
- **移動バッチ処理**: 複数移動をまとめて送信
- **差分アップデート**: 変更部分のみ送信
- **レンダリング最適化**: Canvas部分描画

## 開発計画

### プロトタイプ（2-3週間）
- [x] プロジェクト初期化
- [ ] 基本WebSocket通信
- [ ] 最小盤面（20×20）
- [ ] ポーン移動のみ
- [ ] 基本的な取る/取られる
- [ ] 簡単なUI

### MVP版（1-2ヶ月）
- [ ] 全駒種類実装
- [ ] 成長システム完成
- [ ] 中サイズ盤面（50×50）
- [ ] Redis統合
- [ ] 基本統計機能

### 本格版（3-6ヶ月）
- [ ] 大盤面（200×200+）
- [ ] PostgreSQL統合
- [ ] 詳細統計・ランキング
- [ ] 管理画面
- [ ] パフォーマンス最適化

### 拡張版（将来）
- [ ] 複数盤面
- [ ] チーム戦モード
- [ ] 特殊駒種類
- [ ] イベントシステム
- [ ] モバイル対応

## 技術的課題と対策

### リアルタイム同期
**課題**: 数百人の同時移動での競合処理
**対策**: タイムスタンプベースの優先制御、楽観的ロック

### スケーラビリティ  
**課題**: プレイヤー数増加時の性能劣化
**対策**: 水平分散、盤面分割、CDN活用

### 状態管理
**課題**: 大きな盤面状態の効率的管理
**対策**: Redis活用、差分更新、視野ベース配信

### ネットワーク遅延
**課題**: プレイヤー間の接続遅延差
**対策**: 予測移動、ラグ補償、地域別サーバー

## まとめ

Chess MMOは従来のチェスゲームに**大規模マルチプレイヤー**、**リアルタイム性**、**成長要素**を組み合わせた革新的なゲームコンセプトです。

技術的には challenging ですが、適切なアーキテクチャ設計により実現可能であり、プレイヤーに新しいゲーム体験を提供できると確信しています。