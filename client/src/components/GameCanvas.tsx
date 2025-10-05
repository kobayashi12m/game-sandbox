import React, { useRef, useEffect } from 'react';
import type { GameState, GameConfig } from '../types';

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

    // キャンバスをクリア
    ctx.fillStyle = '#1a1a1a';
    ctx.fillRect(0, 0, gameConfig.fieldWidth, gameConfig.fieldHeight);

    // 食べ物を描画
    ctx.fillStyle = '#ff6b6b';
    gameState.food.forEach(food => {
      ctx.beginPath();
      ctx.arc(food.x, food.y, gameConfig.foodRadius, 0, 2 * Math.PI);
      ctx.fill();
    });

    // プレイヤーを描画
    gameState.players.forEach(player => {
      const snake = player.snake;
      
      // 死んでいる蛇は半透明に
      ctx.globalAlpha = snake.alive ? 1 : 0.3;
      ctx.fillStyle = snake.color;

      // 蛇の体を描画
      snake.body.forEach((segment, index) => {
        ctx.beginPath();
        ctx.arc(segment.x, segment.y, gameConfig.snakeRadius, 0, 2 * Math.PI);
        ctx.fill();

        // 頭部には目を追加
        if (index === 0) {
          ctx.fillStyle = '#000';
          ctx.beginPath();
          ctx.arc(segment.x - 3, segment.y - 3, 2, 0, 2 * Math.PI);
          ctx.fill();
          ctx.beginPath();
          ctx.arc(segment.x + 3, segment.y - 3, 2, 0, 2 * Math.PI);
          ctx.fill();
          ctx.fillStyle = snake.color;
        }
      });

      // プレイヤー名を表示
      if (snake.body.length > 0) {
        const head = snake.body[0];
        ctx.globalAlpha = 1;
        ctx.fillStyle = player.id === playerId ? '#ffd700' : '#fff';
        ctx.font = '12px Arial';
        ctx.textAlign = 'center';
        ctx.fillText(player.name, head.x, head.y - 15);
      }
    });

    ctx.globalAlpha = 1;
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

export default GameCanvas;