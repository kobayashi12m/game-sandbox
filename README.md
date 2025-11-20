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

## 操作

- PC: 矢印キー または WASD
- スマホ: 画面下部の十字キー