import React, { useState, useEffect } from 'react';
import { Canvas } from '@react-three/fiber';
import { DeviceModel } from './DeviceModel';
import { StaticFallback } from './StaticFallback';
import { isWebGLAvailable, prefersReducedMotion } from './webgl';
import type { CanonicalSecurityState } from '../app/deviceSecurityState';

export interface DeviceSceneProps {
  status?: 'idle' | 'scanning' | 'completed';
  findingsCount?: number;
  canonicalState?: CanonicalSecurityState;
  className?: string;
}

export const DeviceScene: React.FC<DeviceSceneProps> = ({
  status = 'idle',
  findingsCount = 0,
  canonicalState,
  className = '',
}) => {
  const safeCount = Math.max(0, Number(findingsCount) || 0);
  const [canRender3D] = useState(() => isWebGLAvailable() && !prefersReducedMotion());
  const [isPaused, setIsPaused] = useState(false);

  useEffect(() => {
    if (!canRender3D) return;

    const handleVisibilityChange = () => {
      setIsPaused(document.hidden);
    };

    const handleWindowBlur = () => setIsPaused(true);
    const handleWindowFocus = () => setIsPaused(false);

    document.addEventListener('visibilitychange', handleVisibilityChange);
    window.addEventListener('blur', handleWindowBlur);
    window.addEventListener('focus', handleWindowFocus);

    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
      window.removeEventListener('blur', handleWindowBlur);
      window.removeEventListener('focus', handleWindowFocus);
    };
  }, [canRender3D]);

  if (!canRender3D) {
    return (
      <StaticFallback
        status={status}
        findingsCount={safeCount}
        canonicalState={canonicalState}
        className={className}
      />
    );
  }

  return (
    <div
      data-testid="device-scene-canvas"
      className={`w-full h-[220px] select-none ${className}`}
    >
      <Canvas
        camera={{ position: [0, 0, 4.5], fov: 45 }}
        dpr={[1, 2]} // PRD §5.5 bounded DPR <= 2
        gl={{ powerPreference: 'low-power', antialias: true }}
      >
        <ambientLight intensity={0.7} />
        <directionalLight position={[4, 5, 3]} intensity={1.2} />
        <DeviceModel
          status={status}
          findingsCount={safeCount}
          isPaused={isPaused}
        />
      </Canvas>
    </div>
  );
};

export default DeviceScene;
