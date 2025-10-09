import { useEffect, useRef, useState, useCallback } from 'react';
import type { 
  GameState, 
  JoinMessage, 
  DirectionMessage, 
  Direction,
  WebSocketMessage,
  GameConfig,
  GameConfigMessage,
  ScoreInfo,
  ScoreboardMessage
} from '../types';
import { DEFAULT_GAME_CONFIG } from '../types';

interface UseWebSocketProps {
  roomId: string;
  playerName: string;
  isConnected: boolean;
}

interface UseWebSocketReturn {
  gameState: GameState;
  playerId: string;
  sendDirection: (direction: Direction) => void;
  isConnecting: boolean;
  gameConfig: GameConfig;
  scoreboard: ScoreInfo[];
}

export const useWebSocket = ({ 
  roomId, 
  playerName, 
  isConnected 
}: UseWebSocketProps): UseWebSocketReturn => {
  const [gameState, setGameState] = useState<GameState>({ players: [], food: [] });
  const [playerId, setPlayerId] = useState<string>('');
  const [isConnecting, setIsConnecting] = useState(false);
  const [gameConfig, setGameConfig] = useState<GameConfig>(DEFAULT_GAME_CONFIG);
  const [scoreboard, setScoreboard] = useState<ScoreInfo[]>([]);
  const wsRef = useRef<WebSocket | null>(null);

  // メッセージハンドラー
  const handleMessage = useCallback((message: WebSocketMessage) => {
    switch (message.type) {
      case 'gameJoined':
        if ('playerId' in message && typeof message.playerId === 'string') {
          setPlayerId(message.playerId);
        }
        break;
      
      case 'gameState':
        if ('state' in message) {
          setGameState(message.state as GameState);
        }
        break;
        
      case 'gameConfig':
        if ('config' in message) {
          const configMsg = message as GameConfigMessage;
          setGameConfig(configMsg.config);
        }
        break;

      case 'scoreboard':
        if ('scoreboard' in message) {
          const scoreboardMsg = message as ScoreboardMessage;
          setScoreboard(scoreboardMsg.scoreboard.players);
        }
        break;
    }
  }, []);

  // WebSocket接続の確立
  useEffect(() => {
    if (!isConnected) return;

    setIsConnecting(true);
    
    const wsUrl = window.location.hostname === 'localhost' 
      ? 'ws://localhost:8081/ws' 
      : `ws://${window.location.hostname}:8081/ws`;
    
    const websocket = new WebSocket(wsUrl);
    wsRef.current = websocket;

    websocket.onopen = () => {
      setIsConnecting(false);
      
      const joinMessage: JoinMessage = {
        type: 'join',
        roomId,
        playerName
      };
      websocket.send(JSON.stringify(joinMessage));
    };

    websocket.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        handleMessage(message);
      } catch {
        // JSON parse error - ignore malformed messages
      }
    };

    websocket.onclose = () => {
      setIsConnecting(false);
      setPlayerId('');
      setGameState({ players: [], food: [] });
      setScoreboard([]);
    };

    websocket.onerror = () => {
      setIsConnecting(false);
    };

    return () => {
      websocket.close();
      wsRef.current = null;
    };
  }, [isConnected, roomId, playerName, handleMessage]);

  // 方向変更メッセージの送信
  const sendDirection = useCallback((direction: Direction) => {
    const websocket = wsRef.current;
    if (!websocket || websocket.readyState !== WebSocket.OPEN) return;

    const directionMessage: DirectionMessage = {
      type: 'changeDirection',
      direction
    };
    
    websocket.send(JSON.stringify(directionMessage));
  }, []);

  return {
    gameState,
    playerId,
    sendDirection,
    isConnecting,
    gameConfig,
    scoreboard
  };
};