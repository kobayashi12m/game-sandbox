// ゲーム関連の型定義

export interface Position {
  x: number;
  y: number;
}

// 球体の物理特性 (キー短縮対応)
export interface Sphere {
  p: Position;      // position → p
  v?: Position;     // velocity → v  
  a?: Position;     // acceleration → a
  r: number;        // radius → r
  mass: number;
}

// 軌道上を回転する衛星
export interface Satellite {
  node: Sphere;
  angle: number;
  orbitalSpeed: number;
  radius: number;
}

// 核と衛星からなる天体システム (キー短縮対応)
export interface CelestialSystem {
  c: Sphere;        // core → c
  n: Sphere[];      // nodes → n  
  satellites: Satellite[];
  col: string;      // color → col
  a: boolean;       // alive → a
}

export interface Player {
  id: string;
  nm: string;               // name → nm
  cel: CelestialSystem;     // celestial → cel
  sc: number;               // score → sc
}

export interface Projectile {
  id: string;
  sph: Sphere;              // sphere → sph
  oid: string;              // ownerId → oid
}

export interface DroppedSatellite {
  p: Position;    // position → p
  r: number;      // radius → r
}

export interface GameState {
  pls: Player[];                     // players → pls
  ds?: DroppedSatellite[];          // droppedSatellites → ds
  proj?: Projectile[];              // projectiles → proj
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
  cullingWidth: 1300,
  cullingHeight: 800,
};