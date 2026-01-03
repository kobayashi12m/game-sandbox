import type { GameConfig, GridLine, ConvertedPlayer, ConvertedDroppedSatellite, ConvertedProjectile } from "../../types";
import { COLORS, GAME_CONSTANTS } from "../../constants/game";

// フィールドの境界を描画
export const drawFieldBoundary = (
  ctx: CanvasRenderingContext2D,
  gameConfig: GameConfig
) => {
  ctx.strokeStyle = "#333";
  ctx.lineWidth = 3;
  ctx.setLineDash([10, 5]);
  ctx.strokeRect(0, 0, gameConfig.fieldWidth, gameConfig.fieldHeight);
  ctx.setLineDash([]);
};

// SpatialGridの線を描画
export const drawSpatialGrid = (
  ctx: CanvasRenderingContext2D,
  gridLines: GridLine[]
) => {
  ctx.strokeStyle = COLORS.GRID;
  ctx.lineWidth = 1.5;
  ctx.setLineDash([3, 3]);

  gridLines.forEach((line) => {
    ctx.beginPath();
    ctx.moveTo(line.startX, line.startY);
    ctx.lineTo(line.endX, line.endY);
    ctx.stroke();
  });

  ctx.setLineDash([]);
};

// サーバーカリング範囲を描画（デバッグ用）
export const drawServerCullingBounds = (
  ctx: CanvasRenderingContext2D,
  playerPosition: { x: number; y: number } | undefined,
  gameConfig: GameConfig
) => {
  if (!playerPosition) return;

  // サーバーから受信したカリング範囲を使用
  const serverViewWidth = gameConfig.cullingWidth;
  const serverViewHeight = gameConfig.cullingHeight;

  const minX = playerPosition.x - serverViewWidth / 2;
  const maxX = playerPosition.x + serverViewWidth / 2;
  const minY = playerPosition.y - serverViewHeight / 2;
  const maxY = playerPosition.y + serverViewHeight / 2;

  // カリング境界を赤い点線で描画
  ctx.strokeStyle = COLORS.CULLING;
  ctx.lineWidth = 2;
  ctx.setLineDash([5, 5]);
  ctx.strokeRect(minX, minY, maxX - minX, maxY - minY);
  ctx.setLineDash([]);
};

// 落ちた衛星の描画（カリング付き）
export const drawDroppedSatellites = (
  ctx: CanvasRenderingContext2D,
  droppedSatellites: ConvertedDroppedSatellite[] | undefined,
  cameraX: number,
  cameraY: number,
  canvasSize: { width: number; height: number },
  zoomScale: number
) => {
  if (!droppedSatellites || droppedSatellites.length === 0) return;

  // カリング境界（ズーム考慮）
  const margin = 100;
  const minX = cameraX - margin / zoomScale;
  const maxX = cameraX + (canvasSize.width + margin) / zoomScale;
  const minY = cameraY - margin / zoomScale;
  const maxY = cameraY + (canvasSize.height + margin) / zoomScale;

  droppedSatellites.forEach((satellite) => {
    // 画面範囲内の落ちた衛星のみ描画
    if (
      satellite.p.x >= minX &&
      satellite.p.x <= maxX &&
      satellite.p.y >= minY &&
      satellite.p.y <= maxY
    ) {
      // 衛星の色を使用
      ctx.fillStyle = satellite.c;
      ctx.shadowBlur = 8;
      ctx.shadowColor = satellite.c;

      ctx.beginPath();
      ctx.arc(satellite.p.x, satellite.p.y, satellite.r, 0, 2 * Math.PI);
      ctx.fill();

      // 内側に小さな光る点を追加
      ctx.fillStyle = COLORS.WHITE;
      ctx.shadowBlur = 4;
      ctx.shadowColor = COLORS.WHITE;
      ctx.beginPath();
      ctx.arc(satellite.p.x, satellite.p.y, satellite.r * 0.3, 0, 2 * Math.PI);
      ctx.fill();
    }
  });

  ctx.shadowBlur = 0;
};

// 射出物の描画（カリング付き）
export const drawProjectiles = (
  ctx: CanvasRenderingContext2D,
  projectiles: ConvertedProjectile[] | undefined,
  radius: number,
  cameraX: number,
  cameraY: number,
  canvasSize: { width: number; height: number },
  zoomScale: number
) => {
  if (!projectiles || projectiles.length === 0) return;

  // カリング境界（ズーム考慮）
  const margin = 100;
  const minX = cameraX - margin / zoomScale;
  const maxX = cameraX + (canvasSize.width + margin) / zoomScale;
  const minY = cameraY - margin / zoomScale;
  const maxY = cameraY + (canvasSize.height + margin) / zoomScale;

  projectiles.forEach((projectile) => {
    const pos = projectile.sph.p;

    // 画面範囲内の射出物のみ描画
    if (pos.x >= minX && pos.x <= maxX && pos.y >= minY && pos.y <= maxY) {
      // 射出物の色を使用
      const projectileColor = projectile.sph.c;
      ctx.fillStyle = projectileColor;
      ctx.shadowBlur = 15;
      ctx.shadowColor = projectileColor;

      ctx.beginPath();
      ctx.arc(pos.x, pos.y, radius, 0, 2 * Math.PI);
      ctx.fill();

      // 軌跡の描画（速度ベクトルの逆方向に短い線）
      if (projectile.sph.v) {
        const vel = projectile.sph.v;
        const speed = Math.sqrt(vel.x * vel.x + vel.y * vel.y);
        if (speed > 0) {
          const trailLength = GAME_CONSTANTS.PROJECTILE_TRAIL_LENGTH;
          const dirX = -vel.x / speed;
          const dirY = -vel.y / speed;

          ctx.strokeStyle = projectileColor;
          ctx.lineWidth = 3;
          ctx.globalAlpha = 0.5;
          ctx.beginPath();
          ctx.moveTo(pos.x, pos.y);
          ctx.lineTo(pos.x + dirX * trailLength, pos.y + dirY * trailLength);
          ctx.stroke();
          ctx.globalAlpha = 1;
        }
      }

      ctx.shadowBlur = 0;
    }
  });
};

// 球体構造の描画
export const drawCelestialSystem = (
  ctx: CanvasRenderingContext2D,
  player: ConvertedPlayer,
  isCurrentPlayer: boolean
) => {
  const celestialSystem = player.cel;

  // 描画状態を保存
  ctx.save();

  // 無敵時の透明度を設定
  const alpha = player.inv ? 0.5 : 1.0;
  const lineAlpha = player.inv ? 0.3 : 0.6;

  // 軌道を先に描画 - 核から各衛星への放射状線
  ctx.lineWidth = 2;
  ctx.globalAlpha = lineAlpha;

  // コアから各ノードへの線を描画
  if (celestialSystem.n && celestialSystem.n.length > 0) {
    celestialSystem.n.forEach((node) => {
      // 各衛星の色で線を描画
      ctx.strokeStyle = node.c;
      ctx.beginPath();
      ctx.moveTo(celestialSystem.c.p.x, celestialSystem.c.p.y);
      ctx.lineTo(node.p.x, node.p.y);
      ctx.stroke();
    });
  }

  // 球体描画用の透明度に設定
  ctx.globalAlpha = alpha;

  // コア（中心球）を描画
  ctx.fillStyle = celestialSystem.c.c;
  drawCoreHead(
    ctx,
    celestialSystem.c.p,
    celestialSystem.c.r,
    isCurrentPlayer
  );

  // ノード（周辺球）を描画
  if (celestialSystem.n && celestialSystem.n.length > 0) {
    celestialSystem.n.forEach((node) => {
      // 透明度を維持
      ctx.globalAlpha = alpha;
      ctx.fillStyle = node.c;
      ctx.beginPath();
      ctx.arc(node.p.x, node.p.y, node.r, 0, 2 * Math.PI);
      ctx.fill();
    });
  }

  // 描画状態を復元
  ctx.restore();
};

// コアの頭部を描画
const drawCoreHead = (
  ctx: CanvasRenderingContext2D,
  position: { x: number; y: number },
  radius: number,
  isCurrentPlayer: boolean
) => {
  // 現在の透明度を保存
  const savedAlpha = ctx.globalAlpha;

  // 頭部の円
  ctx.beginPath();
  ctx.arc(position.x, position.y, radius, 0, 2 * Math.PI);
  ctx.fill();

  // 自分の球体構造にはアウトラインを追加
  if (isCurrentPlayer) {
    ctx.save(); // 状態を保存
    ctx.globalAlpha = savedAlpha; // 透明度を維持
    ctx.strokeStyle = COLORS.GOLD;
    ctx.lineWidth = 3;
    ctx.shadowBlur = 15;
    ctx.shadowColor = COLORS.GOLD;
    ctx.beginPath();
    ctx.arc(position.x, position.y, radius + 2, 0, 2 * Math.PI);
    ctx.stroke();
    ctx.restore(); // 状態を復元
  }

  // 目を描画
  ctx.globalAlpha = savedAlpha; // 透明度を維持
  ctx.fillStyle = "#000"; // 黒い目
  const eyeRadius = radius * 0.25; // 大きな目
  const eyeOffset = radius * 0.4; // 離れた位置

  // 左目
  ctx.beginPath();
  ctx.arc(
    position.x - eyeOffset,
    position.y - eyeOffset, // 上の位置
    eyeRadius,
    0,
    2 * Math.PI
  );
  ctx.fill();

  // 右目
  ctx.beginPath();
  ctx.arc(
    position.x + eyeOffset,
    position.y - eyeOffset, // 上の位置
    eyeRadius,
    0,
    2 * Math.PI
  );
  ctx.fill();

  // 透明度を復元
  ctx.globalAlpha = savedAlpha;
};

// UI要素の描画（画面固定）
export const drawUI = (
  ctx: CanvasRenderingContext2D,
  currentPlayer: ConvertedPlayer | null,
  showGrid: boolean,
  showCulling: boolean,
  showLeftUI: boolean
) => {
  if (!currentPlayer) return;

  // 左側UIの表示制御
  if (!showLeftUI) return;

  // 表示設定とヘルプ
  const helpLines = [];
  if (showGrid) helpLines.push("Grid: ON (G to toggle)");
  if (showCulling) helpLines.push("Culling: ON (C to toggle)");

  if (helpLines.length > 0 || (!showGrid && !showCulling)) {
    const boxHeight = Math.max(60, 20 + helpLines.length * 20);
    ctx.fillStyle = "rgba(0, 0, 0, 0.8)";
    ctx.fillRect(10, 100, 280, boxHeight);

    ctx.fillStyle = COLORS.WHITE;
    ctx.font = "14px Arial";
    ctx.textAlign = "left";

    let yPos = 120;
    if (helpLines.length === 0) {
      ctx.fillText("Press G: Grid lines, C: Culling bounds", 20, yPos);
    } else {
      helpLines.forEach((line) => {
        ctx.fillText(line, 20, yPos);
        yPos += 20;
      });
      ctx.fillText("Press G/C to toggle", 20, yPos);
    }
  }
};