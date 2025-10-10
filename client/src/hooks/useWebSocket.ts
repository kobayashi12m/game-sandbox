import { useEffect, useRef, useState, useCallback } from 'react';
import type { 
  GameState, 
  JoinMessage, 
  DirectionMessage, 
  AccelerationMessage,
  StopMovementMessage,
  MouseMoveMessage,
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
  sendAcceleration: (x: number, y: number) => void;
  sendMouseMove: (x: number, y: number) => void;
  sendStopMovement: () => void;
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

  // 加速度送信（360度自由移動用）
  const sendAcceleration = useCallback((x: number, y: number) => {
    const websocket = wsRef.current;
    if (!websocket || websocket.readyState !== WebSocket.OPEN) return;

    const accelerationMessage: AccelerationMessage = {
      type: 'setAcceleration',
      x,
      y
    };
    
    websocket.send(JSON.stringify(accelerationMessage));
  }, []);

  // マウス位置送信（マウス追従移動用）
  const sendMouseMove = useCallback((x: number, y: number) => {
    const websocket = wsRef.current;
    if (!websocket || websocket.readyState !== WebSocket.OPEN) return;

    const mouseMessage: MouseMoveMessage = {
      type: 'mouseMove',
      x,
      y
    };
    
    websocket.send(JSON.stringify(mouseMessage));
  }, []);

  // 移動停止メッセージの送信
  const sendStopMovement = useCallback(() => {
    const websocket = wsRef.current;
    if (!websocket || websocket.readyState !== WebSocket.OPEN) return;

    const stopMessage: StopMovementMessage = {
      type: 'stopMovement'
    };
    
    websocket.send(JSON.stringify(stopMessage));
  }, []);

  return {
    gameState,
    playerId,
    sendDirection,
    sendAcceleration,
    sendMouseMove,
    sendStopMovement,
    isConnecting,
    gameConfig,
    scoreboard
  };
};