import { useEffect, useRef, useState, useCallback } from 'react';
import type { 
  GameState, 
  JoinMessage, 
  AccelerationMessage,
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
  sendAcceleration: (x: number, y: number) => void;
  sendEjectSatellite: (targetX: number, targetY: number) => void;
  isConnecting: boolean;
  gameConfig: GameConfig;
  scoreboard: ScoreInfo[];
  myScore: ScoreInfo | null;
}

export const useWebSocket = ({ 
  roomId, 
  playerName, 
  isConnected 
}: UseWebSocketProps): UseWebSocketReturn => {
  const [gameState, setGameState] = useState<GameState>({ pls: [] });
  const [playerId, setPlayerId] = useState<string>('');
  const [isConnecting, setIsConnecting] = useState(false);
  const [gameConfig, setGameConfig] = useState<GameConfig>(DEFAULT_GAME_CONFIG);
  const [scoreboard, setScoreboard] = useState<ScoreInfo[]>([]);
  const [myScore, setMyScore] = useState<ScoreInfo | null>(null);
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
          setScoreboard(scoreboardMsg.scoreboard);
          
          // 自分のスコア情報を保存
          if (scoreboardMsg.myScore) {
            setMyScore(scoreboardMsg.myScore);
          }
        }
        break;
    }
  }, []);

  // WebSocket接続の確立
  useEffect(() => {
    console.log('[WS] useEffect called, isConnected:', isConnected);
    if (!isConnected) {
      console.log('[WS] isConnected is false, skipping');
      return;
    }

    setIsConnecting(true);

    // 接続先
    const wsUrl = import.meta.env.VITE_WS_URL || `ws://${window.location.hostname}:8081/ws`;

    console.log('[WS] Creating WebSocket:', wsUrl);
    const websocket = new WebSocket(wsUrl);
    wsRef.current = websocket;
    console.log('[WS] Created, readyState:', websocket.readyState);

    // ハンドラー設定関数（リトライ時にも使用）
    const setupHandlers = (ws: WebSocket, clearTimeoutFn: () => void) => {
      ws.onopen = () => {
        console.log('[WS] onopen fired');
        clearTimeoutFn();
        setIsConnecting(false);

        const joinMessage: JoinMessage = {
          type: 'join',
          roomId,
          playerName
        };
        console.log('[WS] Sending join');
        wsRef.current?.send(JSON.stringify(joinMessage));
      };

      ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          handleMessage(message);
        } catch (error) {
          console.error('📡 JSON parse error:', error, event.data);
        }
      };

      ws.onclose = (event) => {
        console.log('[WS] onclose, code:', event.code, 'reason:', event.reason);
        setIsConnecting(false);
        setPlayerId('');
        setGameState({ pls: [] });
        setScoreboard([]);
      };

      ws.onerror = (event) => {
        console.log('[WS] onerror:', event);
        setIsConnecting(false);
      };
    };

    // 接続タイムアウト（3秒以内に接続できなければ再試行）
    const connectionTimeout = setTimeout(() => {
      if (wsRef.current?.readyState === WebSocket.CONNECTING) {
        console.log('[WS] Connection timeout, retrying...');
        wsRef.current.close();
        // 再接続
        const retryWs = new WebSocket(wsUrl);
        wsRef.current = retryWs;
        setupHandlers(retryWs, () => {});
      }
    }, 3000);

    setupHandlers(websocket, () => clearTimeout(connectionTimeout));

    // ページ離脱時にWebSocketを強制クローズ
    const handlePageHide = () => {
      console.log('[WS] pagehide, closing');
      wsRef.current?.close();
    };
    window.addEventListener('pagehide', handlePageHide);

    return () => {
      console.log('[WS] Cleanup, closing');
      clearTimeout(connectionTimeout);
      window.removeEventListener('pagehide', handlePageHide);
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, [isConnected, roomId, playerName, handleMessage]);


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



  // 衛星射出メッセージの送信
  const sendEjectSatellite = useCallback((targetX: number, targetY: number) => {
    const websocket = wsRef.current;
    if (!websocket || websocket.readyState !== WebSocket.OPEN) return;

    const ejectMessage = {
      type: 'ejectSatellite',
      targetX,
      targetY
    };
    
    websocket.send(JSON.stringify(ejectMessage));
  }, []);

  return {
    gameState,
    playerId,
    sendAcceleration,
    sendEjectSatellite,
    isConnecting,
    gameConfig,
    scoreboard,
    myScore
  };
};