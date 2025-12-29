import React, { useRef, useEffect, useState, memo } from "react";
import type { GameState, GameConfig, GridLine } from "../types";
import { getPlayer, getDroppedSatellite, getProjectile } from "../types";
import type {
  ConvertedPlayer,
  ConvertedDroppedSatellite,
  ConvertedProjectile,
} from "../types";
import { GAME_CONSTANTS, COLORS } from "../constants/game";
import {
  calculateCameraOffset,
  convertMouseToGameCoords,
} from "../utils/viewport";

interface GameCanvasProps {
  gameState: GameState;
  playerId: string;
  gameConfig: GameConfig;
  onMouseMove: (x: number, y: number) => void;
  onMouseClick?: (x: number, y: number) => void;
}

interface Viewport {
  width: number;
  height: number;
  scale: number;
  gameWidth: number;
  gameHeight: number;
  offsetX: number;
  offsetY: number;
}

const GameCanvas: React.FC<GameCanvasProps> = memo(
  ({ gameState, playerId, gameConfig, onMouseMove, onMouseClick }) => {
    const canvasRef = useRef<HTMLCanvasElement>(null);
    const [showGrid, setShowGrid] = useState(true);
    const [showCulling, setShowCulling] = useState(true);
    // 基準解像度（デザイン時の解像度）
    const { BASE_WIDTH, BASE_HEIGHT } = GAME_CONSTANTS;

    // キャンバスサイズとスケール計算
    const [viewport, setViewport] = useState(() => {
      const width = window.innerWidth;
      const height = window.innerHeight;

      // アスペクト比を維持しつつ、画面に収まるスケールを計算
      const scaleX = width / BASE_WIDTH;
      const scaleY = height / BASE_HEIGHT;
      const scale = Math.min(scaleX, scaleY);

      return {
        width,
        height,
        scale,
        // 実際のゲーム描画領域
        gameWidth: BASE_WIDTH * scale,
        gameHeight: BASE_HEIGHT * scale,
        offsetX: (width - BASE_WIDTH * scale) / 2,
        offsetY: (height - BASE_HEIGHT * scale) / 2,
      };
    });

    // ウィンドウリサイズ時のハンドラー
    useEffect(() => {
      const handleResize = () => {
        const width = window.innerWidth;
        const height = window.innerHeight;

        const scaleX = width / BASE_WIDTH;
        const scaleY = height / BASE_HEIGHT;
        const scale = Math.min(scaleX, scaleY);

        setViewport({
          width,
          height,
          scale,
          gameWidth: BASE_WIDTH * scale,
          gameHeight: BASE_HEIGHT * scale,
          offsetX: (width - BASE_WIDTH * scale) / 2,
          offsetY: (height - BASE_HEIGHT * scale) / 2,
        });
      };

      window.addEventListener("resize", handleResize);
      return () => window.removeEventListener("resize", handleResize);
    }, []);

    // 固定キャンバスサイズ（基準解像度を使用）
    const canvasSize = { width: BASE_WIDTH, height: BASE_HEIGHT };

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
      let lastWasInDeadZone = false;
      const SEND_INTERVAL = GAME_CONSTANTS.SEND_INTERVAL;

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
        const currentPlayer = gameState.pls?.find((p) => p[0] === playerId);
        const playerPosition = currentPlayer
          ? getPlayer(currentPlayer)?.cel?.c?.p
          : undefined;

        if (!playerPosition) return;

        // カメラズーム設定を取得
        const gameZoomScale = gameConfig.cameraZoomScale || 1.0;

        // カメラオフセットを考慮したワールド座標に変換
        const { x: cameraX, y: cameraY } = calculateCameraOffset(
          playerPosition,
          gameZoomScale
        );

        // マウス座標をビューポート座標系からゲーム座標系に変換（スケールファクター考慮）
        const gameCoords = convertMouseToGameCoords(event, rect);
        const { x: gameX, y: gameY } = gameCoords;

        const worldX = gameX / gameZoomScale + cameraX;
        const worldY = gameY / gameZoomScale + cameraY;

        // コアからマウス位置への方向ベクトルを計算
        const dx = worldX - playerPosition.x;
        const dy = worldY - playerPosition.y;
        const distance = Math.sqrt(dx * dx + dy * dy);

        // 距離に応じた速度制御
        const { MIN_DISTANCE: minDistance, MAX_DISTANCE: maxDistance } =
          GAME_CONSTANTS;

        if (distance > minDistance) {
          // 正規化された方向ベクトル
          const normalizedX = dx / distance;
          const normalizedY = dy / distance;

          // 距離に応じた速度倍率
          const speedRatio = Math.min(
            0.1 +
              0.9 * ((distance - minDistance) / (maxDistance - minDistance)),
            1.0
          );

          onMouseMove(normalizedX * speedRatio, normalizedY * speedRatio);
          lastWasInDeadZone = false;
        } else {
          // デッドゾーンに入った時、一度だけ停止信号を送信
          if (!lastWasInDeadZone) {
            onMouseMove(0, 0);
            lastWasInDeadZone = true;
          }
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

      let hasLeftWindow = false;

      const handleMouseLeave = () => {
        if (!hasLeftWindow) {
          hasLeftWindow = true;
          onMouseMove(0, 0); // マウスがウィンドウ外に出たら停止
        }
      };

      const handleMouseEnter = () => {
        hasLeftWindow = false;
      };

      canvas.addEventListener("mousemove", handleMouseMove);
      canvas.addEventListener("mouseleave", handleMouseLeave);
      canvas.addEventListener("mouseenter", handleMouseEnter);

      return () => {
        canvas.removeEventListener("mousemove", handleMouseMove);
        canvas.removeEventListener("mouseleave", handleMouseLeave);
        canvas.removeEventListener("mouseenter", handleMouseEnter);
        if (animationFrameId) {
          cancelAnimationFrame(animationFrameId);
        }
      };
    }, [
      gameState,
      playerId,
      canvasSize,
      onMouseMove,
      gameConfig.cameraZoomScale,
      viewport,
    ]);

    // マウスクリックの処理
    useEffect(() => {
      const canvas = canvasRef.current;
      if (!canvas || !onMouseClick) return;

      const handleClick = (event: MouseEvent) => {
        const rect = canvas.getBoundingClientRect();
        const currentPlayer = gameState.pls?.find((p) => p[0] === playerId);
        const playerPosition = currentPlayer
          ? getPlayer(currentPlayer)?.cel?.c?.p
          : undefined;

        if (!playerPosition) return;

        // カメラズーム設定を取得
        const gameZoomScale = gameConfig.cameraZoomScale || 1.0;

        // カメラオフセットを考慮したワールド座標に変換
        const { x: cameraX, y: cameraY } = calculateCameraOffset(
          playerPosition,
          gameZoomScale
        );

        // マウス座標をビューポート座標系からゲーム座標系に変換（スケールファクター考慮）
        const gameCoords = convertMouseToGameCoords(event, rect);
        const { x: gameX, y: gameY } = gameCoords;

        const worldX = gameX / gameZoomScale + cameraX;
        const worldY = gameY / gameZoomScale + cameraY;

        onMouseClick(worldX, worldY);
      };

      canvas.addEventListener("click", handleClick);
      return () => {
        canvas.removeEventListener("click", handleClick);
      };
    }, [
      gameState,
      playerId,
      canvasSize,
      onMouseClick,
      gameConfig.cameraZoomScale,
      viewport,
    ]);

    useEffect(() => {
      const canvas = canvasRef.current;
      if (!canvas) return;

      const ctx = canvas.getContext("2d");
      if (!ctx) return;

      try {
        if (!gameState || !gameState.pls) return;

        drawGame(
          ctx,
          gameState,
          playerId,
          gameConfig,
          canvasSize,
          viewport,
          showGrid,
          showCulling
        );
      } catch (error) {
        console.error("🚨 DRAW ERROR:", error);
        console.error("GameState:", gameState);
        console.error("PlayerID:", playerId);
      }
    }, [
      gameState,
      playerId,
      gameConfig,
      canvasSize,
      showGrid,
      showCulling,
      viewport,
    ]);

    return (
      <canvas
        ref={canvasRef}
        width={1920}
        height={1080}
        style={{
          display: "block",
          width: "100%",
          height: "100%",
        }}
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
  viewport: Viewport,
  showGrid: boolean,
  showCulling: boolean
) => {
  // プレイヤーの位置を取得
  const currentPlayer = gameState.pls?.find((p) => p[0] === playerId);
  const playerData = currentPlayer ? getPlayer(currentPlayer) : null;
  const playerPosition = playerData?.cel?.c?.p;

  // カメラズーム設定を取得
  const zoomScale = gameConfig.cameraZoomScale || 1.0;

  // カメラの中心位置を計算
  const { x: cameraX, y: cameraY } = playerPosition
    ? calculateCameraOffset(playerPosition, gameConfig.cameraZoomScale || 1.0)
    : { x: 0, y: 0 };

  // キャンバスをクリア
  ctx.fillStyle = COLORS.BACKGROUND;
  ctx.fillRect(0, 0, canvasSize.width, canvasSize.height);

  // カメラ変換を適用
  ctx.save();

  // ゲームのズームを適用
  ctx.scale(
    gameConfig.cameraZoomScale || 1.0,
    gameConfig.cameraZoomScale || 1.0
  );

  // カメラの位置を適用
  ctx.translate(-cameraX, -cameraY);

  // フィールドの境界を描画
  drawFieldBoundary(ctx, gameConfig);

  // SpatialGridの線を描画
  if (showGrid && gameConfig.gridLines) {
    drawSpatialGrid(ctx, gameConfig.gridLines);
  }

  // サーバーカリング範囲を描画
  if (showCulling) {
    // 実際のサーバー側カリング範囲を計算（ズーム調整済み）
    const actualCullingConfig = {
      ...gameConfig,
      cullingWidth: gameConfig.cullingWidth / zoomScale,
      cullingHeight: gameConfig.cullingHeight / zoomScale,
    };
    drawServerCullingBounds(ctx, playerPosition, actualCullingConfig);
  }

  // 落ちた衛星を描画（カリング付き）
  drawDroppedSatellites(
    ctx,
    gameState.ds?.map(getDroppedSatellite) || [],
    cameraX,
    cameraY,
    canvasSize,
    zoomScale
  );

  // 射出物を描画（カリング付き）
  drawProjectiles(
    ctx,
    gameState.proj?.map(getProjectile) || [],
    gameConfig.sphereRadius,
    cameraX,
    cameraY,
    canvasSize,
    zoomScale
  );

  // プレイヤーを描画（カリング付き）
  if (gameState.pls && gameState.pls.length > 0) {
    // カリング用の画面境界計算（余裕を持たせる、ズーム考慮）
    const cullingMargin = GAME_CONSTANTS.CULLING_MARGIN;
    const minX = cameraX - cullingMargin / zoomScale;
    const maxX = cameraX + (canvasSize.width + cullingMargin) / zoomScale;
    const minY = cameraY - cullingMargin / zoomScale;
    const maxY = cameraY + (canvasSize.height + cullingMargin) / zoomScale;

    gameState.pls.forEach((player) => {
      const playerData = getPlayer(player);

      // 死んでいるプレイヤーは表示しない
      if (!playerData.cel?.a) {
        return;
      }

      // プレイヤーが画面範囲内にいるかチェック
      if (playerData.cel?.c?.p) {
        const head = playerData.cel.c.p;

        // 頭が画面範囲内にあるかチェック
        if (
          head.x >= minX &&
          head.x <= maxX &&
          head.y >= minY &&
          head.y <= maxY
        ) {
          drawCelestialSystem(ctx, playerData, playerData.id === playerId);
        }
      }
    });
  }

  // カメラ変換を元に戻す
  ctx.restore();

  // UI要素を描画（画面固定）
  const currentPlayerData = playerData || null;
  drawUI(ctx, currentPlayerData, showGrid, showCulling);
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
  playerPosition: { x: number; y: number } | undefined,
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
  ctx.strokeStyle = COLORS.CULLING;
  ctx.lineWidth = 2;
  ctx.setLineDash([5, 5]);
  ctx.strokeRect(minX, minY, maxX - minX, maxY - minY);
  ctx.setLineDash([]);
};

// 落ちた衛星の描画（カリング付き）
const drawDroppedSatellites = (
  ctx: CanvasRenderingContext2D,
  droppedSatellites: ConvertedDroppedSatellite[] | undefined,
  cameraX: number,
  cameraY: number,
  canvasSize: { width: number; height: number },
  zoomScale: number
) => {
  if (!droppedSatellites || droppedSatellites.length === 0) return;

  // カリング境界（ズーム考慮）
  const margin = 100;
  const minX = cameraX - margin / zoomScale;
  const maxX = cameraX + (canvasSize.width + margin) / zoomScale;
  const minY = cameraY - margin / zoomScale;
  const maxY = cameraY + (canvasSize.height + margin) / zoomScale;

  droppedSatellites.forEach((satellite) => {
    // 画面範囲内の落ちた衛星のみ描画
    if (
      satellite.p.x >= minX &&
      satellite.p.x <= maxX &&
      satellite.p.y >= minY &&
      satellite.p.y <= maxY
    ) {
      // 衛星の色を使用
      ctx.fillStyle = satellite.c;
      ctx.shadowBlur = 8;
      ctx.shadowColor = satellite.c;

      ctx.beginPath();
      ctx.arc(satellite.p.x, satellite.p.y, satellite.r, 0, 2 * Math.PI);
      ctx.fill();

      // 内側に小さな光る点を追加
      ctx.fillStyle = COLORS.WHITE;
      ctx.shadowBlur = 4;
      ctx.shadowColor = COLORS.WHITE;
      ctx.beginPath();
      ctx.arc(satellite.p.x, satellite.p.y, satellite.r * 0.3, 0, 2 * Math.PI);
      ctx.fill();
    }
  });

  ctx.shadowBlur = 0;
};

// 射出物の描画（カリング付き）
const drawProjectiles = (
  ctx: CanvasRenderingContext2D,
  projectiles: ConvertedProjectile[] | undefined,
  radius: number,
  cameraX: number,
  cameraY: number,
  canvasSize: { width: number; height: number },
  zoomScale: number
) => {
  if (!projectiles || projectiles.length === 0) return;

  // カリング境界（ズーム考慮）
  const margin = 100;
  const minX = cameraX - margin / zoomScale;
  const maxX = cameraX + (canvasSize.width + margin) / zoomScale;
  const minY = cameraY - margin / zoomScale;
  const maxY = cameraY + (canvasSize.height + margin) / zoomScale;

  projectiles.forEach((projectile) => {
    const pos = projectile.sph.p;

    // 画面範囲内の射出物のみ描画
    if (pos.x >= minX && pos.x <= maxX && pos.y >= minY && pos.y <= maxY) {
      // 射出物の色を使用
      const projectileColor = projectile.sph.c;
      ctx.fillStyle = projectileColor;
      ctx.shadowBlur = 15;
      ctx.shadowColor = projectileColor;

      ctx.beginPath();
      ctx.arc(pos.x, pos.y, radius, 0, 2 * Math.PI);
      ctx.fill();

      // 軌跡の描画（速度ベクトルの逆方向に短い線）
      if (projectile.sph.v) {
        const vel = projectile.sph.v;
        const speed = Math.sqrt(vel.x * vel.x + vel.y * vel.y);
        if (speed > 0) {
          const trailLength = GAME_CONSTANTS.PROJECTILE_TRAIL_LENGTH;
          const dirX = -vel.x / speed;
          const dirY = -vel.y / speed;

          ctx.strokeStyle = projectileColor;
          ctx.lineWidth = 3;
          ctx.globalAlpha = 0.5;
          ctx.beginPath();
          ctx.moveTo(pos.x, pos.y);
          ctx.lineTo(pos.x + dirX * trailLength, pos.y + dirY * trailLength);
          ctx.stroke();
          ctx.globalAlpha = 1;
        }
      }

      ctx.shadowBlur = 0;
    }
  });
};

// 球体構造の描画
const drawCelestialSystem = (
  ctx: CanvasRenderingContext2D,
  player: ConvertedPlayer,
  isCurrentPlayer: boolean
) => {
  const celestialSystem = player.cel;
  ctx.fillStyle = celestialSystem.col;

  // 軌道を先に描画 - 核から各衛星への放射状線
  ctx.lineWidth = 2;
  ctx.globalAlpha = player.inv ? 0.3 : 0.6; // 無敵時は線も透明に

  // コアから各ノードへの線を描画
  if (celestialSystem.n && celestialSystem.n.length > 0) {
    celestialSystem.n.forEach((node) => {
      // 各衛星の色で線を描画
      ctx.strokeStyle = node.c;
      ctx.beginPath();
      ctx.moveTo(celestialSystem.c.p.x, celestialSystem.c.p.y);
      ctx.lineTo(node.p.x, node.p.y);
      ctx.stroke();
    });
  }

  ctx.globalAlpha = player.inv ? 0.5 : 1.0; // 球体描画用の透明度に設定

  // コア（中心球）を描画
  drawCoreHead(
    ctx,
    celestialSystem.c.p,
    celestialSystem.c.r,
    celestialSystem.c.c, // コア球体の色を使用
    isCurrentPlayer
  );

  // ノード（周辺球）を描画
  if (celestialSystem.n && celestialSystem.n.length > 0) {
    celestialSystem.n.forEach((node) => {
      // 各衛星の色を使用
      ctx.fillStyle = node.c;
      ctx.beginPath();
      ctx.arc(node.p.x, node.p.y, node.r, 0, 2 * Math.PI);
      ctx.fill();
    });
  }

  // プレイヤー名を描画
  drawPlayerName(ctx, player.nm, celestialSystem.c.p, isCurrentPlayer);

  ctx.globalAlpha = 1;
};

// 球体構造の頭部（コア）の描画
const drawCoreHead = (
  ctx: CanvasRenderingContext2D,
  position: { x: number; y: number },
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
    ctx.strokeStyle = COLORS.GOLD;
    ctx.lineWidth = 3;
    ctx.shadowBlur = 15;
    ctx.shadowColor = COLORS.GOLD;
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
  headPosition: { x: number; y: number },
  isCurrentPlayer: boolean
) => {
  ctx.globalAlpha = 1;
  ctx.fillStyle = isCurrentPlayer ? COLORS.GOLD : COLORS.WHITE;
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
  ctx.fillStyle = isCurrentPlayer ? COLORS.GOLD : COLORS.WHITE;
  ctx.fillText(name, headPosition.x, headPosition.y - 15);
};

// UI要素の描画（画面固定）
const drawUI = (
  ctx: CanvasRenderingContext2D,
  currentPlayer: ConvertedPlayer | null,
  showGrid: boolean,
  showCulling: boolean
) => {
  if (!currentPlayer) return;

  // スコア表示
  ctx.fillStyle = "rgba(0, 0, 0, 0.8)";
  ctx.fillRect(10, 10, 200, 80);

  ctx.fillStyle = COLORS.WHITE;
  ctx.font = "bold 18px Arial";
  ctx.textAlign = "left";
  ctx.fillText(`Score: ${currentPlayer.sc}`, 20, 35);
  ctx.fillText(`Length: ${(currentPlayer.cel.n?.length || 0) + 1}`, 20, 55);

  // 死んでいる場合はDEAD表示
  if (!currentPlayer.cel.a) {
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

    ctx.fillStyle = COLORS.WHITE;
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
};

// SpatialGridの線を描画
const drawSpatialGrid = (
  ctx: CanvasRenderingContext2D,
  gridLines: GridLine[]
) => {
  ctx.strokeStyle = COLORS.GRID;
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

export default GameCanvas;
