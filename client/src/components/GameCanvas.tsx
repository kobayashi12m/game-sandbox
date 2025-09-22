import React, { useRef, useEffect, useState } from 'react';
import type { GameState } from '../types';
import { GAME_CONFIG } from '../types';

interface GameCanvasProps {
  gameState: GameState;
  playerId: string;
}

const GameCanvas: React.FC<GameCanvasProps> = ({ gameState, playerId }) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [canvasSize, setCanvasSize] = useState({
    width: GAME_CONFIG.DISPLAY_WIDTH,
    height: GAME_CONFIG.DISPLAY_HEIGHT
  });

  // ウィンドウサイズに応じてCanvasサイズを調整
  useEffect(() => {
    const updateCanvasSize = () => {
      if (!containerRef.current) return;
      
      const container = containerRef.current;
      const rect = container.getBoundingClientRect();
      const isMobile = window.innerWidth <= 768;
      
      if (isMobile) {
        // モバイルでは利用可能な幅に合わせる
        const maxSize = Math.min(rect.width - 20, window.innerHeight * 0.5) as number;
        setCanvasSize({ width: maxSize, height: maxSize });
      } else {
        // デスクトップでは固定サイズ
        setCanvasSize({
          width: GAME_CONFIG.DISPLAY_WIDTH,
          height: GAME_CONFIG.DISPLAY_HEIGHT
        });
      }
    };

    updateCanvasSize();
    window.addEventListener('resize', updateCanvasSize);
    return () => window.removeEventListener('resize', updateCanvasSize);
  }, []);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // スケールを計算
    const scale = canvasSize.width / GAME_CONFIG.FIELD_WIDTH;
    ctx.save();
    ctx.scale(scale, scale);
    
    drawGame(ctx, gameState, playerId);
    
    ctx.restore();
  }, [gameState, playerId, canvasSize]);

  return (
    <div ref={containerRef} className="canvas-container">
      <canvas
        ref={canvasRef}
        width={canvasSize.width}
        height={canvasSize.height}
        className="game-canvas"
      />
    </div>
  );
};

// ゲーム描画の関数を分離
const drawGame = (
  ctx: CanvasRenderingContext2D, 
  gameState: GameState, 
  playerId: string
) => {
  const { FIELD_WIDTH, FIELD_HEIGHT, SNAKE_RADIUS, FOOD_RADIUS } = GAME_CONFIG;
  
  // キャンバスをクリア
  ctx.fillStyle = '#1a1a1a';
  ctx.fillRect(0, 0, FIELD_WIDTH, FIELD_HEIGHT);

  // グリッドを描画（オプション）
  drawGrid(ctx, FIELD_WIDTH, FIELD_HEIGHT);
  
  // 食べ物を描画
  drawFood(ctx, gameState.food, FOOD_RADIUS);
  
  // 蛇を描画
  drawSnakes(ctx, gameState.players, playerId, SNAKE_RADIUS);
};

// グリッド描画（オプション）
const drawGrid = (ctx: CanvasRenderingContext2D, width: number, height: number) => {
  ctx.strokeStyle = '#333';
  ctx.lineWidth = 0.5;
  const gridSpacing = 30; // グリッド間隔
  
  // 縦線
  for (let x = 0; x <= width; x += gridSpacing) {
    ctx.beginPath();
    ctx.moveTo(x, 0);
    ctx.lineTo(x, height);
    ctx.stroke();
  }
  
  // 横線
  for (let y = 0; y <= height; y += gridSpacing) {
    ctx.beginPath();
    ctx.moveTo(0, y);
    ctx.lineTo(width, y);
    ctx.stroke();
  }
};

// 食べ物描画
const drawFood = (ctx: CanvasRenderingContext2D, food: Array<{x: number, y: number}>, radius: number) => {
  ctx.fillStyle = '#ff6b6b';
  
  food.forEach(foodItem => {
    ctx.beginPath();
    ctx.arc(
      foodItem.x,
      foodItem.y,
      radius,
      0,
      2 * Math.PI
    );
    ctx.fill();
  });
};

// 蛇描画
const drawSnakes = (
  ctx: CanvasRenderingContext2D, 
  players: GameState['players'], 
  playerId: string, 
  snakeRadius: number
) => {
  players.forEach(player => {
    const snake = player.snake;
    const isCurrentPlayer = player.id === playerId;
    
    // 死んでいる蛇は半透明に
    if (!snake.alive) {
      ctx.globalAlpha = 0.3;
    }

    // 蛇の体を描画
    snake.body.forEach((segment, index) => {
      ctx.fillStyle = snake.color;
      
      if (index === 0) {
        // 頭部
        drawSnakeHead(ctx, segment, snakeRadius, snake.color, isCurrentPlayer);
      } else {
        // 体部
        ctx.beginPath();
        ctx.arc(segment.x, segment.y, snakeRadius, 0, 2 * Math.PI);
        ctx.fill();
      }
    });

    // 透明度をリセット
    ctx.globalAlpha = 1;

    // プレイヤー名を描画
    if (snake.body.length > 0) {
      drawPlayerName(ctx, player.name, snake.body[0], isCurrentPlayer);
    }
  });
};

// 蛇の頭部描画
const drawSnakeHead = (
  ctx: CanvasRenderingContext2D, 
  segment: {x: number, y: number}, 
  radius: number,
  color: string,
  isCurrentPlayer: boolean
) => {
  // 頭部の円
  ctx.beginPath();
  ctx.arc(segment.x, segment.y, radius * 1.1, 0, 2 * Math.PI);
  ctx.fillStyle = color;
  ctx.fill();
  
  // 目を描画
  ctx.fillStyle = '#000';
  const eyeRadius = radius * 0.25;
  const eyeOffset = radius * 0.5;
  
  ctx.beginPath();
  ctx.arc(segment.x - eyeOffset, segment.y - eyeOffset, eyeRadius, 0, 2 * Math.PI);
  ctx.fill();
  
  ctx.beginPath();
  ctx.arc(segment.x + eyeOffset, segment.y - eyeOffset, eyeRadius, 0, 2 * Math.PI);
  ctx.fill();
  
  // 自分の蛇にはアウトラインを追加
  if (isCurrentPlayer) {
    ctx.strokeStyle = '#ffd700';
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.arc(segment.x, segment.y, radius * 1.2, 0, 2 * Math.PI);
    ctx.stroke();
  }
};


// プレイヤー名描画
const drawPlayerName = (
  ctx: CanvasRenderingContext2D, 
  name: string, 
  headPosition: {x: number, y: number}, 
  isCurrentPlayer: boolean
) => {
  ctx.fillStyle = isCurrentPlayer ? '#ffd700' : '#fff';
  ctx.font = isCurrentPlayer ? 'bold 12px Arial' : '12px Arial';
  ctx.textAlign = 'center';
  
  const x = headPosition.x;
  const y = headPosition.y - 20;
  
  // 文字の背景（可読性向上）
  const textWidth = ctx.measureText(name).width;
  ctx.fillStyle = 'rgba(0, 0, 0, 0.7)';
  ctx.fillRect(x - textWidth/2 - 2, y - 12, textWidth + 4, 14);
  
  // 文字を描画
  ctx.fillStyle = isCurrentPlayer ? '#ffd700' : '#fff';
  ctx.fillText(name, x, y);
};

export default GameCanvas;