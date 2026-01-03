import type { GameState, GameConfig } from "../../types";
import { getPlayer, getDroppedSatellite, getProjectile } from "../../types";
import { COLORS } from "../../constants/game";
import { calculateCameraOffset } from "../../utils/viewport";
import {
  drawFieldBoundary,
  drawSpatialGrid,
  drawServerCullingBounds,
  drawDroppedSatellites,
  drawProjectiles,
  drawCelestialSystem,
  drawUI,
} from "./drawingFunctions";

// メインの描画関数（カメラ追従付き）
export const drawGame = (
  ctx: CanvasRenderingContext2D,
  gameState: GameState,
  playerId: string,
  gameConfig: GameConfig,
  canvasSize: { width: number; height: number },
  showGrid: boolean,
  showCulling: boolean,
  showLeftUI: boolean
) => {
  // プレイヤーの位置を取得
  const currentPlayer = gameState.pls?.find((p) => p[0] === playerId);
  const playerData = currentPlayer ? getPlayer(currentPlayer) : null;
  const playerPosition = playerData?.cel?.c?.p;

  // カメラズーム設定を取得
  const zoomScale = gameConfig.cameraZoomScale || 1.0;

  // カメラの中心位置を計算
  const { x: cameraX, y: cameraY } = playerPosition
    ? calculateCameraOffset(playerPosition, gameConfig.cameraZoomScale || 1.0)
    : { x: 0, y: 0 };

  // キャンバスをクリア
  ctx.fillStyle = COLORS.BACKGROUND;
  ctx.fillRect(0, 0, canvasSize.width, canvasSize.height);

  // カメラ変換を適用
  ctx.save();

  // ゲームのズームを適用
  ctx.scale(
    gameConfig.cameraZoomScale || 1.0,
    gameConfig.cameraZoomScale || 1.0
  );

  // カメラの位置を適用
  ctx.translate(-cameraX, -cameraY);

  // フィールドの境界を描画
  drawFieldBoundary(ctx, gameConfig);

  // SpatialGridの線を描画
  if (showGrid && gameConfig.gridLines) {
    drawSpatialGrid(ctx, gameConfig.gridLines);
  }

  // サーバーカリング範囲を描画
  if (showCulling) {
    // 実際のサーバー側カリング範囲を計算（ズーム調整済み）
    const actualCullingConfig = {
      ...gameConfig,
      cullingWidth: gameConfig.cullingWidth / zoomScale,
      cullingHeight: gameConfig.cullingHeight / zoomScale,
    };
    drawServerCullingBounds(ctx, playerPosition, actualCullingConfig);
  }

  // 落ちた衛星を描画（カリング付き）
  drawDroppedSatellites(
    ctx,
    gameState.ds?.map(getDroppedSatellite) || [],
    cameraX,
    cameraY,
    canvasSize,
    zoomScale
  );

  // 射出物を描画（カリング付き）
  drawProjectiles(
    ctx,
    gameState.proj?.map(getProjectile) || [],
    gameConfig.sphereRadius,
    cameraX,
    cameraY,
    canvasSize,
    zoomScale
  );

  // プレイヤーを描画
  if (gameState.pls && gameState.pls.length > 0) {
    gameState.pls.forEach((player) => {
      const playerData = getPlayer(player);

      // 死んでいるプレイヤーは表示しない
      if (!playerData.cel?.a) {
        return;
      }

      drawCelestialSystem(ctx, playerData, playerData.id === playerId);
    });
  }

  // カメラ変換を元に戻す
  ctx.restore();

  // UI要素を描画（画面固定）
  const currentPlayerData = playerData || null;
  drawUI(ctx, currentPlayerData, showGrid, showCulling, showLeftUI);
};