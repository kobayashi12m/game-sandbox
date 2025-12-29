import React, { useState, useEffect } from "react";
import "./CelestialGame.css";
import GameCanvas from "./GameCanvas";
import Scoreboard from "./Scoreboard";
import { VectorDisplay } from "./VectorDisplay";
import { useWebSocket } from "../hooks/useWebSocket";
import { useGameInput } from "../hooks/useGameInput";
import { getPlayer } from "../types";
import { calculateViewportScale } from "../utils/viewport";
import { PLAYER_CONFIG } from "../constants/game";

const CelestialGame: React.FC = () => {
  const [isConnected, setIsConnected] = useState(false);
  const [hasInitiallyConnected, setHasInitiallyConnected] = useState(false);

  // ビューポートスケール計算
  const [viewportScale, setViewportScale] = useState(calculateViewportScale);

  // ページ読み込み時に自動接続
  useEffect(() => {
    setIsConnected(true);
    setHasInitiallyConnected(true);
  }, []);

  // ウィンドウリサイズ時のスケール更新
  useEffect(() => {
    const handleResize = () => {
      setViewportScale(calculateViewportScale());
    };

    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  // カスタムフックを使用
  const {
    gameState,
    playerId,
    sendAcceleration,
    sendEjectSatellite,
    isConnecting,
    gameConfig,
    scoreboard,
    myScore,
  } = useWebSocket({
    roomId: PLAYER_CONFIG.ROOM_ID,
    playerName: PLAYER_CONFIG.PLAYER_NAME,
    isConnected,
  });

  // ゲーム入力処理
  useGameInput({
    onAccelerationChange: sendAcceleration,
    onMovementStop: () => sendAcceleration(0, 0),
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
    <div className="celestial-game">
      {/* ゲームコンテンツ */}
      {isConnected ? (
        <div
          className="game-container"
          style={{
            transform: `scale(${viewportScale})`,
            transformOrigin: "center center",
            width: `${1920}px`,
            height: `${1080}px`,
            position: "absolute",
            top: "50%",
            left: "50%",
            marginTop: `${-1080 / 2}px`,
            marginLeft: `${-1920 / 2}px`,
          }}
        >
          {/* メインのゲーム画面 */}
          <GameCanvas
            gameState={gameState}
            playerId={playerId}
            gameConfig={gameConfig}
            onMouseMove={sendAcceleration}
            onMouseClick={sendEjectSatellite}
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
                myScore={myScore}
              />
            </div>

            {/* ベクトル表示（左下） */}
            <div className="vector-display-overlay">
              {gameState &&
                playerId &&
                (() => {
                  const currentPlayer = gameState.pls.find(
                    (p) => p[0] === playerId
                  );
                  if (currentPlayer) {
                    const playerData = getPlayer(currentPlayer);
                    if (playerData?.cel?.c) {
                      const velocity = playerData.cel.c.v;
                      const acceleration = playerData.cel.c.a;
                      return (
                        <VectorDisplay
                          velocity={
                            velocity ? [velocity.x, velocity.y] : undefined
                          }
                          acceleration={
                            acceleration
                              ? [acceleration.x, acceleration.y]
                              : undefined
                          }
                          maxSpeed={500}
                          npcDebug={gameState?.npcDebug}
                        />
                      );
                    }
                  }
                  return null;
                })()}
            </div>
          </div>
        </div>
      ) : hasInitiallyConnected ? (
        <div className="waiting-screen">
          <h2>🔵 celestial Game</h2>
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
          <h2>🔵 celestial Game</h2>
          <p>ゲームに接続中...</p>
        </div>
      )}
    </div>
  );
};

export default CelestialGame;
