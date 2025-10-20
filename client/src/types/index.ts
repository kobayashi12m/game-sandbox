// ゲーム関連の型定義

export interface Position {
  x: number;
  y: number;
}

// 球体の物理特性
export interface Sphere {
  position: Position;
  velocity?: Position;
  acceleration?: Position;
  radius: number;
  mass: number;
}

// 軌道上を回転する衛星
export interface Satellite {
  node: Sphere;
  angle: number;
  orbitalSpeed: number;
  radius: number;
}

// 核と衛星からなる天体システム
export interface CelestialSystem {
  core: Sphere;
  nodes: Sphere[];
  satellites: Satellite[];
  color: string;
  alive: boolean;
}

export interface Player {
  id: string;
  name: string;
  celestial: CelestialSystem;
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

export interface AccelerationMessage extends WebSocketMessage {
  type: 'setAcceleration';
  x: number;
  y: number;
}

export interface StopMovementMessage extends WebSocketMessage {
  type: 'stopMovement';
}

export interface MouseMoveMessage extends WebSocketMessage {
  type: 'mouseMove';
  x: number;
  y: number;
}

export interface GameJoinedMessage extends WebSocketMessage {
  type: 'gameJoined';
  playerId: string;
}

export interface GameStateMessage extends WebSocketMessage {
  type: 'gameState';
  state: GameState;
  scoreboard?: ScoreInfo[];
}

export interface ScoreInfo {
  id: string;
  name: string;
  score: number;
  alive: boolean;
  color: string;
}

export interface ScoreboardMessage extends WebSocketMessage {
  type: 'scoreboard';
  scoreboard: {
    players: ScoreInfo[];
  };
}

export interface GridLine {
  startX: number;
  startY: number;
  endX: number;
  endY: number;
}

export interface GameConfig {
  fieldWidth: number;
  fieldHeight: number;
  sphereRadius: number;
  foodRadius: number;
  cullingWidth: number;
  cullingHeight: number;
  gridLines?: GridLine[];
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
  sphereRadius: 7.5,
  foodRadius: 5,
  cullingWidth: 1300,
  cullingHeight: 800,
};