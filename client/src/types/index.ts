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

// 方向の型
export type Direction = 'UP' | 'DOWN' | 'LEFT' | 'RIGHT';

// ゲーム設定
export const GAME_CONFIG = {
  GRID_SIZE: 40,
  CELL_SIZE: 15,
} as const;