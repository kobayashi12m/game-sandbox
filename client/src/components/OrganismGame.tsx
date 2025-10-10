import React, { useState, useEffect } from "react";
import "./OrganismGame.css";
import TouchControls from "./TouchControls";
import GameCanvas from "./GameCanvas";
import Scoreboard from "./Scoreboard";
import { useWebSocket } from "../hooks/useWebSocket";
import { useGameInput } from "../hooks/useGameInput";

const OrganismGame: React.FC = () => {
  const [playerName] = useState<string>("Player");
  const [roomId] = useState<string>("default");
  const [isConnected, setIsConnected] = useState(false);
  const [hasInitiallyConnected, setHasInitiallyConnected] = useState(false);

  // ページ読み込み時に自動接続
  useEffect(() => {
    setIsConnected(true);
    setHasInitiallyConnected(true);
  }, []);

  // カスタムフックを使用
  const {
    gameState,
    playerId,
    sendDirection,
    sendAcceleration,
    sendStopMovement,
    isConnecting,
    gameConfig,
    scoreboard,
  } = useWebSocket({
    roomId,
    playerName,
    isConnected,
  });

  const { handleTouchDirection } = useGameInput({
    onDirectionChange: sendDirection,
    onAccelerationChange: sendAcceleration,
    onMovementStop: sendStopMovement,
    isEnabled: isConnected,
  });

  const handleConnect = () => {
    setIsConnected(true);
  };

  const handleDisconnect = () => {
    setIsConnected(false);
  };

  // 接続状況の表示テキスト
  const getConnectionStatus = () => {
    if (isConnecting) return "接続中...";
    if (isConnected && playerId) return "接続済み";
    if (isConnected && !playerId) return "接続中...";
    return "未接続";
  };

  // 接続状況の色
  const getConnectionStatusColor = () => {
    if (isConnecting) return "#FFA500";
    if (isConnected && playerId) return "#4CAF50";
    if (isConnected && !playerId) return "#FFA500";
    return "#F44336";
  };

  return (
    <div className="organism-game">
      {/* ゲームコンテンツ */}
      {isConnected ? (
        <div className="game-container">
          {/* メインのゲーム画面 */}
          <GameCanvas
            gameState={gameState}
            playerId={playerId}
            gameConfig={gameConfig}
          />

          {/* オーバーレイUI */}
          <div className="overlay-ui">
            {/* 接続状況（左上） */}
            <div className="connection-status-overlay">
              <span
                className="status-indicator"
                style={{ color: getConnectionStatusColor() }}
              >
                ● {getConnectionStatus()}
              </span>
              {isConnected ? (
                <button
                  className="disconnect-button-small"
                  onClick={handleDisconnect}
                >
                  切断
                </button>
              ) : (
                <button
                  className="connect-button-small"
                  onClick={handleConnect}
                  disabled={isConnecting}
                >
                  再接続
                </button>
              )}
            </div>

            {/* リーダーボード（右上） */}
            <div className="leaderboard-overlay">
              <Scoreboard
                players={scoreboard}
                currentPlayerId={playerId}
                roomId={roomId}
              />
            </div>
          </div>

          {/* タッチコントロール */}
          <div className="touch-controls-overlay">
            <TouchControls onDirectionChange={handleTouchDirection} />
          </div>
        </div>
      ) : hasInitiallyConnected ? (
        <div className="waiting-screen">
          <h2>🔵 Organism Game</h2>
          <p>
            接続が切断されました。再接続ボタンを押してゲームに復帰してください
          </p>
          <button
            className="connect-button"
            onClick={handleConnect}
            disabled={isConnecting}
          >
            再接続
          </button>
        </div>
      ) : (
        <div className="waiting-screen">
          <h2>🔵 Organism Game</h2>
          <p>ゲームに接続中...</p>
        </div>
      )}
    </div>
  );
};

export default OrganismGame;
