import React, { useRef, useEffect } from 'react';
import type { GameState, GameConfig, Position, Player } from '../types';

interface GameCanvasProps {
  gameState: GameState;
  playerId: string;
  gameConfig: GameConfig;
}

const GameCanvas: React.FC<GameCanvasProps> = ({ gameState, playerId, gameConfig }) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    drawGame(ctx, gameState, playerId, gameConfig);
  }, [gameState, playerId, gameConfig]);

  return (
    <canvas
      ref={canvasRef}
      width={gameConfig.fieldWidth}
      height={gameConfig.fieldHeight}
      style={{ border: '2px solid #333' }}
    />
  );
};

// メインの描画関数
const drawGame = (
  ctx: CanvasRenderingContext2D,
  gameState: GameState,
  playerId: string,
  gameConfig: GameConfig
) => {
  // キャンバスをクリア
  ctx.fillStyle = '#1a1a1a';
  ctx.fillRect(0, 0, gameConfig.fieldWidth, gameConfig.fieldHeight);

  // 食べ物を描画
  drawFood(ctx, gameState.food, gameConfig.foodRadius);

  // プレイヤーを描画
  if (gameState.players && gameState.players.length > 0) {
    gameState.players.forEach(player => {
      drawSnake(ctx, player, player.id === playerId, gameConfig.snakeRadius);
    });
  }
};

// 食べ物の描画
const drawFood = (
  ctx: CanvasRenderingContext2D,
  food: Position[],
  radius: number
) => {
  if (!food || food.length === 0) return;
  
  ctx.fillStyle = '#ff6b6b';
  food.forEach(item => {
    ctx.beginPath();
    ctx.arc(item.x, item.y, radius, 0, 2 * Math.PI);
    ctx.fill();
  });
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
      drawSnakeHead(ctx, segment, radius, snake.color);
    } else {
      // 体を描画
      ctx.beginPath();
      ctx.arc(segment.x, segment.y, radius, 0, 2 * Math.PI);
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
  color: string
) => {
  // 頭部の円
  ctx.beginPath();
  ctx.arc(position.x, position.y, radius, 0, 2 * Math.PI);
  ctx.fill();

  // 目を描画
  ctx.fillStyle = '#000';
  const eyeRadius = radius * 0.3;
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
  ctx.font = '12px Arial';
  ctx.textAlign = 'center';
  ctx.fillText(name, headPosition.x, headPosition.y - 15);
};

export default GameCanvas;