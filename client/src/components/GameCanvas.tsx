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
    width: GAME_CONFIG.GRID_SIZE * GAME_CONFIG.CELL_SIZE,
    height: GAME_CONFIG.GRID_SIZE * GAME_CONFIG.CELL_SIZE
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
        const maxSize = Math.min(rect.width - 20, window.innerHeight * 0.5);
        const cellSize = Math.floor(maxSize / GAME_CONFIG.GRID_SIZE);
        const size = cellSize * GAME_CONFIG.GRID_SIZE;
        
        setCanvasSize({ width: size, height: size });
      } else {
        // デスクトップでは固定サイズ
        setCanvasSize({
          width: GAME_CONFIG.GRID_SIZE * GAME_CONFIG.CELL_SIZE,
          height: GAME_CONFIG.GRID_SIZE * GAME_CONFIG.CELL_SIZE
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
    const scale = canvasSize.width / (GAME_CONFIG.GRID_SIZE * GAME_CONFIG.CELL_SIZE);
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
  const { GRID_SIZE, CELL_SIZE } = GAME_CONFIG;
  
  // キャンバスをクリア
  ctx.fillStyle = '#1a1a1a';
  ctx.fillRect(0, 0, GRID_SIZE * CELL_SIZE, GRID_SIZE * CELL_SIZE);

  // グリッドを描画
  drawGrid(ctx, GRID_SIZE, CELL_SIZE);
  
  // 食べ物を描画
  drawFood(ctx, gameState.food, CELL_SIZE);
  
  // 蛇を描画
  drawSnakes(ctx, gameState.players, playerId, CELL_SIZE);
};

// グリッド描画
const drawGrid = (ctx: CanvasRenderingContext2D, gridSize: number, cellSize: number) => {
  ctx.strokeStyle = '#333';
  ctx.lineWidth = 0.5;
  
  for (let i = 0; i <= gridSize; i++) {
    // 縦線
    ctx.beginPath();
    ctx.moveTo(i * cellSize, 0);
    ctx.lineTo(i * cellSize, gridSize * cellSize);
    ctx.stroke();
    
    // 横線
    ctx.beginPath();
    ctx.moveTo(0, i * cellSize);
    ctx.lineTo(gridSize * cellSize, i * cellSize);
    ctx.stroke();
  }
};

// 食べ物描画
const drawFood = (ctx: CanvasRenderingContext2D, food: Array<{x: number, y: number}>, cellSize: number) => {
  ctx.fillStyle = '#ff6b6b';
  
  food.forEach(foodItem => {
    ctx.beginPath();
    ctx.arc(
      foodItem.x * cellSize + cellSize / 2,
      foodItem.y * cellSize + cellSize / 2,
      cellSize / 3,
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
  cellSize: number
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
        drawSnakeHead(ctx, segment, cellSize, isCurrentPlayer);
      } else {
        // 体部
        drawSnakeBody(ctx, segment, cellSize);
      }
    });

    // 透明度をリセット
    ctx.globalAlpha = 1;

    // プレイヤー名を描画
    if (snake.body.length > 0) {
      drawPlayerName(ctx, player.name, snake.body[0], cellSize, isCurrentPlayer);
    }
  });
};

// 蛇の頭部描画
const drawSnakeHead = (
  ctx: CanvasRenderingContext2D, 
  segment: {x: number, y: number}, 
  cellSize: number,
  isCurrentPlayer: boolean
) => {
  const x = segment.x * cellSize;
  const y = segment.y * cellSize;
  
  // 頭部の背景
  ctx.fillRect(x + 1, y + 1, cellSize - 2, cellSize - 2);
  
  // 目を描画
  ctx.fillStyle = '#000';
  const eyeSize = 3;
  const eyeOffset = 3;
  
  ctx.fillRect(x + eyeOffset, y + eyeOffset, eyeSize, eyeSize);
  ctx.fillRect(x + cellSize - eyeOffset - eyeSize, y + eyeOffset, eyeSize, eyeSize);
  
  // 自分の蛇には王冠マークを追加
  if (isCurrentPlayer) {
    ctx.fillStyle = '#ffd700';
    ctx.fillRect(x + cellSize/2 - 2, y - 4, 4, 3);
  }
};

// 蛇の体部描画
const drawSnakeBody = (
  ctx: CanvasRenderingContext2D, 
  segment: {x: number, y: number}, 
  cellSize: number
) => {
  ctx.fillRect(
    segment.x * cellSize + 2,
    segment.y * cellSize + 2,
    cellSize - 4,
    cellSize - 4
  );
};

// プレイヤー名描画
const drawPlayerName = (
  ctx: CanvasRenderingContext2D, 
  name: string, 
  headPosition: {x: number, y: number}, 
  cellSize: number,
  isCurrentPlayer: boolean
) => {
  ctx.fillStyle = isCurrentPlayer ? '#ffd700' : '#fff';
  ctx.font = isCurrentPlayer ? 'bold 12px Arial' : '12px Arial';
  ctx.textAlign = 'center';
  
  const x = headPosition.x * cellSize + cellSize / 2;
  const y = headPosition.y * cellSize - 5;
  
  // 文字の背景（可読性向上）
  const textWidth = ctx.measureText(name).width;
  ctx.fillStyle = 'rgba(0, 0, 0, 0.7)';
  ctx.fillRect(x - textWidth/2 - 2, y - 12, textWidth + 4, 14);
  
  // 文字を描画
  ctx.fillStyle = isCurrentPlayer ? '#ffd700' : '#fff';
  ctx.fillText(name, x, y);
};

export default GameCanvas;