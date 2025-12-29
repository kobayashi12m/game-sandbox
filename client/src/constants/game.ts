// ゲーム共通定数
export const GAME_CONSTANTS = {
  BASE_WIDTH: 1920,
  BASE_HEIGHT: 1080,
  SEND_INTERVAL: 33, // 30fps
  MIN_DISTANCE: 150,  // マウスデッドゾーン
  MAX_DISTANCE: 400,  // 最大速度距離
  CULLING_MARGIN: 300,
  PROJECTILE_TRAIL_LENGTH: 20,
} as const;

export const COLORS = {
  BACKGROUND: '#0a0a0a',
  GRID: 'rgba(0, 200, 255, 0.8)',
  CULLING: '#ff0000',
  GOLD: '#ffd700',
  WHITE: '#fff',
  PRIMARY: '#4ECDC4',
} as const;

export const PLAYER_CONFIG = {
  ROOM_ID: 'default',
  PLAYER_NAME: '',
} as const;