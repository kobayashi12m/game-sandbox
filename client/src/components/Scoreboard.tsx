import React, { memo } from "react";
import type { ScoreInfo } from "../types";
import "./Scoreboard.css";

interface ScoreboardProps {
  players: ScoreInfo[];
  currentPlayerId: string;
  roomId: string;
}

const Scoreboard: React.FC<ScoreboardProps> = memo(
  ({ players, currentPlayerId, roomId }) => {
    return (
      <div className="scoreboard">
        <div className="scoreboard-header">
          <h3>スコア</h3>
          <div className="room-info">ルーム: {roomId}</div>
        </div>

        <div className="players-list">
          {players.map((player, index) => (
            <div
              key={player.id}
              className={`player-score ${
                player.id === currentPlayerId ? "current-player" : ""
              }`}
            >
              <div className="player-rank">#{index + 1}</div>
              <div className="player-info">
                <span
                  className="player-color"
                  style={{ backgroundColor: player.color }}
                />
                <span className="player-name">
                  {player.name}
                  {player.id === currentPlayerId && " (あなた)"}
                </span>
                <span
                  className={`player-status ${player.alive ? "alive" : "dead"}`}
                >
                  {player.alive ? "生存" : "死亡"}
                </span>
              </div>
              <div className="player-score-value">{player.score}</div>
            </div>
          ))}
        </div>

        {players.length === 0 && (
          <div className="no-players">プレイヤーを待機中...</div>
        )}
      </div>
    );
  }
);

Scoreboard.displayName = "Scoreboard";

export default Scoreboard;
