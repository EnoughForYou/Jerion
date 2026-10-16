import React, { useRef, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';

export interface DeviceModelProps {
  status?: 'idle' | 'scanning' | 'completed';
  findingsCount?: number;
  isPaused?: boolean;
}

export const DeviceModel: React.FC<DeviceModelProps> = ({
  status = 'idle',
  findingsCount = 0,
  isPaused = false,
}) => {
  const ringRef = useRef<THREE.Mesh>(null);
  const chassisRef = useRef<THREE.Group>(null);

  const ringColor =
    status === 'scanning'
      ? '#0A84FF'
      : status === 'completed'
      ? findingsCount > 0
        ? '#FF6A3D'
        : '#30D158'
      : '#8E8E93';

  const boxGeo = useMemo(() => new THREE.BoxGeometry(1.4, 2.8, 0.15), []);

  useFrame((_, delta) => {
    if (isPaused) return;

    if (ringRef.current) {
      ringRef.current.rotation.z += delta * 0.8;
      if (status === 'scanning') {
        ringRef.current.position.y = Math.sin(Date.now() * 0.003) * 0.8;
      } else {
        ringRef.current.position.y = 0;
      }
    }

    if (chassisRef.current) {
      chassisRef.current.rotation.y = Math.sin(Date.now() * 0.001) * 0.15;
    }
  });

  return (
    <group>
      {/* Procedural Phone Chassis */}
      <group ref={chassisRef}>
        <mesh position={[0, 0, 0]}>
          <boxGeometry args={[1.4, 2.8, 0.15]} />
          <meshStandardMaterial
            color="#2C2C2E"
            roughness={0.4}
            metalness={0.6}
            wireframe={false}
          />
        </mesh>

        {/* Chassis Wireframe Outline Accent */}
        <lineSegments position={[0, 0, 0]}>
          <edgesGeometry args={[boxGeo]} />
          <lineBasicMaterial color="#0A84FF" opacity={0.3} transparent />
        </lineSegments>
      </group>

      {/* Procedural Scanner Ring */}
      <mesh ref={ringRef} position={[0, 0, 0]} rotation={[Math.PI / 2.5, 0, 0]}>
        <torusGeometry args={[1.8, 0.03, 16, 64]} />
        <meshBasicMaterial color={ringColor} transparent opacity={0.8} />
      </mesh>
    </group>
  );
};
