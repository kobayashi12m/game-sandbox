// ゲーム関連の型定義

// Positionは配列形式: [x, y]
export type Position = [number, number];

// Sphereは配列形式: [[x,y], radius, color, [vx,vy]?, [ax,ay]?]
export type Sphere = [Position, number, string, Position?, Position?];

// 軌道上を回転する衛星
export interface Satellite {
  node: Sphere;
  angle: number;
  orbitalSpeed: number;
  radius: number;
}

// CelestialSystemは配列形式: [core, color, alive, nodes]
export type CelestialSystem = [Sphere, string, boolean, Sphere[]];

// Playerは配列形式: [id, name, celestial, score]
export type Player = [string, string, CelestialSystem, number];

// Projectileは配列形式: [id, sphere, ownerId]
export type Projectile = [string, Sphere, string];

// DroppedSatelliteは配列形式: [position, radius, color]
export type DroppedSatellite = [Position, number, string];

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

// 配列形式データ用の型定義
export interface ConvertedPosition {
  x: number;
  y: number;
}

export interface ConvertedSphere {
  p: ConvertedPosition;
  r: number;
  c: string; // color
  v?: ConvertedPosition;
  a?: ConvertedPosition;
}

export interface ConvertedCelestialSystem {
  c: ConvertedSphere;
  col: string;
  a: boolean;
  n: ConvertedSphere[];
}

export interface ConvertedPlayer {
  id: string;
  nm: string;
  cel: ConvertedCelestialSystem;
  sc: number;
}

export interface ConvertedProjectile {
  id: string;
  sph: ConvertedSphere;
  oid: string;
}

export interface ConvertedDroppedSatellite {
  p: ConvertedPosition;
  r: number;
  c: string; // color
}

// 配列形式データ用のヘルパー関数
export const getPosition = (pos: Position): ConvertedPosition => {
  if (!pos || !Array.isArray(pos) || pos.length < 2) {
    return { x: 0, y: 0 };
  }
  return { x: pos[0], y: pos[1] };
};

export const getSphere = (sphere: Sphere): ConvertedSphere => {
  if (!sphere || !Array.isArray(sphere) || sphere.length < 3) {
    return { p: { x: 0, y: 0 }, r: 0, c: '#000000' };
  }
  return {
    p: getPosition(sphere[0]),
    r: sphere[1],
    c: sphere[2],
    v: sphere[3] ? getPosition(sphere[3]) : undefined,
    a: sphere[4] ? getPosition(sphere[4]) : undefined
  };
};

export const getCelestialSystem = (cel: CelestialSystem): ConvertedCelestialSystem => {
  if (!cel || !Array.isArray(cel) || cel.length < 4) {
    return { c: { p: { x: 0, y: 0 }, r: 0, c: '#000000' }, col: '#000', a: false, n: [] };
  }
  return {
    c: getSphere(cel[0]),
    col: cel[1],
    a: cel[2],
    n: Array.isArray(cel[3]) ? cel[3].map(getSphere) : []
  };
};

export const getPlayer = (player: Player): ConvertedPlayer => {
  if (!player || !Array.isArray(player) || player.length < 4) {
    return { id: '', nm: '', cel: { c: { p: { x: 0, y: 0 }, r: 0, c: '#000000' }, col: '#000', a: false, n: [] }, sc: 0 };
  }
  return {
    id: player[0],
    nm: player[1],
    cel: getCelestialSystem(player[2]),
    sc: player[3]
  };
};
export const getDroppedSatellite = (ds: DroppedSatellite): ConvertedDroppedSatellite => ({
  p: getPosition(ds[0]),
  r: ds[1],
  c: ds[2]
});
export const getProjectile = (proj: Projectile): ConvertedProjectile => ({
  id: proj[0],
  sph: getSphere(proj[1]),
  oid: proj[2]
});

// デフォルトのゲーム設定（サーバーから受信するまでの暫定値）
export const DEFAULT_GAME_CONFIG: GameConfig = {
  fieldWidth: 600,
  fieldHeight: 600,
  sphereRadius: 7.5,
  cullingWidth: 1300,
  cullingHeight: 800,
};