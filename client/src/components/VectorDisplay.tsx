import type { Position, NPCDebugStats } from "../types";

interface VectorDisplayProps {
  velocity: Position | undefined;
  acceleration: Position | undefined;
  maxSpeed: number;
  npcDebug?: NPCDebugStats | null;
}

// ベクトルの大きさを計算する共通関数
const calculateMagnitude = (x: number, y: number): number =>
  Math.sqrt(x * x + y * y);

export const VectorDisplay: React.FC<VectorDisplayProps> = ({
  velocity,
  acceleration,
  maxSpeed,
  npcDebug,
}) => {
  // 配列形式の Position [x, y] にアクセス
  const vx = velocity?.[0] ?? 0;
  const vy = velocity?.[1] ?? 0;
  const ax = acceleration?.[0] ?? 0;
  const ay = acceleration?.[1] ?? 0;

  // ベクトルの大きさを計算
  const velocityMagnitude = calculateMagnitude(vx, vy);
  const accelerationMagnitude = calculateMagnitude(ax, ay);

  const npcVelocityMagnitude = npcDebug
    ? calculateMagnitude(npcDebug.velX, npcDebug.velY)
    : 0;
  const npcAccelerationMagnitude = npcDebug
    ? calculateMagnitude(npcDebug.accelX, npcDebug.accelY)
    : 0;

  return (
    <div className="vector-display">
      <div className="vector-info">
        <h3>プレイヤー情報</h3>
        <div className="vector-item">
          <span className="vector-label">速度:</span>
          <div className="vector-values">
            <span>X: {vx.toFixed(1)}</span>
            <span>Y: {vy.toFixed(1)}</span>
            <span>
              大きさ: {velocityMagnitude.toFixed(1)} / {maxSpeed}
            </span>
          </div>
        </div>
        <div className="vector-item">
          <span className="vector-label">加速度:</span>
          <div className="vector-values">
            <span>X: {ax.toFixed(1)}</span>
            <span>Y: {ay.toFixed(1)}</span>
            <span>大きさ: {accelerationMagnitude.toFixed(1)}</span>
          </div>
        </div>
      </div>

      {npcDebug && (
        <div
          className="vector-info"
          style={{
            marginTop: "10px",
            borderTop: "1px solid #ccc",
            paddingTop: "10px",
          }}
        >
          <h3>NPC情報 ({npcDebug.name})</h3>
          <div className="vector-item">
            <span className="vector-label">速度:</span>
            <div className="vector-values">
              <span>X: {npcDebug.velX.toFixed(1)}</span>
              <span>Y: {npcDebug.velY.toFixed(1)}</span>
              <span>
                大きさ: {npcVelocityMagnitude.toFixed(1)} / {npcDebug.maxSpeed}
              </span>
            </div>
          </div>
          <div className="vector-item">
            <span className="vector-label">加速度:</span>
            <div className="vector-values">
              <span>X: {npcDebug.accelX.toFixed(1)}</span>
              <span>Y: {npcDebug.accelY.toFixed(1)}</span>
              <span>大きさ: {npcAccelerationMagnitude.toFixed(1)}</span>
            </div>
          </div>
          <div className="vector-item">
            <span className="vector-label">衛星数:</span>
            <span>{npcDebug.satellites}</span>
          </div>
        </div>
      )}
    </div>
  );
};
