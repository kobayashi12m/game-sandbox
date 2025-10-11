import React, { useRef, useEffect, useState, useMemo, memo } from "react";
import type {
  GameState,
  GameConfig,
  Position,
  Player,
  GridLine,
} from "../types";

interface GameCanvasProps {
  gameState: GameState;
  playerId: string;
  gameConfig: GameConfig;
  onMouseMove: (x: number, y: number) => void;
}

const GameCanvas: React.FC<GameCanvasProps> = memo(
  ({ gameState, playerId, gameConfig, onMouseMove }) => {
    const canvasRef = useRef<HTMLCanvasElement>(null);
    const [showGrid, setShowGrid] = useState(true);
    const [showCulling, setShowCulling] = useState(true);
    // サーバー設定を基に固定サイズを計算（レイアウトシフト防止）
    const canvasSize = useMemo(() => {
      // 画面いっぱいにキャンバスを表示
      const windowWidth = window.innerWidth;
      const windowHeight = window.innerHeight;

      return { width: windowWidth, height: windowHeight };
    }, []);

    const frameCountRef = useRef(0);
    const lastLogTimeRef = useRef(Date.now());

    // キーボードショートカット
    useEffect(() => {
      const handleKeyPress = (event: KeyboardEvent) => {
        if (event.key === "g" || event.key === "G") {
          setShowGrid((prev) => !prev);
        }
        if (event.key === "c" || event.key === "C") {
          setShowCulling((prev) => !prev);
        }
      };

      window.addEventListener("keydown", handleKeyPress);
      return () => window.removeEventListener("keydown", handleKeyPress);
    }, []);

    // マウス移動の処理
    useEffect(() => {
      const canvas = canvasRef.current;
      if (!canvas) return;

      let animationFrameId: number | null = null;
      let pendingMouseEvent: MouseEvent | null = null;
      let lastSendTime = 0;
      const SEND_INTERVAL = 33; // 約30fps（33ms間隔）

      const processMouseMove = () => {
        if (!pendingMouseEvent) return;

        const now = Date.now();
        if (now - lastSendTime < SEND_INTERVAL) {
          // まだ送信間隔に達していない場合は次のフレームで再試行
          animationFrameId = requestAnimationFrame(processMouseMove);
          return;
        }

        const event = pendingMouseEvent;
        pendingMouseEvent = null;
        lastSendTime = now;
        const rect = canvas.getBoundingClientRect();
        const currentPlayer = gameState.players?.find((p) => p.id === playerId);
        const playerPosition = currentPlayer?.organism?.core?.position;

        if (!playerPosition) return;

        // カメラオフセットを考慮したワールド座標に変換
        const cameraX = playerPosition.x - canvasSize.width / 2;
        const cameraY = playerPosition.y - canvasSize.height / 2;

        const worldX = event.clientX - rect.left + cameraX;
        const worldY = event.clientY - rect.top + cameraY;

        // コアからマウス位置への方向ベクトルを計算
        const dx = worldX - playerPosition.x;
        const dy = worldY - playerPosition.y;
        const distance = Math.sqrt(dx * dx + dy * dy);

        // 最小距離チェック（デッドゾーン） px
        const minDistance = 150;

        if (distance > minDistance) {
          // 正規化された方向ベクトル
          const normalizedX = dx / distance;
          const normalizedY = dy / distance;

          onMouseMove(normalizedX, normalizedY);
        } else {
          // 距離が近い場合は停止
          onMouseMove(0, 0);
        }
      };

      const handleMouseMove = (event: MouseEvent) => {
        pendingMouseEvent = event;

        if (!animationFrameId) {
          animationFrameId = requestAnimationFrame(() => {
            processMouseMove();
            animationFrameId = null;
          });
        }
      };

      const handleMouseLeave = () => {
        onMouseMove(0, 0); // マウスがウィンドウ外に出たら停止
      };

      canvas.addEventListener("mousemove", handleMouseMove);
      canvas.addEventListener("mouseleave", handleMouseLeave);

      return () => {
        canvas.removeEventListener("mousemove", handleMouseMove);
        canvas.removeEventListener("mouseleave", handleMouseLeave);
        if (animationFrameId) {
          cancelAnimationFrame(animationFrameId);
        }
      };
    }, [gameState, playerId, canvasSize, onMouseMove]);

    useEffect(() => {
      const canvas = canvasRef.current;
      if (!canvas) return;

      const ctx = canvas.getContext("2d");
      if (!ctx) return;

      // パフォーマンス監視
      const startTime = performance.now();
      frameCountRef.current++;

      try {
        drawGame(
          ctx,
          gameState,
          playerId,
          gameConfig,
          canvasSize,
          showGrid,
          showCulling
        );

        // 20秒毎に軽量ログ出力
        const now = Date.now();
        if (now - lastLogTimeRef.current > 20000) {
          const drawTime = performance.now() - startTime;
          const playerCount = gameState.players?.length || 0;
          const totalSegments =
            gameState.players?.reduce(
              (sum, p) => sum + (p.organism?.nodes?.length || 0) + 1, // コア + ノード
              0
            ) || 0;
          const foodCount = gameState.food?.length || 0;

          console.log(`🎮 CLIENT PERFORMANCE:
Frame: ${frameCountRef.current}
Players: ${playerCount} (Segments: ${totalSegments})
Food: ${foodCount}
Draw Time: ${drawTime.toFixed(2)}ms`);

          lastLogTimeRef.current = now;
        }
      } catch (error) {
        console.error("🚨 DRAW ERROR:", error);
        console.error("GameState:", gameState);
        console.error("PlayerID:", playerId);
      }
    }, [gameState, playerId, gameConfig, canvasSize, showGrid, showCulling]);

    return (
      <canvas
        ref={canvasRef}
        width={canvasSize.width}
        height={canvasSize.height}
        style={{ display: "block" }}
      />
    );
  }
);

GameCanvas.displayName = "GameCanvas";

// メインの描画関数（カメラ追従付き）
const drawGame = (
  ctx: CanvasRenderingContext2D,
  gameState: GameState,
  playerId: string,
  gameConfig: GameConfig,
  canvasSize: { width: number; height: number },
  showGrid: boolean,
  showCulling: boolean
) => {
  // プレイヤーの位置を取得
  const currentPlayer = gameState.players?.find((p) => p.id === playerId);
  const playerPosition = currentPlayer?.organism?.core?.position;

  // カメラの中心位置を計算
  const cameraX = playerPosition ? playerPosition.x - canvasSize.width / 2 : 0;
  const cameraY = playerPosition ? playerPosition.y - canvasSize.height / 2 : 0;

  // キャンバスをクリア
  ctx.fillStyle = "#0a0a0a";
  ctx.fillRect(0, 0, canvasSize.width, canvasSize.height);

  // カメラ変換を適用
  ctx.save();
  ctx.translate(-cameraX, -cameraY);

  // フィールドの境界を描画
  drawFieldBoundary(ctx, gameConfig);

  // SpatialGridの線を描画
  if (showGrid && gameConfig.gridLines) {
    drawSpatialGrid(ctx, gameConfig.gridLines);
  }

  // サーバーカリング範囲を描画
  if (showCulling) {
    drawServerCullingBounds(ctx, playerPosition, gameConfig);
  }

  // 食べ物を描画（カリング付き）
  drawFood(
    ctx,
    gameState.food,
    gameConfig.foodRadius,
    cameraX,
    cameraY,
    canvasSize
  );

  // プレイヤーを描画（カリング付き）
  if (gameState.players && gameState.players.length > 0) {
    // カリング用の画面境界計算（余裕を持たせる）
    const cullingMargin = 300; // 画面外300pxまで描画
    const minX = cameraX - cullingMargin;
    const maxX = cameraX + canvasSize.width + cullingMargin;
    const minY = cameraY - cullingMargin;
    const maxY = cameraY + canvasSize.height + cullingMargin;

    gameState.players.forEach((player) => {
      // プレイヤーが画面範囲内にいるかチェック
      if (player.organism?.core?.position) {
        const head = player.organism.core.position;

        // 頭が画面範囲内にあるかチェック
        if (
          head.x >= minX &&
          head.x <= maxX &&
          head.y >= minY &&
          head.y <= maxY
        ) {
          drawOrganism(ctx, player, player.id === playerId);
        }
      }
    });

    // デバッグ用（20秒毎のログに追加情報を含める） - 削除してエラーを回避
  }

  // カメラ変換を元に戻す
  ctx.restore();

  // UI要素を描画（画面固定）
  drawUI(ctx, currentPlayer, canvasSize, showGrid, showCulling);
};

// フィールドの境界を描画
const drawFieldBoundary = (
  ctx: CanvasRenderingContext2D,
  gameConfig: GameConfig
) => {
  ctx.strokeStyle = "#333";
  ctx.lineWidth = 3;
  ctx.setLineDash([10, 5]);
  ctx.strokeRect(0, 0, gameConfig.fieldWidth, gameConfig.fieldHeight);
  ctx.setLineDash([]);
};

// サーバーカリング範囲を描画（デバッグ用）
const drawServerCullingBounds = (
  ctx: CanvasRenderingContext2D,
  playerPosition: Position | undefined,
  gameConfig: GameConfig
) => {
  if (!playerPosition) return;

  // サーバーから受信したカリング範囲を使用
  const serverViewWidth = gameConfig.cullingWidth;
  const serverViewHeight = gameConfig.cullingHeight;

  const minX = playerPosition.x - serverViewWidth / 2;
  const maxX = playerPosition.x + serverViewWidth / 2;
  const minY = playerPosition.y - serverViewHeight / 2;
  const maxY = playerPosition.y + serverViewHeight / 2;

  // カリング境界を赤い点線で描画
  ctx.strokeStyle = "#ff0000";
  ctx.lineWidth = 2;
  ctx.setLineDash([5, 5]);
  ctx.strokeRect(minX, minY, maxX - minX, maxY - minY);
  ctx.setLineDash([]);
};

// 食べ物の描画（カリング付き）
const drawFood = (
  ctx: CanvasRenderingContext2D,
  food: Position[],
  radius: number,
  cameraX: number,
  cameraY: number,
  canvasSize: { width: number; height: number }
) => {
  if (!food || food.length === 0) return;

  // カリング境界
  const margin = 100;
  const minX = cameraX - margin;
  const maxX = cameraX + canvasSize.width + margin;
  const minY = cameraY - margin;
  const maxY = cameraY + canvasSize.height + margin;

  ctx.fillStyle = "#ff6b6b";
  ctx.shadowBlur = 10;
  ctx.shadowColor = "#ff6b6b";

  food.forEach((item) => {
    // 画面範囲内の食べ物のみ描画
    if (item.x >= minX && item.x <= maxX && item.y >= minY && item.y <= maxY) {
      ctx.beginPath();
      ctx.arc(item.x, item.y, radius, 0, 2 * Math.PI);
      ctx.fill();
    }
  });

  ctx.shadowBlur = 0;
};

// 球体構造の描画
const drawOrganism = (
  ctx: CanvasRenderingContext2D,
  player: Player,
  isCurrentPlayer: boolean
) => {
  const organism = player.organism;

  // 死んでいる球体構造は半透明に
  ctx.globalAlpha = organism.alive ? 1 : 0.3;
  ctx.fillStyle = organism.color;

  // 接続（線）を先に描画 - 原子構造（コアから各ノードへの放射状＋ノード間環状）
  ctx.strokeStyle = organism.color;
  ctx.lineWidth = 2;
  ctx.globalAlpha = 0.6; // 線を少し透明に

  // コアから各ノードへの線を描画
  organism.nodes.forEach((node) => {
    ctx.beginPath();
    ctx.moveTo(organism.core.position.x, organism.core.position.y);
    ctx.lineTo(node.position.x, node.position.y);
    ctx.stroke();
  });

  // ノード間の環状接続を描画
  ctx.lineWidth = 1.5;
  ctx.globalAlpha = 0.4; // より薄く
  for (let i = 0; i < organism.nodes.length; i++) {
    const nextIndex = (i + 1) % organism.nodes.length;
    const currentNode = organism.nodes[i];
    const nextNode = organism.nodes[nextIndex];

    ctx.beginPath();
    ctx.moveTo(currentNode.position.x, currentNode.position.y);
    ctx.lineTo(nextNode.position.x, nextNode.position.y);
    ctx.stroke();
  }

  ctx.globalAlpha = organism.alive ? 1 : 0.3; // 透明度を戻す

  // コア（中心球）を描画
  drawOrganismHead(
    ctx,
    organism.core.position,
    organism.core.radius,
    organism.color,
    isCurrentPlayer
  );

  // ノード（周辺球）を描画
  organism.nodes.forEach((node) => {
    ctx.beginPath();
    ctx.arc(node.position.x, node.position.y, node.radius, 0, 2 * Math.PI);
    ctx.fill();
  });

  // プレイヤー名を描画
  drawPlayerName(ctx, player.name, organism.core.position, isCurrentPlayer);

  ctx.globalAlpha = 1;
};

// 球体構造の頭部（コア）の描画
const drawOrganismHead = (
  ctx: CanvasRenderingContext2D,
  position: Position,
  radius: number,
  color: string,
  isCurrentPlayer: boolean
) => {
  // 頭部の円
  ctx.beginPath();
  ctx.arc(position.x, position.y, radius, 0, 2 * Math.PI);
  ctx.fill();

  // 自分の球体構造にはアウトラインを追加
  if (isCurrentPlayer) {
    ctx.strokeStyle = "#ffd700";
    ctx.lineWidth = 3;
    ctx.shadowBlur = 15;
    ctx.shadowColor = "#ffd700";
    ctx.beginPath();
    ctx.arc(position.x, position.y, radius + 2, 0, 2 * Math.PI);
    ctx.stroke();
    ctx.shadowBlur = 0;
  }

  // 目を描画
  ctx.fillStyle = "#000";
  const eyeRadius = radius * 0.25;
  const eyeOffset = radius * 0.4;

  ctx.beginPath();
  ctx.arc(
    position.x - eyeOffset,
    position.y - eyeOffset,
    eyeRadius,
    0,
    2 * Math.PI
  );
  ctx.fill();

  ctx.beginPath();
  ctx.arc(
    position.x + eyeOffset,
    position.y - eyeOffset,
    eyeRadius,
    0,
    2 * Math.PI
  );
  ctx.fill();

  // 色を元に戻す
  ctx.fillStyle = color;
};

// プレイヤー名の描画
const drawPlayerName = (
  ctx: CanvasRenderingContext2D,
  name: string,
  headPosition: Position,
  isCurrentPlayer: boolean
) => {
  ctx.globalAlpha = 1;
  ctx.fillStyle = isCurrentPlayer ? "#ffd700" : "#fff";
  ctx.font = isCurrentPlayer ? "bold 16px Arial" : "14px Arial";
  ctx.textAlign = "center";

  // 文字の背景（可読性向上）
  const textWidth = ctx.measureText(name).width;
  ctx.fillStyle = "rgba(0, 0, 0, 0.8)";
  ctx.fillRect(
    headPosition.x - textWidth / 2 - 4,
    headPosition.y - 28,
    textWidth + 8,
    18
  );

  // 文字を描画
  ctx.fillStyle = isCurrentPlayer ? "#ffd700" : "#fff";
  ctx.fillText(name, headPosition.x, headPosition.y - 15);
};

// UI要素の描画（画面固定）
const drawUI = (
  ctx: CanvasRenderingContext2D,
  currentPlayer: Player | undefined,
  canvasSize: { width: number; height: number },
  showGrid: boolean,
  showCulling: boolean
) => {
  if (!currentPlayer) return;

  // スコア表示
  ctx.fillStyle = "rgba(0, 0, 0, 0.8)";
  ctx.fillRect(10, 10, 200, 80);

  ctx.fillStyle = "#fff";
  ctx.font = "bold 18px Arial";
  ctx.textAlign = "left";
  ctx.fillText(`Score: ${currentPlayer.score}`, 20, 35);
  ctx.fillText(`Length: ${currentPlayer.organism.nodes.length + 1}`, 20, 55);

  // 死んでいる場合はDEAD表示
  if (!currentPlayer.organism.alive) {
    ctx.fillStyle = "#ff4444";
    ctx.font = "bold 20px Arial";
    ctx.fillText("DEAD", 20, 80);
  }

  // 表示設定とヘルプ
  const helpLines = [];
  if (showGrid) helpLines.push("Grid: ON (G to toggle)");
  if (showCulling) helpLines.push("Culling: ON (C to toggle)");

  if (helpLines.length > 0 || (!showGrid && !showCulling)) {
    const boxHeight = Math.max(60, 20 + helpLines.length * 20);
    ctx.fillStyle = "rgba(0, 0, 0, 0.8)";
    ctx.fillRect(10, 100, 280, boxHeight);

    ctx.fillStyle = "#fff";
    ctx.font = "14px Arial";
    ctx.textAlign = "left";

    let yPos = 120;
    if (helpLines.length === 0) {
      ctx.fillText("Press G: Grid lines, C: Culling bounds", 20, yPos);
    } else {
      helpLines.forEach((line) => {
        ctx.fillText(line, 20, yPos);
        yPos += 20;
      });
      ctx.fillText("Press G/C to toggle", 20, yPos);
    }
  }

  // ミニマップ（右下）
  drawMinimap(ctx, currentPlayer, canvasSize);
};

// SpatialGridの線を描画
const drawSpatialGrid = (
  ctx: CanvasRenderingContext2D,
  gridLines: GridLine[]
) => {
  ctx.strokeStyle = "rgba(0, 200, 255, 0.8)"; // 青色、より濃く
  ctx.lineWidth = 1.5; // 線を太く
  ctx.setLineDash([3, 3]); // 点線を少し長く

  gridLines.forEach((line) => {
    ctx.beginPath();
    ctx.moveTo(line.startX, line.startY);
    ctx.lineTo(line.endX, line.endY);
    ctx.stroke();
  });

  ctx.setLineDash([]); // 点線をリセット
};

// ミニマップの描画
const drawMinimap = (
  ctx: CanvasRenderingContext2D,
  currentPlayer: Player,
  canvasSize: { width: number; height: number }
) => {
  const mapSize = 120;
  const mapX = canvasSize.width - mapSize - 10;
  const mapY = canvasSize.height - mapSize - 10;

  // ミニマップ背景
  ctx.fillStyle = "rgba(0, 0, 0, 0.8)";
  ctx.fillRect(mapX, mapY, mapSize, mapSize);

  ctx.strokeStyle = "#333";
  ctx.lineWidth = 2;
  ctx.strokeRect(mapX, mapY, mapSize, mapSize);

  // プレイヤーの位置を表示
  if (currentPlayer.organism.core) {
    const head = currentPlayer.organism.core.position;
    const playerX = mapX + (head.x / 5000) * mapSize;
    const playerY = mapY + (head.y / 3000) * mapSize;

    ctx.fillStyle = "#ffd700";
    ctx.beginPath();
    ctx.arc(playerX, playerY, 3, 0, 2 * Math.PI);
    ctx.fill();
  }
};

export default GameCanvas;
