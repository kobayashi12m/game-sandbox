import React, { useRef, useEffect } from 'react';

interface Player {
  id: string;
  x: number;
  y: number;
  coreSize: number;
  guardSize: number;
}

interface InterpolatedPlayer extends Player {
  targetX: number;
  targetY: number;
  renderX: number;
  renderY: number;
}

interface GameCanvasProps {
  player: Player | null;
  otherPlayers: Player[];
  onPositionUpdate?: (x: number, y: number) => void;
}

const GameCanvas: React.FC<GameCanvasProps> = ({ player, otherPlayers, onPositionUpdate }) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [mousePos, setMousePos] = React.useState({ x: 400, y: 300 });
  const interpolatedPlayersRef = useRef<Map<string, InterpolatedPlayer>>(new Map());
  const animationFrameRef = useRef<number>();

  // マウス・タッチ移動ハンドラー
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const getPosition = (event: MouseEvent | TouchEvent) => {
      const rect = canvas.getBoundingClientRect();
      let x, y;
      
      if ('touches' in event) {
        // タッチイベント
        const touch = event.touches[0] || event.changedTouches[0];
        x = touch.clientX - rect.left;
        y = touch.clientY - rect.top;
      } else {
        // マウスイベント
        x = event.clientX - rect.left;
        y = event.clientY - rect.top;
      }
      
      return { x, y };
    };

    const handleMove = (event: MouseEvent | TouchEvent) => {
      const pos = getPosition(event);
      setMousePos(pos);
      
      // 位置更新をコールバック
      if (onPositionUpdate) {
        onPositionUpdate(pos.x, pos.y);
      }
    };

    // マウスとタッチの両方に対応
    canvas.addEventListener('mousemove', handleMove);
    canvas.addEventListener('touchmove', handleMove, { passive: true });

    return () => {
      canvas.removeEventListener('mousemove', handleMove);
      canvas.removeEventListener('touchmove', handleMove);
    };
  }, [onPositionUpdate]);

  // 他プレイヤーの位置更新時に補間データを設定
  useEffect(() => {
    const interpolatedPlayers = interpolatedPlayersRef.current;
    
    otherPlayers.forEach(player => {
      const existing = interpolatedPlayers.get(player.id);
      
      if (existing) {
        // 既存プレイヤーの目標位置を更新
        existing.targetX = player.x;
        existing.targetY = player.y;
        existing.x = player.x;
        existing.y = player.y;
        existing.coreSize = player.coreSize;
        existing.guardSize = player.guardSize;
      } else {
        // 新しいプレイヤーを追加
        interpolatedPlayers.set(player.id, {
          ...player,
          targetX: player.x,
          targetY: player.y,
          renderX: player.x,
          renderY: player.y,
        });
      }
    });

    // 存在しなくなったプレイヤーを削除
    const currentPlayerIds = new Set(otherPlayers.map(p => p.id));
    for (const [playerId] of interpolatedPlayers) {
      if (!currentPlayerIds.has(playerId)) {
        interpolatedPlayers.delete(playerId);
      }
    }
  }, [otherPlayers]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // Canvasサイズをレスポンシブに設定
    const isMobile = window.innerWidth < 768;
    canvas.width = isMobile ? Math.min(window.innerWidth - 20, 400) : 800;
    canvas.height = isMobile ? Math.min(window.innerHeight - 200, 300) : 600;

    // 位置補間関数
    const lerp = (start: number, end: number, factor: number) => {
      return start + (end - start) * factor;
    };

    // 描画関数
    const draw = () => {
      // 背景をクリア
      ctx.fillStyle = '#f0f0f0';
      ctx.fillRect(0, 0, canvas.width, canvas.height);

      // デバッグ情報を描画（PCのみ）
      if (!isMobile) {
        ctx.fillStyle = 'black';
        ctx.font = '14px Arial';
        ctx.textAlign = 'left';
        ctx.fillText(`他プレイヤー数: ${otherPlayers.length}`, 10, 20);
      }

      // 他のプレイヤーを補間位置で描画
      const interpolatedPlayers = interpolatedPlayersRef.current;
      interpolatedPlayers.forEach(otherPlayer => {
        // 位置を補間（スムーズな移動）
        const lerpFactor = 0.15; // 補間の強さ（0.1-0.2が適切）
        otherPlayer.renderX = lerp(otherPlayer.renderX, otherPlayer.targetX, lerpFactor);
        otherPlayer.renderY = lerp(otherPlayer.renderY, otherPlayer.targetY, lerpFactor);

        // ガード（外側の円）を描画
        ctx.beginPath();
        ctx.arc(otherPlayer.renderX, otherPlayer.renderY, otherPlayer.guardSize, 0, Math.PI * 2);
        ctx.fillStyle = 'rgba(100, 100, 100, 0.3)';
        ctx.fill();
        ctx.strokeStyle = 'rgba(100, 100, 100, 0.8)';
        ctx.lineWidth = 2;
        ctx.stroke();

        // コア（内側の円）を描画
        ctx.beginPath();
        ctx.arc(otherPlayer.renderX, otherPlayer.renderY, otherPlayer.coreSize, 0, Math.PI * 2);
        ctx.fillStyle = 'rgba(150, 150, 150, 0.8)';
        ctx.fill();
        ctx.strokeStyle = 'rgba(100, 100, 100, 1)';
        ctx.lineWidth = 2;
        ctx.stroke();

        // プレイヤーIDを表示
        ctx.fillStyle = 'black';
        ctx.font = '12px Arial';
        ctx.textAlign = 'center';
        ctx.fillText(otherPlayer.id, otherPlayer.renderX, otherPlayer.renderY - otherPlayer.guardSize - 10);
      });

      // 自分のプレイヤーを描画（最前面）
      if (player) {
        // マウス位置にプレイヤーを描画
        const drawX = mousePos.x;
        const drawY = mousePos.y;

        // ガード（外側の円）を描画
        ctx.beginPath();
        ctx.arc(drawX, drawY, player.guardSize, 0, Math.PI * 2);
        ctx.fillStyle = 'rgba(0, 100, 255, 0.3)';
        ctx.fill();
        ctx.strokeStyle = 'rgba(0, 100, 255, 0.8)';
        ctx.lineWidth = 2;
        ctx.stroke();

        // コア（内側の円）を描画
        ctx.beginPath();
        ctx.arc(drawX, drawY, player.coreSize, 0, Math.PI * 2);
        ctx.fillStyle = 'rgba(255, 0, 0, 0.8)';
        ctx.fill();
        ctx.strokeStyle = 'rgba(200, 0, 0, 1)';
        ctx.lineWidth = 2;
        ctx.stroke();

        // プレイヤーIDを表示
        ctx.fillStyle = 'black';
        ctx.font = '12px Arial';
        ctx.textAlign = 'center';
        ctx.fillText(player.id, drawX, drawY - player.guardSize - 10);
      }
    };

    // アニメーションループ（補間のために必要）
    const animate = () => {
      draw();
      animationFrameRef.current = requestAnimationFrame(animate);
    };

    // アニメーション開始
    animationFrameRef.current = requestAnimationFrame(animate);

    return () => {
      if (animationFrameRef.current) {
        cancelAnimationFrame(animationFrameRef.current);
      }
    };
  }, [player, mousePos]);

  return (
    <canvas
      ref={canvasRef}
      style={{
        border: '1px solid #ccc',
        display: 'block',
        margin: '20px auto',
        borderRadius: '8px',
        boxShadow: '0 2px 8px rgba(0, 0, 0, 0.1)'
      }}
    />
  );
};

export default GameCanvas;