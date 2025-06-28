<template>
  <section class="feature-panel web-gateway">
    <div class="feature-content-row">
      <div class="feature-content feature-content-overlay">
        <h2 class="feature-title">Simple Web Gateway Access</h2>
        <p class="feature-description">
          Expose any service to the public web with a simple Kubernetes resource. No complex Ingress controllers. Domain and prefix-based routing is built-in.
        </p>
        <div class="feature-benefits">
          <v-chip class="ma-1" color="white" variant="outlined" size="small">Simple configuration</v-chip>
          <v-chip class="ma-1" color="white" variant="outlined" size="small">Built-in routing</v-chip>
          <v-chip class="ma-1" color="white" variant="outlined" size="small">No ingress controllers</v-chip>
        </div>
      </div>
      <div class="feature-animation-with-glow">
        <div class="feature-animation-glow"></div>
        <div class="feature-animation">
          <canvas ref="threeCanvas" class="three-canvas" @mousemove="onCanvasMouseMove" @mouseleave="onCanvasMouseLeave"></canvas>
          <div v-if="hoveredNode" class="node-label" :style="{ left: hoveredNode.pos.x + 'px', top: hoveredNode.pos.y + 'px' }">
            {{ hoveredNode.type }}
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, shallowRef } from 'vue'
import * as THREE from 'three'

const threeCanvas = ref<HTMLCanvasElement | null>(null)
let renderer: THREE.WebGLRenderer | null = null
let animationId: number | null = null
let scene: THREE.Scene | null = null
let camera: THREE.PerspectiveCamera | null = null

// Network topology constants
const CLIENT_COUNT = 5
const SERVER_COUNT = 4
const GATEWAY_NODES = 3
const GATEWAY_RADIUS = 0.8

// Sophisticated color palette
const PRIMARY_COLOR = 0x60a5fa
const ACCENT_COLOR = 0x8ecfff
const SUCCESS_COLOR = 0x34d399

// Network nodes and connections
const clients: { mesh: THREE.Mesh, glow: THREE.Mesh, basePos: THREE.Vector3, phase: number }[] = []
const servers: { mesh: THREE.Mesh, glow: THREE.Mesh, basePos: THREE.Vector3, phase: number, type: 'public' | 'private' }[] = []
const gatewayNodes: { mesh: THREE.Mesh, glow: THREE.Mesh, basePos: THREE.Vector3, phase: number }[] = []
const connections: { line: THREE.Line, nodes: [THREE.Mesh, THREE.Mesh] }[] = []

// Add a ref to track gateway pulse state
const gatewayPulse = Array(GATEWAY_NODES).fill(0)

// Add hover state for node labels
const hoveredNode = shallowRef<{ mesh: THREE.Mesh, type: string, pos: { x: number, y: number } } | null>(null)

// Add at the top:
let pulses: { mesh: THREE.Mesh, path: THREE.Vector3[], t: number, stage: 0 | 1, active: boolean }[] = [];
let pulseCooldown = 0;

// Helper to check if a connection exists between two nodes
function hasConnection(a: THREE.Mesh, b: THREE.Mesh): boolean {
  return connections.some(conn => {
    const posArr = (conn.line.geometry as THREE.BufferGeometry).attributes.position.array as Float32Array;
    const p1 = new THREE.Vector3(posArr[0], posArr[1], posArr[2]);
    const p2 = new THREE.Vector3(posArr[3], posArr[4], posArr[5]);
    // Compare both directions
    return (
      (a.position.distanceTo(p1) < 1e-6 && b.position.distanceTo(p2) < 1e-6) ||
      (b.position.distanceTo(p1) < 1e-6 && a.position.distanceTo(p2) < 1e-6)
    );
  });
}

// Add mouse event handlers for hover detection
function onCanvasMouseMove(event: MouseEvent) {
  if (!threeCanvas.value || !renderer || !camera) return
  const rect = threeCanvas.value.getBoundingClientRect()
  const mouse = new THREE.Vector2(
    ((event.clientX - rect.left) / rect.width) * 2 - 1,
    -((event.clientY - rect.top) / rect.height) * 2 + 1
  )
  // Raycast against all nodes
  const allMeshes = [
    ...clients.map(c => ({ mesh: c.mesh, type: 'Client' })),
    ...gatewayNodes.map(g => ({ mesh: g.mesh, type: 'Gateway' })),
    ...servers.map(s => ({ mesh: s.mesh, type: s.type === 'public' ? 'Public Server' : 'Private Server' }))
  ]
  const raycaster = new THREE.Raycaster()
  raycaster.setFromCamera(mouse, camera)
  const intersects = raycaster.intersectObjects(allMeshes.map(n => n.mesh))
  if (intersects.length > 0) {
    const found = allMeshes.find(n => n.mesh === intersects[0].object)
    if (found) {
      // Project 3D position to 2D overlay
      const pos = found.mesh.position.clone().project(camera)
      const x = ((pos.x + 1) / 2) * rect.width
      const y = ((-pos.y + 1) / 2) * rect.height
      hoveredNode.value = { mesh: found.mesh, type: found.type, pos: { x, y } }
      return
    }
  }
  hoveredNode.value = null
}
function onCanvasMouseLeave() {
  hoveredNode.value = null
}

onMounted(() => {
  if (!threeCanvas.value) return

  // Initialize Three.js with sophisticated settings
  renderer = new THREE.WebGLRenderer({
    canvas: threeCanvas.value,
    alpha: true,
    antialias: true,
    powerPreference: "high-performance"
  })
  renderer.setSize(threeCanvas.value.clientWidth, threeCanvas.value.clientHeight, false)
  renderer.setClearColor(0x000000, 0) // Transparent background like Mycelium

  scene = new THREE.Scene()
  camera = new THREE.PerspectiveCamera(
    60,
    threeCanvas.value.clientWidth / threeCanvas.value.clientHeight,
    0.1,
    1000
  )
  camera.position.z = 6

  // Create client nodes (left side) - small, elegant points
  const clientGeometry = new THREE.SphereGeometry(0.06, 16, 16)
  for (let i = 0; i < CLIENT_COUNT; i++) {
    // Place clients in a leftward arc, closer to center
    const angle = (i / CLIENT_COUNT) * Math.PI * 1.2 - Math.PI * 1.1 // More leftward
    const radius = 1.7 + Math.random() * 0.2 // Reduced from 2.2 + Math.random() * 0.4
    const basePos = new THREE.Vector3(
      -radius * Math.cos(angle),
      radius * Math.sin(angle),
      (Math.random() - 0.5) * 0.3
    )

    const mesh = new THREE.Mesh(clientGeometry, new THREE.MeshBasicMaterial({
      color: PRIMARY_COLOR,
      transparent: true,
      opacity: 0.9
    }))
    mesh.position.copy(basePos)
    scene.add(mesh)

    // Subtle glow
    const glowGeometry = new THREE.SphereGeometry(0.14, 16, 16)
    const glow = new THREE.Mesh(glowGeometry, new THREE.MeshBasicMaterial({
      color: PRIMARY_COLOR,
      transparent: true,
      opacity: 0.15
    }))
    glow.position.copy(basePos)
    scene.add(glow)

    clients.push({
      mesh,
      glow,
      basePos: basePos.clone(),
      phase: Math.random() * Math.PI * 2
    })
  }

  // Create gateway nodes (center) - sophisticated cluster
  const gatewayGeometry = new THREE.SphereGeometry(0.09, 18, 18)
  for (let i = 0; i < GATEWAY_NODES; i++) {
    const angle = (i / GATEWAY_NODES) * Math.PI * 2
    const radius = GATEWAY_RADIUS * (0.7 + Math.random() * 0.6)
    const basePos = new THREE.Vector3(
      radius * Math.cos(angle),
      radius * Math.sin(angle),
      (Math.random() - 0.5) * 0.2
    )

    const mesh = new THREE.Mesh(gatewayGeometry, new THREE.MeshBasicMaterial({
      color: ACCENT_COLOR,
      transparent: true,
      opacity: 0.95
    }))
    mesh.position.copy(basePos)
    scene.add(mesh)

    // Enhanced glow for gateway nodes (sphere glow)
    const glowGeometry = new THREE.SphereGeometry(0.15, 18, 18)
    const glow = new THREE.Mesh(glowGeometry, new THREE.MeshBasicMaterial({
      color: ACCENT_COLOR,
      transparent: true,
      opacity: 0.2
    }))
    glow.position.copy(basePos)
    scene.add(glow)

    gatewayNodes.push({
      mesh,
      glow,
      basePos: basePos.clone(),
      phase: Math.random() * Math.PI * 2
    })
  }

  // Create server nodes (right side) - elegant endpoints
  const serverGeometry = new THREE.BoxGeometry(0.18, 0.11, 0.09)
  // Randomly assign half as private, half as public
  const serverTypes: ('public' | 'private')[] = Array(SERVER_COUNT).fill('public').map((v, i) => i < Math.floor(SERVER_COUNT / 2) ? 'private' : 'public')
  // Shuffle for randomness
  for (let i = serverTypes.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [serverTypes[i], serverTypes[j]] = [serverTypes[j], serverTypes[i]];
  }
  for (let i = 0; i < SERVER_COUNT; i++) {
    // Place servers in a rightward arc, further out
    const angle = (i / SERVER_COUNT) * Math.PI * 1.2 - Math.PI * 0.1 // More rightward
    const radius = 2.6 + Math.random() * 0.3 // Increased from 2.4 + Math.random() * 0.3
    const basePos = new THREE.Vector3(
      radius * Math.cos(angle),
      radius * Math.sin(angle),
      (Math.random() - 0.5) * 0.2
    )
    const type = serverTypes[i]
    const mesh = new THREE.Mesh(serverGeometry, new THREE.MeshBasicMaterial({
      color: SUCCESS_COLOR,
      transparent: true,
      opacity: 0.85
    }))
    mesh.position.copy(basePos)
    if (scene) scene.add(mesh)
    // Glow: thicker and brighter for private
    const glowGeometry = new THREE.BoxGeometry(type === 'private' ? 0.28 : 0.21, type === 'private' ? 0.17 : 0.13, type === 'private' ? 0.13 : 0.09)
    const glow = new THREE.Mesh(glowGeometry, new THREE.MeshBasicMaterial({
      color: SUCCESS_COLOR,
      transparent: true,
      opacity: type === 'private' ? 0.28 : 0.12
    }))
    glow.position.copy(basePos)
    if (scene) scene.add(glow)
    servers.push({
      mesh,
      glow,
      basePos: basePos.clone(),
      phase: Math.random() * Math.PI * 2,
      type
    })
  }

  // Create sophisticated network connections (like Mycelium)
  // Only allow connections to private servers via gateway
  const connectionDistance = 1.8
  // Connect clients <-> gateway, gateway <-> servers (all), clients <-> public servers (direct)
  // 1. Clients <-> Gateway
  clients.forEach(client => {
    gatewayNodes.forEach(gateway => {
      const lineMaterial = new THREE.LineBasicMaterial({
        color: PRIMARY_COLOR,
        transparent: true,
        opacity: 0.15
      })
      const p1 = client.mesh.position.clone()
      const p2 = gateway.mesh.position.clone()
      const lineGeom = new THREE.BufferGeometry().setFromPoints([p1, p2])
      const line = new THREE.Line(lineGeom, lineMaterial)
      if (scene) scene.add(line)
      connections.push({
        line,
        nodes: [client.mesh, gateway.mesh]
      })
    })
  })
  // 2. Gateway <-> All servers
  gatewayNodes.forEach(gateway => {
    servers.forEach(server => {
      const lineMaterial = new THREE.LineBasicMaterial({
        color: PRIMARY_COLOR,
        transparent: true,
        opacity: 0.15
      })
      const p1b = gateway.mesh.position.clone()
      const p2b = server.mesh.position.clone()
      const lineGeomb = new THREE.BufferGeometry().setFromPoints([p1b, p2b])
      const lineb = new THREE.Line(lineGeomb, lineMaterial)
      if (scene) scene.add(lineb)
      connections.push({
        line: lineb,
        nodes: [gateway.mesh, server.mesh]
      })
    })
  })
  // 3. Clients <-> Public servers (direct)
  clients.forEach(client => {
    servers.forEach(server => {
      if (server.type === 'public') {
        const lineMaterial = new THREE.LineBasicMaterial({
          color: PRIMARY_COLOR,
          transparent: true,
          opacity: 0.15
        })
        const p1c = client.mesh.position.clone()
        const p2c = server.mesh.position.clone()
        const lineGeomc = new THREE.BufferGeometry().setFromPoints([p1c, p2c])
        const linec = new THREE.Line(lineGeomc, lineMaterial)
        if (scene) scene.add(linec)
        connections.push({
          line: linec,
          nodes: [client.mesh, server.mesh]
        })
      }
    })
  })

  // Sophisticated animation function (Mycelium-style)
  function animate(time = 0) {
    animationId = requestAnimationFrame(animate)

    // Organic node movements (like Mycelium)
    const t = time * 0.001

    // Animate client nodes with subtle organic motion
    clients.forEach((client, idx) => {
      client.mesh.position.x = client.basePos.x + Math.sin(t * 0.6 + client.phase + idx) * 0.12
      client.mesh.position.y = client.basePos.y + Math.cos(t * 0.8 + client.phase + idx * 1.1) * 0.12
      client.mesh.position.z = client.basePos.z + Math.sin(t * 0.4 + client.phase + idx * 0.7) * 0.08
      client.glow.position.copy(client.mesh.position)

      // Subtle pulsing
      const pulse = 1 + Math.sin(t * 2 + idx) * 0.06
      client.mesh.scale.set(pulse, pulse, pulse)
      client.glow.scale.set(pulse * 1.1, pulse * 1.1, pulse * 1.1)
    })

    // Animate gateway nodes with more pronounced movement and glow (toned down)
    gatewayNodes.forEach((gateway, idx) => {
      gateway.mesh.position.x = gateway.basePos.x + Math.sin(t * 0.7 + gateway.phase + idx) * 0.09
      gateway.mesh.position.y = gateway.basePos.y + Math.cos(t * 0.9 + gateway.phase + idx * 1.3) * 0.09
      gateway.mesh.position.z = gateway.basePos.z + Math.sin(t * 0.5 + gateway.phase + idx * 0.8) * 0.06
      gateway.glow.position.copy(gateway.mesh.position)
      // Subtle pulsing for gateway, plus pulse effect
      let pulse = 1 + Math.sin(t * 2.5 + idx) * 0.04
      if (gatewayPulse[idx] > 0) {
        pulse += 0.13 * gatewayPulse[idx]
        gatewayPulse[idx] -= 0.06
        if (gatewayPulse[idx] < 0) gatewayPulse[idx] = 0
      }
      gateway.mesh.scale.set(pulse, pulse, pulse)
      gateway.glow.scale.set(pulse * 1.18, pulse * 1.18, pulse * 1.18)
      let glowMat = gateway.glow.material
      if (Array.isArray(glowMat)) glowMat = glowMat[0]
      if (glowMat && 'opacity' in glowMat) {
        glowMat.opacity = 0.22 + 0.10 * Math.abs(Math.sin(t * 2.5 + idx)) + 0.13 * (gatewayPulse[idx] > 0 ? gatewayPulse[idx] : 0)
      }
    })

    // Animate server nodes (toned down pulsing)
    servers.forEach((server, idx) => {
      server.mesh.position.x = server.basePos.x + Math.sin(t * 0.5 + server.phase + idx) * 0.06
      server.mesh.position.y = server.basePos.y + Math.cos(t * 0.7 + server.phase + idx * 1.2) * 0.06
      server.mesh.position.z = server.basePos.z + Math.sin(t * 0.3 + server.phase + idx * 0.6) * 0.04
      server.glow.position.copy(server.mesh.position)
      // Gentle pulsing
      const pulse = 1 + Math.sin(t * 1.8 + idx) * 0.025
      server.mesh.scale.set(pulse, pulse, pulse)
      server.glow.scale.set(pulse * 1.05, pulse * 1.05, pulse * 1.05)
    })

    // --- SYNCHRONIZED UPDATE: Connection lines and packet spheres ---
    // Always update connection lines and packet spheres in the same frame using current mesh positions

    // Update dynamic connections (like Mycelium)
    let connectionIdx = 0
    const allNodes = [...clients, ...gatewayNodes, ...servers]
    const connectionDistance = 1.8

    for (let i = 0; i < allNodes.length; i++) {
      for (let j = i + 1; j < allNodes.length; j++) {
        if (allNodes[i].basePos.distanceTo(allNodes[j].basePos) < connectionDistance) {
          if (connectionIdx < connections.length) {
            const conn = connections[connectionIdx]
            // Always use current mesh positions for curve endpoints
            const p1 = allNodes[i].mesh.position
            const p2 = allNodes[j].mesh.position
            conn.line.geometry.setFromPoints([p1, p2])
            // Subtle opacity animation
            const material = conn.line.material as THREE.LineBasicMaterial
            material.opacity = 0.12 + Math.sin(t * 2 + connectionIdx * 0.3) * 0.08
          }
          connectionIdx++
        }
      }
    }

    // Animate pulses
    pulseCooldown--;
    if (pulseCooldown <= 0) {
      // Pick a random client, gateway, and server
      const clientIdx = Math.floor(Math.random() * clients.length);
      const gatewayIdx = Math.floor(Math.random() * gatewayNodes.length);
      const serverIdx = Math.floor(Math.random() * servers.length);
      const path = [
        clients[clientIdx].mesh.position.clone(),
        gatewayNodes[gatewayIdx].mesh.position.clone(),
        servers[serverIdx].mesh.position.clone()
      ];
      pulses.push({ mesh: createPulseMesh(), path, t: 0, stage: 0, active: true });
      pulseCooldown = 30 + Math.random() * 40;
    }
    pulses.forEach((pulse, idx) => {
      if (!pulse.active) return;
      pulse.t += 0.012;
      let pos: THREE.Vector3;
      if (pulse.stage === 0) {
        // Client to Gateway
        pos = new THREE.Vector3().lerpVectors(pulse.path[0], pulse.path[1], pulse.t * 2);
        if (pulse.t >= 0.5) {
          pulse.stage = 1;
          pulse.t = 0;
        }
      } else {
        // Gateway to Server
        pos = new THREE.Vector3().lerpVectors(pulse.path[1], pulse.path[2], pulse.t * 2);
        if (pulse.t >= 0.5) {
          pulse.active = false;
          scene!.remove(pulse.mesh);
        }
      }
      pulse.mesh.position.copy(pos);
      let mat = pulse.mesh.material;
      if (Array.isArray(mat)) mat = mat[0];
      if (mat && 'opacity' in mat) mat.opacity = 0.7 + 0.25 * Math.sin(Math.PI * pulse.t);
    });
    pulses = pulses.filter(p => p.active);

    renderer?.render(scene!, camera!)
  }
  animate()
})

onBeforeUnmount(() => {
  if (animationId) cancelAnimationFrame(animationId)

  // Clean up Three.js objects
  if (scene) {
    // Dispose of geometries and materials
    scene.traverse((object) => {
      if (object instanceof THREE.Mesh) {
        if (object.geometry) object.geometry.dispose()
        if (object.material) {
          if (Array.isArray(object.material)) {
            object.material.forEach(material => material.dispose())
          } else {
            object.material.dispose()
          }
        }
      }
    })
    scene.clear()
  }

  if (renderer) {
    renderer.dispose()
    renderer = null
  }

  // Clear arrays
  clients.length = 0
  servers.length = 0
  gatewayNodes.length = 0
  connections.length = 0
})

// Add this helper function if not present:
function createPulseMesh() {
  const pulseGeom = new THREE.SphereGeometry(0.06, 16, 16);
  const pulseMat = new THREE.MeshBasicMaterial({ color: 0xfbbf24, transparent: true, opacity: 0.95 });
  const mesh = new THREE.Mesh(pulseGeom, pulseMat);
  scene!.add(mesh);
  return mesh;
}
</script>

<script lang="ts">
export default {}
</script>

<style scoped>
.feature-panel {
  min-height: 60vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
}
.feature-content-row {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  max-width: 1600px;
  gap: 1.5rem;
  padding: 3rem 1rem;
  position: relative;
}
.feature-animation-with-glow {
  position: relative;
  width: 60vw;
  min-width: 400px;
  max-width: 900px;
  height: 600px;
  display: flex;
  align-items: center;
  justify-content: flex-end;
}
.feature-animation-glow {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 90%;
  height: 90%;
  transform: translate(-50%, -50%);
  background: radial-gradient(circle, rgba(96, 165, 250, 0.18) 0%, transparent 80%);
  z-index: 0;
  pointer-events: none;
  filter: blur(32px);
}
.feature-animation {
  position: relative;
  width: 100%;
  height: 100%;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}
.three-canvas {
  width: 100%;
  height: 100%;
  display: block;
  background: transparent;
}
.feature-content {
  flex: 1 1 350px;
  min-width: 260px;
  max-width: 420px;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  padding: 2rem 1.5rem;
  background: none;
  backdrop-filter: none;
  border-radius: 0;
  color: #fff;
  box-shadow: none;
  margin-left: 0;
  z-index: 2;
}
.feature-title {
  font-size: 1.7rem;
  font-weight: 500;
  margin-bottom: 1rem;
}
.feature-description {
  font-size: 1.1rem;
  font-weight: 400;
  color: #60a5fa;
}
.feature-benefits {
  margin: 1.2rem 0 0 0;
  color: #b6d6ff;
  font-size: 1rem;
  line-height: 1.7;
  display: flex;
  gap: 0.2rem;
}
@media (max-width: 1200px) {
  .feature-animation-with-glow {
    width: 90vw;
    max-width: 100vw;
    height: 400px;
    min-width: 0;
  }
  .feature-content {
    margin-left: -40px;
    max-width: 340px;
  }
}
@media (max-width: 900px) {
  .feature-content-row {
    flex-direction: column;
    gap: 2rem;
    padding: 2rem 0.5rem;
  }
  .feature-animation-with-glow {
    width: 100vw;
    max-width: 100vw;
    height: 320px;
    min-width: 0;
    justify-content: center;
  }
  .feature-content {
    align-items: center;
    text-align: center;
    margin-left: 0;
    max-width: 100vw;
    padding: 1.5rem 0.5rem;
  }
}
.node-label {
  position: absolute;
  background: rgba(30, 41, 59, 0.92);
  color: #8ecfff;
  font-size: 1rem;
  font-weight: 500;
  padding: 0.25rem 0.7rem;
  border-radius: 8px;
  pointer-events: none;
  z-index: 20;
  white-space: nowrap;
  box-shadow: 0 2px 8px rgba(96, 165, 250, 0.12);
  border: 1px solid #60a5fa33;
  transform: translate(-50%, -120%);
  user-select: none;
}
</style>
