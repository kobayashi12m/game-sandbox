import {
  calculateCameraOffset,
  convertMouseToGameCoords,
} from "./viewport";

// マウス座標をワールド座標に変換する共通関数
export const convertMouseToWorldCoords = (
  event: MouseEvent,
  rect: DOMRect,
  playerPosition: { x: number; y: number } | undefined,
  gameZoomScale: number
): { worldX: number; worldY: number } | null => {
  if (!playerPosition) return null;

  const { x: cameraX, y: cameraY } = calculateCameraOffset(
    playerPosition,
    gameZoomScale
  );

  const gameCoords = convertMouseToGameCoords(event, rect);
  const { x: gameX, y: gameY } = gameCoords;

  const worldX = gameX / gameZoomScale + cameraX;
  const worldY = gameY / gameZoomScale + cameraY;

  return { worldX, worldY };
};