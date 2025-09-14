import React, { useEffect, useRef, useState } from 'react';
import './SnakeGame.css';
import TouchControls from './TouchControls';

const GRID_SIZE = 40;
const CELL_SIZE = 15;

interface Position {
  x: number;
  y: number;
}

interface Snake {
  id: string;
  body: Position[];
  color: string;
  alive: boolean;
}

interface Player {
  id: string;
  name: string;
  snake: Snake;
  score: number;
}

interface GameState {
  players: Player[];
  food: Position[];
}

const SnakeGame: React.FC = () => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [ws, setWs] = useState<WebSocket | null>(null);
  const [gameState, setGameState] = useState<GameState>({ players: [], food: [] });
  const [playerId, setPlayerId] = useState<string>('');
  const [playerName, setPlayerName] = useState<string>('');
  const [roomId, setRoomId] = useState<string>('default');
  const [isConnected, setIsConnected] = useState(false);
  const [showJoinForm, setShowJoinForm] = useState(true);

  useEffect(() => {
    if (!isConnected) return;

    const wsUrl = window.location.hostname === 'localhost' 
      ? 'ws://localhost:8081/ws' 
      : `ws://${window.location.hostname}:8081/ws`;
    const websocket = new WebSocket(wsUrl);

    websocket.onopen = () => {
      const joinMessage = {
        type: 'join',
        roomId: roomId,
        playerName: playerName || 'Player'
      };
      websocket.send(JSON.stringify(joinMessage));
    };

    websocket.onmessage = (event) => {
      const message = JSON.parse(event.data);
      
      switch (message.type) {
        case 'gameJoined':
          setPlayerId(message.playerId);
          break;
        case 'gameState':
          setGameState(message.state);
          break;
        case 'gameInit':
          // 古いフォーマットの処理（互換性のため）
          if (message.data && message.data.id) {
            setPlayerId(message.data.id);
          }
          break;
      }
    };

    websocket.onclose = () => {
      setIsConnected(false);
      setShowJoinForm(true);
    };

    websocket.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    setWs(websocket);

    return () => {
      websocket.close();
    };
  }, [isConnected, roomId, playerName]);

  useEffect(() => {
    const handleKeyPress = (e: KeyboardEvent) => {
      if (!ws || ws.readyState !== WebSocket.OPEN) return;

      let direction = '';
      switch (e.key) {
        case 'ArrowUp':
        case 'w':
        case 'W':
          direction = 'UP';
          break;
        case 'ArrowDown':
        case 's':
        case 'S':
          direction = 'DOWN';
          break;
        case 'ArrowLeft':
        case 'a':
        case 'A':
          direction = 'LEFT';
          break;
        case 'ArrowRight':
        case 'd':
        case 'D':
          direction = 'RIGHT';
          break;
      }

      if (direction) {
        ws.send(JSON.stringify({
          type: 'changeDirection',
          direction: direction
        }));
      }
    };

    window.addEventListener('keydown', handleKeyPress);
    return () => window.removeEventListener('keydown', handleKeyPress);
  }, [ws]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    
    // Clear canvas
    ctx.fillStyle = '#1a1a1a';
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    // Draw grid
    ctx.strokeStyle = '#333';
    ctx.lineWidth = 0.5;
    for (let i = 0; i <= GRID_SIZE; i++) {
      ctx.beginPath();
      ctx.moveTo(i * CELL_SIZE, 0);
      ctx.lineTo(i * CELL_SIZE, GRID_SIZE * CELL_SIZE);
      ctx.stroke();
      
      ctx.beginPath();
      ctx.moveTo(0, i * CELL_SIZE);
      ctx.lineTo(GRID_SIZE * CELL_SIZE, i * CELL_SIZE);
      ctx.stroke();
    }

    // Draw food
    gameState.food.forEach(food => {
      ctx.fillStyle = '#ff6b6b';
      ctx.beginPath();
      ctx.arc(
        food.x * CELL_SIZE + CELL_SIZE / 2,
        food.y * CELL_SIZE + CELL_SIZE / 2,
        CELL_SIZE / 3,
        0,
        2 * Math.PI
      );
      ctx.fill();
    });

    // Draw snakes
    gameState.players.forEach(player => {
      const snake = player.snake;
      if (!snake.alive) {
        ctx.globalAlpha = 0.3;
      }

      // Draw body
      snake.body.forEach((segment, index) => {
        ctx.fillStyle = snake.color;
        if (index === 0) {
          // Head
          ctx.fillRect(
            segment.x * CELL_SIZE + 1,
            segment.y * CELL_SIZE + 1,
            CELL_SIZE - 2,
            CELL_SIZE - 2
          );
          // Eyes
          ctx.fillStyle = '#000';
          ctx.fillRect(
            segment.x * CELL_SIZE + 3,
            segment.y * CELL_SIZE + 3,
            3,
            3
          );
          ctx.fillRect(
            segment.x * CELL_SIZE + CELL_SIZE - 6,
            segment.y * CELL_SIZE + 3,
            3,
            3
          );
        } else {
          // Body segments
          ctx.fillRect(
            segment.x * CELL_SIZE + 2,
            segment.y * CELL_SIZE + 2,
            CELL_SIZE - 4,
            CELL_SIZE - 4
          );
        }
      });

      ctx.globalAlpha = 1;

      // Draw player name
      if (snake.body.length > 0) {
        ctx.fillStyle = '#fff';
        ctx.font = '12px Arial';
        ctx.textAlign = 'center';
        ctx.fillText(
          player.name,
          snake.body[0].x * CELL_SIZE + CELL_SIZE / 2,
          snake.body[0].y * CELL_SIZE - 5
        );
      }
    });
  }, [gameState]);

  const handleJoin = (e: React.FormEvent) => {
    e.preventDefault();
    setIsConnected(true);
    setShowJoinForm(false);
  };

  const handleDirectionChange = (direction: string) => {
    if (!ws || ws.readyState !== WebSocket.OPEN) return;
    
    ws.send(JSON.stringify({
      type: 'changeDirection',
      direction: direction
    }));
  };

  if (showJoinForm) {
    return (
      <div className="join-form">
        <h2>Snake Game</h2>
        <form onSubmit={handleJoin}>
          <input
            type="text"
            placeholder="Your name"
            value={playerName}
            onChange={(e) => setPlayerName(e.target.value)}
          />
          <input
            type="text"
            placeholder="Room ID"
            value={roomId}
            onChange={(e) => setRoomId(e.target.value)}
          />
          <button type="submit">Join Game</button>
        </form>
      </div>
    );
  }

  return (
    <div className="snake-game">
      <div className="game-header">
        <h2>Snake Game - Room: {roomId}</h2>
        <div className="scoreboard">
          <h3>Scores</h3>
          {gameState.players.map(player => (
            <div key={player.id} className="player-score">
              <span style={{ color: player.snake.color }}>●</span>
              {player.name}: {player.score}
              {player.id === playerId && ' (You)'}
            </div>
          ))}
        </div>
      </div>
      <canvas
        ref={canvasRef}
        width={GRID_SIZE * CELL_SIZE}
        height={GRID_SIZE * CELL_SIZE}
        className="game-canvas"
      />
      <div className="controls">
        <p className="desktop-controls">Use Arrow Keys or WASD to move</p>
        <TouchControls onDirectionChange={handleDirectionChange} />
      </div>
    </div>
  );
};

export default SnakeGame;