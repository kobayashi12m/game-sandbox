import React, { useState } from 'react';
import './JoinForm.css';

interface JoinFormProps {
  onJoin: (playerName: string, roomId: string) => void;
  isConnecting?: boolean;
}

const JoinForm: React.FC<JoinFormProps> = ({ onJoin, isConnecting = false }) => {
  const [playerName, setPlayerName] = useState<string>('');
  const [roomId, setRoomId] = useState<string>('default');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!isConnecting && playerName.trim() && roomId.trim()) {
      onJoin(playerName.trim(), roomId.trim());
    }
  };

  return (
    <div className="join-form">
      <div className="join-form-container">
        <h1 className="game-title">🐍 Snake Game</h1>
        <p className="game-subtitle">マルチプレイヤー対応のスネークゲーム</p>
        
        <form onSubmit={handleSubmit} className="join-form-content">
          <div className="input-group">
            <label htmlFor="playerName">プレイヤー名</label>
            <input
              id="playerName"
              type="text"
              placeholder="あなたの名前を入力"
              value={playerName}
              onChange={(e) => setPlayerName(e.target.value)}
              disabled={isConnecting}
              maxLength={20}
              required
            />
          </div>
          
          <div className="input-group">
            <label htmlFor="roomId">ルームID</label>
            <input
              id="roomId"
              type="text"
              placeholder="ルーム名を入力"
              value={roomId}
              onChange={(e) => setRoomId(e.target.value)}
              disabled={isConnecting}
              maxLength={50}
              required
            />
          </div>
          
          <button 
            type="submit" 
            className={`join-button ${isConnecting ? 'connecting' : ''}`}
            disabled={isConnecting || !playerName.trim() || !roomId.trim()}
          >
            {isConnecting ? (
              <>
                <span className="spinner"></span>
                接続中...
              </>
            ) : (
              'ゲームに参加'
            )}
          </button>
        </form>
        
        <div className="game-instructions">
          <h3>操作方法</h3>
          <ul>
            <li>矢印キー または WASD で移動</li>
            <li>スマホの場合はタッチコントロール</li>
            <li>食べ物を食べてスコアを上げよう！</li>
            <li>他の蛇や自分の体に当たると死亡</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default JoinForm;