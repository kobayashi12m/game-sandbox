# Game Sandbox

リアルタイム同期のマルチプレイヤー対戦ゲーム。

## ゲーム内容

- プレイヤーは中心球と周囲を回転する球で構成
- 落ちている球を拾うと自分の球が増え、スコアも獲得
- 球を射出、または体当たりで敵を攻撃
- 球同士はぶつかると相殺して消滅
- 敵の中心球を破壊すると撃破、スコア獲得

## PRポイント

- WebSocketによるリアルタイム同期
- 空間分割による効率的な衝突判定
- 視野範囲カリングで通信量を削減
- 球の増減に応じて軌道を滑らかに再形成

## 操作方法

### PC

- **マウス移動**: カーソル方向に移動
- **クリック**: 球を射出して攻撃

### スマホ

- **タッチ&ドラッグ**: タッチ方向に移動

## セットアップ

### サーバー

```bash
cd server
go mod tidy
```

### クライアント

```bash
cd client
npm install
```

## 起動方法

### サーバー起動

通常起動:

```bash
cd server
go run main.go
```

Air でホットリロード開発:

```bash
# Airインストール（初回のみ）
go install github.com/cosmtrek/air@latest

# 起動
cd server
air
```

### クライアント起動

```bash
cd client
npm run dev
```

## アクセス方法

- PC: http://localhost:5173/
- スマホ: ターミナルに表示されるNetwork URL (例: http://192.168.x.x:5173/)
