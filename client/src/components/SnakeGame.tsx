import React, { useState, useEffect } from 'react';
import './SnakeGame.css';
import TouchControls from './TouchControls';
import GameCanvas from './GameCanvas';
import Scoreboard from './Scoreboard';
import { useWebSocket } from '../hooks/useWebSocket';
import { useGameInput } from '../hooks/useGameInput';

const SnakeGame: React.FC = () => {
  const [playerName] = useState<string>('Player');
  const [roomId] = useState<string>('default');
  const [isConnected, setIsConnected] = useState(false);
  const [hasInitiallyConnected, setHasInitiallyConnected] = useState(false);

  // ページ読み込み時に自動接続
  useEffect(() => {
    setIsConnected(true);
    setHasInitiallyConnected(true);
  }, []);

  // カスタムフックを使用
  const { gameState, playerId, sendDirection, isConnecting, gameConfig } = useWebSocket({
    roomId,
    playerName,
    isConnected
  });

  const { handleTouchDirection } = useGameInput({
    onDirectionChange: sendDirection,
    isEnabled: isConnected
  });

  const handleConnect = () => {
    setIsConnected(true);
  };

  const handleDisconnect = () => {
    setIsConnected(false);
  };

  // 接続状況の表示テキスト
  const getConnectionStatus = () => {
    if (isConnecting) return '接続中...';
    if (isConnected && playerId) return '接続済み';
    if (isConnected && !playerId) return '接続中...';
    return '未接続';
  };

  // 接続状況の色
  const getConnectionStatusColor = () => {
    if (isConnecting) return '#FFA500';
    if (isConnected && playerId) return '#4CAF50';
    if (isConnected && !playerId) return '#FFA500';
    return '#F44336';
  };

  return (
    <div className="snake-game">
      {/* 接続状況とボタンのヘッダー */}
      <div className="connection-header">
        <div className="connection-status">
          <span 
            className="status-indicator"
            style={{ color: getConnectionStatusColor() }}
          >
            ● {getConnectionStatus()}
          </span>
        </div>
        <div className="connection-buttons">
          {!isConnected && hasInitiallyConnected ? (
            <button 
              className="connect-button"
              onClick={handleConnect}
              disabled={isConnecting}
            >
              再接続
            </button>
          ) : isConnected ? (
            <button 
              className="disconnect-button"
              onClick={handleDisconnect}
            >
              切断
            </button>
          ) : null}
        </div>
      </div>

      {/* ゲームコンテンツ */}
      {isConnected ? (
        <>
          <div className="game-header">
            <div className="game-title-section">
              <h2>🐍 Snake Game</h2>
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
            gameConfig={gameConfig}
          />
          
          <div className="controls">
            <p className="desktop-controls">矢印キー または WASD で移動</p>
            <TouchControls onDirectionChange={handleTouchDirection} />
          </div>
        </>
      ) : hasInitiallyConnected ? (
        <div className="waiting-screen">
          <h2>🐍 Snake Game</h2>
          <p>接続が切断されました。再接続ボタンを押してゲームに復帰してください</p>
        </div>
      ) : (
        <div className="waiting-screen">
          <h2>🐍 Snake Game</h2>
          <p>ゲームに接続中...</p>
        </div>
      )}
    </div>
  );
};

export default SnakeGame;