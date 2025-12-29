import { GAME_CONSTANTS } from '../constants/game';

export const calculateViewportScale = (): number => {
  const scaleX = window.innerWidth / GAME_CONSTANTS.BASE_WIDTH;
  const scaleY = window.innerHeight / GAME_CONSTANTS.BASE_HEIGHT;
  return Math.min(scaleX, scaleY);
};

export const calculateCameraOffset = (
  playerPosition: { x: number; y: number },
  zoomScale: number
) => ({
  x: playerPosition.x - GAME_CONSTANTS.BASE_WIDTH / 2 / zoomScale,
  y: playerPosition.y - GAME_CONSTANTS.BASE_HEIGHT / 2 / zoomScale,
});

export const convertMouseToGameCoords = (
  event: MouseEvent,
  rect: DOMRect
) => ({
  x: (event.clientX - rect.left) / (rect.width / GAME_CONSTANTS.BASE_WIDTH),
  y: (event.clientY - rect.top) / (rect.height / GAME_CONSTANTS.BASE_HEIGHT),
});