// ゲーム関連の型定義

export interface Position {
  x: number;
  y: number;
}

export interface Snake {
  id: string;
  body: Position[];
  color: string;
  alive: boolean;
}

export interface Player {
  id: string;
  name: string;
  snake: Snake;
  score: number;
}

export interface GameState {
  players: Player[];
  food: Position[];
}

// WebSocket メッセージの型
export interface WebSocketMessage {
  type: string;
  [key: string]: unknown;
}

export interface JoinMessage extends WebSocketMessage {
  type: 'join';
  roomId: string;
  playerName: string;
}

export interface DirectionMessage extends WebSocketMessage {
  type: 'changeDirection';
  direction: Direction;
}

export interface GameJoinedMessage extends WebSocketMessage {
  type: 'gameJoined';
  playerId: string;
}

export interface GameStateMessage extends WebSocketMessage {
  type: 'gameState';
  state: GameState;
}

export interface GameConfig {
  fieldWidth: number;
  fieldHeight: number;
  snakeRadius: number;
  foodRadius: number;
  cullingWidth: number;
  cullingHeight: number;
  cullingMargin: number;
}

export interface GameConfigMessage extends WebSocketMessage {
  type: 'gameConfig';
  config: GameConfig;
}

// 方向の型
export type Direction = 'UP' | 'DOWN' | 'LEFT' | 'RIGHT';

// デフォルトのゲーム設定（サーバーから受信するまでの暫定値）
export const DEFAULT_GAME_CONFIG: GameConfig = {
  fieldWidth: 600,
  fieldHeight: 600,
  snakeRadius: 7.5,
  foodRadius: 5,
  cullingWidth: 800,
  cullingHeight: 600,
  cullingMargin: 500,
};