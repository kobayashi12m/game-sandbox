import { useEffect, useRef, useState, useCallback } from 'react';
import type { 
  GameState, 
  JoinMessage, 
  DirectionMessage, 
  Direction,
  WebSocketMessage 
} from '../types';

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
}

export const useWebSocket = ({ 
  roomId, 
  playerName, 
  isConnected 
}: UseWebSocketProps): UseWebSocketReturn => {
  const [gameState, setGameState] = useState<GameState>({ players: [], food: [] });
  const [playerId, setPlayerId] = useState<string>('');
  const [isConnecting, setIsConnecting] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

  // メッセージハンドラー
  const handleMessage = useCallback((message: WebSocketMessage) => {
    switch (message.type) {
      case 'gameJoined':
        if (typeof message.playerId === 'string') {
          setPlayerId(message.playerId);
          console.log('ゲームに参加しました, プレイヤーID:', message.playerId);
        }
        break;
      
      case 'gameState':
        if (message.state && typeof message.state === 'object') {
          setGameState(message.state as GameState);
        }
        break;
      
      case 'gameInit':
        // 後方互換性のため
        if (message.data && typeof message.data === 'object' && 
            message.data !== null && 'id' in message.data && 
            typeof (message.data as Record<string, unknown>).id === 'string') {
          setPlayerId((message.data as Record<string, unknown>).id as string);
        }
        break;
      
      default:
        console.log('不明なメッセージタイプ:', message.type);
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
      console.log('WebSocket接続が確立されました');
      setIsConnecting(false);
      
      const joinMessage: JoinMessage = {
        type: 'join',
        roomId,
        playerName: playerName || 'Player'
      };
      websocket.send(JSON.stringify(joinMessage));
    };

    websocket.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        handleMessage(message);
      } catch (error) {
        console.error('メッセージの解析に失敗しました:', error);
      }
    };

    websocket.onclose = () => {
      console.log('WebSocket接続が閉じられました');
      setIsConnecting(false);
      setPlayerId('');
      setGameState({ players: [], food: [] });
    };

    websocket.onerror = (error) => {
      console.error('WebSocketエラー:', error);
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
    if (!websocket || websocket.readyState !== WebSocket.OPEN) {
      console.warn('WebSocket接続が利用できません');
      return;
    }

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
    isConnecting
  };
};