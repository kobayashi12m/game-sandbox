import React, { useRef, useEffect, useState, useMemo } from 'react';
import type { GameState, GameConfig, Position, Player } from '../types';

interface GameCanvasProps {
  gameState: GameState;
  playerId: string;
  gameConfig: GameConfig;
}

const GameCanvas: React.FC<GameCanvasProps> = ({ gameState, playerId, gameConfig }) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  // サーバー設定を基に固定サイズを計算（レイアウトシフト防止）
  const canvasSize = useMemo(() => {
    // ゲームフィールドのアスペクト比を維持しつつ、ウィンドウに収まるサイズ
    const fieldRatio = gameConfig.fieldWidth / gameConfig.fieldHeight;
    const windowWidth = window.innerWidth;
    const windowHeight = window.innerHeight - 120; // オーバーレイUI分を除く
    
    let width = windowWidth;
    let height = width / fieldRatio;
    
    // 高さが足りない場合は高さ基準で調整
    if (height > windowHeight) {
      height = windowHeight;
      width = height * fieldRatio;
    }
    
    return { width: Math.floor(width), height: Math.floor(height) };
  }, [gameConfig.fieldWidth, gameConfig.fieldHeight]);
  
  const frameCountRef = useRef(0);
  const lastLogTimeRef = useRef(Date.now());

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // パフォーマンス監視
    const startTime = performance.now();
    frameCountRef.current++;
    
    try {
      drawGame(ctx, gameState, playerId, gameConfig, canvasSize);
      
      // 20秒毎に軽量ログ出力
      const now = Date.now();
      if (now - lastLogTimeRef.current > 20000) {
        const memory = (performance as any).memory;
        const drawTime = performance.now() - startTime;
        const playerCount = gameState.players?.length || 0;
        const totalSegments = gameState.players?.reduce((sum, p) => sum + (p.snake?.body?.length || 0), 0) || 0;
        const foodCount = gameState.food?.length || 0;
        
        console.log(`🎮 CLIENT PERFORMANCE:
Frame: ${frameCountRef.current}
Players: ${playerCount} (Segments: ${totalSegments})
Food: ${foodCount}
Draw Time: ${drawTime.toFixed(2)}ms
Memory Used: ${memory ? (memory.usedJSHeapSize / 1024 / 1024).toFixed(1) : 'N/A'}MB
Memory Limit: ${memory ? (memory.jsHeapSizeLimit / 1024 / 1024).toFixed(1) : 'N/A'}MB`);
        
        lastLogTimeRef.current = now;
      }
    } catch (error) {
      console.error('🚨 DRAW ERROR:', error);
      console.error('GameState:', gameState);
      console.error('PlayerID:', playerId);
    }
  }, [gameState, playerId, gameConfig, canvasSize]);

  return (
    <canvas
      ref={canvasRef}
      width={canvasSize.width}
      height={canvasSize.height}
      style={{ display: 'block' }}
    />
  );
};

// メインの描画関数（カメラ追従付き）
const drawGame = (
  ctx: CanvasRenderingContext2D,
  gameState: GameState,
  playerId: string,
  gameConfig: GameConfig,
  canvasSize: { width: number; height: number }
) => {
  // プレイヤーの位置を取得
  const currentPlayer = gameState.players?.find(p => p.id === playerId);
  const playerPosition = currentPlayer?.snake?.body?.[0];
  
  // カメラの中心位置を計算
  const cameraX = playerPosition ? playerPosition.x - canvasSize.width / 2 : 0;
  const cameraY = playerPosition ? playerPosition.y - canvasSize.height / 2 : 0;

  // キャンバスをクリア
  ctx.fillStyle = '#0a0a0a';
  ctx.fillRect(0, 0, canvasSize.width, canvasSize.height);

  // カメラ変換を適用
  ctx.save();
  ctx.translate(-cameraX, -cameraY);

  // フィールドの境界を描画
  drawFieldBoundary(ctx, gameConfig);

  // 食べ物を描画（カリング付き）
  drawFood(ctx, gameState.food, gameConfig.foodRadius, cameraX, cameraY, canvasSize);

  // プレイヤーを描画（カリング付き）
  if (gameState.players && gameState.players.length > 0) {
    // カリング用の画面境界計算（余裕を持たせる）
    const cullingMargin = 300; // 画面外300pxまで描画
    const minX = cameraX - cullingMargin;
    const maxX = cameraX + canvasSize.width + cullingMargin;
    const minY = cameraY - cullingMargin;
    const maxY = cameraY + canvasSize.height + cullingMargin;
    
    let drawnCount = 0;
    let culledCount = 0;
    
    gameState.players.forEach(player => {
      // プレイヤーが画面範囲内にいるかチェック
      if (player.snake?.body?.[0]) {
        const head = player.snake.body[0];
        
        // 頭が画面範囲内にあるかチェック
        if (head.x >= minX && head.x <= maxX && head.y >= minY && head.y <= maxY) {
          drawSnake(ctx, player, player.id === playerId, gameConfig.snakeRadius);
          drawnCount++;
        } else {
          culledCount++;
        }
      }
    });
    
    // デバッグ用（20秒毎のログに追加情報を含める） - 削除してエラーを回避
  }

  // カメラ変換を元に戻す
  ctx.restore();

  // UI要素を描画（画面固定）
  drawUI(ctx, currentPlayer, canvasSize);
};

// フィールドの境界を描画
const drawFieldBoundary = (
  ctx: CanvasRenderingContext2D,
  gameConfig: GameConfig
) => {
  ctx.strokeStyle = '#333';
  ctx.lineWidth = 3;
  ctx.setLineDash([10, 5]);
  ctx.strokeRect(0, 0, gameConfig.fieldWidth, gameConfig.fieldHeight);
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
  
  ctx.fillStyle = '#ff6b6b';
  ctx.shadowBlur = 10;
  ctx.shadowColor = '#ff6b6b';
  
  let drawnFoodCount = 0;
  
  food.forEach(item => {
    // 画面範囲内の食べ物のみ描画
    if (item.x >= minX && item.x <= maxX && item.y >= minY && item.y <= maxY) {
      ctx.beginPath();
      ctx.arc(item.x, item.y, radius, 0, 2 * Math.PI);
      ctx.fill();
      drawnFoodCount++;
    }
  });
  
  ctx.shadowBlur = 0;
};

// 蛇の描画
const drawSnake = (
  ctx: CanvasRenderingContext2D,
  player: Player,
  isCurrentPlayer: boolean,
  radius: number
) => {
  const snake = player.snake;
  
  // 死んでいる蛇は半透明に
  ctx.globalAlpha = snake.alive ? 1 : 0.3;
  ctx.fillStyle = snake.color;

  // 蛇の体を描画
  snake.body.forEach((segment, index) => {
    if (index === 0) {
      // 頭部を描画
      drawSnakeHead(ctx, segment, radius, snake.color, isCurrentPlayer);
    } else {
      // 体を描画（半径が負にならないよう制限）
      const segmentRadius = Math.max(radius * (1 - index * 0.02), radius * 0.1);
      ctx.beginPath();
      ctx.arc(segment.x, segment.y, segmentRadius, 0, 2 * Math.PI);
      ctx.fill();
    }
  });

  // プレイヤー名を描画
  if (snake.body.length > 0) {
    drawPlayerName(ctx, player.name, snake.body[0], isCurrentPlayer);
  }

  ctx.globalAlpha = 1;
};

// 蛇の頭部の描画
const drawSnakeHead = (
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

  // 自分の蛇にはアウトラインを追加
  if (isCurrentPlayer) {
    ctx.strokeStyle = '#ffd700';
    ctx.lineWidth = 3;
    ctx.shadowBlur = 15;
    ctx.shadowColor = '#ffd700';
    ctx.beginPath();
    ctx.arc(position.x, position.y, radius + 2, 0, 2 * Math.PI);
    ctx.stroke();
    ctx.shadowBlur = 0;
  }

  // 目を描画
  ctx.fillStyle = '#000';
  const eyeRadius = radius * 0.25;
  const eyeOffset = radius * 0.4;
  
  ctx.beginPath();
  ctx.arc(position.x - eyeOffset, position.y - eyeOffset, eyeRadius, 0, 2 * Math.PI);
  ctx.fill();
  
  ctx.beginPath();
  ctx.arc(position.x + eyeOffset, position.y - eyeOffset, eyeRadius, 0, 2 * Math.PI);
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
  ctx.fillStyle = isCurrentPlayer ? '#ffd700' : '#fff';
  ctx.font = isCurrentPlayer ? 'bold 16px Arial' : '14px Arial';
  ctx.textAlign = 'center';
  
  // 文字の背景（可読性向上）
  const textWidth = ctx.measureText(name).width;
  ctx.fillStyle = 'rgba(0, 0, 0, 0.8)';
  ctx.fillRect(headPosition.x - textWidth/2 - 4, headPosition.y - 28, textWidth + 8, 18);
  
  // 文字を描画
  ctx.fillStyle = isCurrentPlayer ? '#ffd700' : '#fff';
  ctx.fillText(name, headPosition.x, headPosition.y - 15);
};

// UI要素の描画（画面固定）
const drawUI = (
  ctx: CanvasRenderingContext2D,
  currentPlayer: Player | undefined,
  canvasSize: { width: number; height: number }
) => {
  if (!currentPlayer) return;

  // スコア表示
  ctx.fillStyle = 'rgba(0, 0, 0, 0.8)';
  ctx.fillRect(10, 10, 200, 60);
  
  ctx.fillStyle = '#fff';
  ctx.font = 'bold 18px Arial';
  ctx.textAlign = 'left';
  ctx.fillText(`Score: ${currentPlayer.score}`, 20, 35);
  ctx.fillText(`Length: ${currentPlayer.snake.body.length}`, 20, 55);

  // ミニマップ（右下）
  drawMinimap(ctx, currentPlayer, canvasSize);
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
  ctx.fillStyle = 'rgba(0, 0, 0, 0.8)';
  ctx.fillRect(mapX, mapY, mapSize, mapSize);
  
  ctx.strokeStyle = '#333';
  ctx.lineWidth = 2;
  ctx.strokeRect(mapX, mapY, mapSize, mapSize);

  // プレイヤーの位置を表示
  if (currentPlayer.snake.body.length > 0) {
    const head = currentPlayer.snake.body[0];
    const playerX = mapX + (head.x / 5000) * mapSize;
    const playerY = mapY + (head.y / 3000) * mapSize;
    
    ctx.fillStyle = '#ffd700';
    ctx.beginPath();
    ctx.arc(playerX, playerY, 3, 0, 2 * Math.PI);
    ctx.fill();
  }
};

export default GameCanvas;