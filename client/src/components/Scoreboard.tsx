import React from 'react';
import type { Player } from '../types';
import './Scoreboard.css';

interface ScoreboardProps {
  players: Player[];
  currentPlayerId: string;
  roomId: string;
}

const Scoreboard: React.FC<ScoreboardProps> = ({ players, currentPlayerId, roomId }) => {
  // プレイヤーをスコア順にソート
  const sortedPlayers = [...players].sort((a, b) => b.score - a.score);

  return (
    <div className="scoreboard">
      <div className="scoreboard-header">
        <h3>スコア</h3>
        <div className="room-info">ルーム: {roomId}</div>
      </div>
      
      <div className="players-list">
        {sortedPlayers.map((player, index) => (
          <div 
            key={player.id} 
            className={`player-score ${player.id === currentPlayerId ? 'current-player' : ''}`}
          >
            <div className="player-rank">#{index + 1}</div>
            <div className="player-info">
              <span 
                className="player-color" 
                style={{ backgroundColor: player.snake.color }}
              />
              <span className="player-name">
                {player.name}
                {player.id === currentPlayerId && ' (あなた)'}
              </span>
              <span className={`player-status ${player.snake.alive ? 'alive' : 'dead'}`}>
                {player.snake.alive ? '生存' : '死亡'}
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
};

export default Scoreboard;