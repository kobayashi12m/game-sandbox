import { memo } from "react";
import type { ScoreInfo } from "../types";
import "./Scoreboard.css";

interface ScoreboardProps {
  players: ScoreInfo[];
  currentPlayerId: string;
  myScore?: ScoreInfo | null;
}

// プレイヤー情報を表示する共通コンポーネント
const PlayerInfo: React.FC<{ player: ScoreInfo; showStatus?: boolean }> = ({
  player,
  showStatus = true,
}) => (
  <div className="player-info">
    <span className="player-color" style={{ backgroundColor: player.color }} />
    <span className="player-name">{player.name}</span>
    {showStatus && (
      <span className={`player-status ${player.alive ? "alive" : "dead"}`}>
        {player.alive ? "生存" : "死亡"}
      </span>
    )}
  </div>
);

const Scoreboard: React.FC<ScoreboardProps> = memo(
  ({ players, currentPlayerId, myScore }) => {
    return (
      <div className="scoreboard">
        <div className="scoreboard-header">
          <h3>スコア</h3>
        </div>

        {/* 自分のスコア表示（ラベルなし） */}
        {myScore && (
          <div className="my-score-card">
            <PlayerInfo player={myScore} />
            <div className="my-score-value">{myScore.score}</div>
          </div>
        )}

        <div className="ranking-section">
          <div className="ranking-title">トップ10</div>
          <div className="players-list">
            {players.map((player, index) => (
              <div
                key={player.id}
                className={`player-score ${
                  player.id === currentPlayerId ? "current-player" : ""
                }`}
              >
                <div className="player-rank">#{index + 1}</div>
                <PlayerInfo
                  player={{
                    ...player,
                    name:
                      player.name +
                      (player.id === currentPlayerId ? " (あなた)" : ""),
                  }}
                />
                <div className="player-score-value">{player.score}</div>
              </div>
            ))}
          </div>
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
