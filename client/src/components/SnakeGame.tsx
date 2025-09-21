import React, { useState } from 'react';
import './SnakeGame.css';
import TouchControls from './TouchControls';
import GameCanvas from './GameCanvas';
import Scoreboard from './Scoreboard';
import JoinForm from './JoinForm';
import { useWebSocket } from '../hooks/useWebSocket';
import { useGameInput } from '../hooks/useGameInput';

const SnakeGame: React.FC = () => {
  const [playerName, setPlayerName] = useState<string>('');
  const [roomId, setRoomId] = useState<string>('default');
  const [isConnected, setIsConnected] = useState(false);

  // カスタムフックを使用
  const { gameState, playerId, sendDirection, isConnecting } = useWebSocket({
    roomId,
    playerName,
    isConnected
  });

  const { handleTouchDirection } = useGameInput({
    onDirectionChange: sendDirection,
    isEnabled: isConnected
  });

  const handleJoin = (name: string, room: string) => {
    setPlayerName(name);
    setRoomId(room);
    setIsConnected(true);
  };

  const handleDisconnect = () => {
    setIsConnected(false);
  };

  if (!isConnected) {
    return (
      <JoinForm 
        onJoin={handleJoin} 
        isConnecting={isConnecting}
      />
    );
  }

  return (
    <div className="snake-game">
      <div className="game-header">
        <div className="game-title-section">
          <h2>🐍 Snake Game</h2>
          <button 
            className="disconnect-button"
            onClick={handleDisconnect}
            title="ゲームから退出"
          >
            退出
          </button>
        </div>
        <Scoreboard 
          players={gameState.players}
          currentPlayerId={playerId}
          roomId={roomId}
        />
      </div>
      
      <GameCanvas 
        gameState={gameState}
        playerId={playerId}
      />
      
      <div className="controls">
        <p className="desktop-controls">矢印キー または WASD で移動</p>
        <TouchControls onDirectionChange={handleTouchDirection} />
      </div>
    </div>
  );
};

export default SnakeGame;