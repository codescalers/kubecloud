<template>
  <section class="feature-panel multi-master">
    <div class="feature-content-row">
      <div class="feature-animation-with-glow">
        <div class="feature-animation-glow"></div>
        <div class="feature-animation">
          <canvas ref="threeCanvas" class="three-canvas" @mousemove="onCanvasMouseMove" @mouseleave="onCanvasMouseLeave"></canvas>
          <div v-if="hoveredNode" class="node-label" :style="{ left: hoveredNode.pos.x + 'px', top: hoveredNode.pos.y + 'px' }">
            {{ hoveredNode.type }}
          </div>
        </div>
      </div>
      <div class="feature-content feature-content-overlay">
        <h2 class="feature-title">Multi-Master Clusters</h2>
        <p class="feature-description">
          High-availability Kubernetes clusters with multiple control plane nodes. Automatic failover, leader election, and zero-downtime upgrades built-in.
        </p>
        <div class="feature-benefits">
          <v-chip class="ma-1" color="white" variant="outlined" size="small">HA Control Plane</v-chip>
          <v-chip class="ma-1" color="white" variant="outlined" size="small">Automatic Failover</v-chip>
          <v-chip class="ma-1" color="white" variant="outlined" size="small">Zero-downtime Upgrades</v-chip>
        </div>
      </div>
    </div>
  </section>
</template>

<script lang="ts">
export default {}
</script>

<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, shallowRef } from 'vue'
import * as THREE from 'three'

const threeCanvas = ref<HTMLCanvasElement | null>(null)
let renderer: THREE.WebGLRenderer | null = null
let animationId: number | null = null
let scene: THREE.Scene | null = null
let camera: THREE.PerspectiveCamera | null = null

const MASTER_COUNT = 5
const WORKER_COUNT = 7
const PRIMARY_COLOR = 0x60a5fa
const MASTER_COLOR = 0x8ecfff // More sophisticated light blue instead of cartoonish yellow
const LEADER_COLOR = 0x60a5fa // Use primary blue for leader instead of bright yellow
const WORKER_COLOR = 0x34d399
const CONNECTION_COLOR = 0x38bdf8 // Cyan for connections

const MASTER_RING_RADIUS = 1.2
const WORKER_RING_RADIUS = 2.0

const masters: { mesh: THREE.Mesh, glow: THREE.Mesh, basePos: THREE.Vector3, phase: number, crown?: THREE.Mesh }[] = []
const workers: { mesh: THREE.Mesh, glow: THREE.Mesh, basePos: THREE.Vector3, phase: number }[] = []
const connections: { line: THREE.Line, nodes: [THREE.Mesh, THREE.Mesh] }[] = []
const workerConnections: { line: THREE.Line, nodes: [THREE.Mesh, THREE.Mesh] }[] = []


// Leader state
let leaderIdx = 0
let leaderChangeCooldown = 0
let leaderTransition = 0 // 0..1 for smooth transition
let prevLeaderIdx = 0

// Globe-compatible animation states
let autoRotate = true
let rotationY = 0

// Add communication pulses between masters and workers
let mwCommPulses: { mesh: THREE.Mesh, from: THREE.Vector3, to: THREE.Vector3, t: number, active: boolean, direction: 'm2w' | 'w2m', workerIdx: number }[] = []
let mwCommCooldown = 0
let workerHighlightTimers: number[] = Array(WORKER_COUNT).fill(0)

const hoveredNode = shallowRef<{ mesh: THREE.Mesh, type: string, pos: { x: number, y: number } } | null>(null)

// Declare handleResize outside onMounted so it can be accessed in cleanup
let handleResize: (() => void) | null = null

function onCanvasMouseMove(event: MouseEvent) {
  if (!threeCanvas.value || !renderer || !camera) return
  const rect = threeCanvas.value.getBoundingClientRect()
  const mouse = new THREE.Vector2(
    ((event.clientX - rect.left) / rect.width) * 2 - 1,
    -((event.clientY - rect.top) / rect.height) * 2 + 1
  )
  const allMeshes = [
    ...masters.map((m, i) => ({ mesh: m.mesh, type: i === leaderIdx ? 'Leader' : 'Master' })),
    ...workers.map(w => ({ mesh: w.mesh, type: 'Worker' }))
  ]
  const raycaster = new THREE.Raycaster()
  raycaster.setFromCamera(mouse, camera)
  const intersects = raycaster.intersectObjects(allMeshes.map(n => n.mesh))
  if (intersects.length > 0) {
    const found = allMeshes.find(n => n.mesh === intersects[0].object)
    if (found) {
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
  
  // Set initial canvas size
  const canvas = threeCanvas.value
  const rect = canvas.getBoundingClientRect()
  
  // Set explicit canvas size if dimensions are 0
  if (rect.width === 0 || rect.height === 0) {
    canvas.width = 800
    canvas.height = 600
  }
  
  renderer = new THREE.WebGLRenderer({
    canvas: canvas,
    alpha: true,
    antialias: true,
    powerPreference: 'high-performance'
  })
  const renderWidth = rect.width > 0 ? rect.width : 800
  const renderHeight = rect.height > 0 ? rect.height : 600
  renderer.setSize(renderWidth, renderHeight, false)
  renderer.setClearColor(0x000000, 0) // Restore transparent background
  
  scene = new THREE.Scene()
  camera = new THREE.PerspectiveCamera(
    60,
    renderWidth / renderHeight,
    0.1,
    1000
  )
  camera.position.z = 5.5

  // Handle resize
  handleResize = () => {
    if (!threeCanvas.value || !renderer || !camera) return
    const rect = threeCanvas.value.getBoundingClientRect()
    const renderWidth = rect.width > 0 ? rect.width : 800
    const renderHeight = rect.height > 0 ? rect.height : 600
    renderer.setSize(renderWidth, renderHeight, false)
    camera.aspect = renderWidth / renderHeight
    camera.updateProjectionMatrix()
  }
  
  window.addEventListener('resize', handleResize)

  // Place master nodes in a ring
  const MASTER_RADIUS = 0.09; // Reduced from 0.18
  const MASTER_GLOW_RADIUS = 0.15; // Reduced from 0.32
  for (let i = 0; i < MASTER_COUNT; i++) {
    const angle = (i / MASTER_COUNT) * Math.PI * 2 - Math.PI / 2
    const basePos = new THREE.Vector3(
      MASTER_RING_RADIUS * Math.cos(angle),
      MASTER_RING_RADIUS * Math.sin(angle),
      0
    )
    const masterGeometry = new THREE.SphereGeometry(MASTER_RADIUS, 18, 18);
    const masterGlowGeometry = new THREE.SphereGeometry(MASTER_GLOW_RADIUS, 18, 18);
    const mesh = new THREE.Mesh(masterGeometry, new THREE.MeshBasicMaterial({ color: MASTER_COLOR }));
    mesh.position.copy(basePos);
    scene.add(mesh);
    const glow = new THREE.Mesh(masterGlowGeometry, new THREE.MeshBasicMaterial({ color: MASTER_COLOR, transparent: true, opacity: 0.18 }));
    glow.position.copy(basePos);
    scene.add(glow);
    // Add leader highlight (ring) to the leader master
    let crown = undefined;
    if (i === leaderIdx) {
      const ringGeometry = new THREE.TorusGeometry(MASTER_RADIUS + 0.03, 0.012, 16, 32);
      const ringMaterial = new THREE.MeshBasicMaterial({ color: 0xfbbf24, transparent: true, opacity: 0.7 });
      crown = new THREE.Mesh(ringGeometry, ringMaterial);
      crown.position.copy(basePos);
      crown.rotation.x = Math.PI / 2;
      scene.add(crown);
    }
    masters.push({ mesh, glow, basePos, phase: Math.random() * Math.PI * 2, crown });
  }
  // Place worker nodes in a ring outside
  const WORKER_RADIUS = 0.07; // Reduced from 0.12
  const WORKER_GLOW_RADIUS = 0.12; // Reduced from 0.21
  for (let i = 0; i < WORKER_COUNT; i++) {
    const angle = (i / WORKER_COUNT) * Math.PI * 2
    const basePos = new THREE.Vector3(
      WORKER_RING_RADIUS * Math.cos(angle),
      WORKER_RING_RADIUS * Math.sin(angle),
      0
    )
    const workerGeometry = new THREE.SphereGeometry(WORKER_RADIUS, 18, 18);
    const workerGlowGeometry = new THREE.SphereGeometry(WORKER_GLOW_RADIUS, 18, 18);
    const mesh = new THREE.Mesh(workerGeometry, new THREE.MeshBasicMaterial({ color: WORKER_COLOR }));
    mesh.position.copy(basePos);
    scene.add(mesh);
    const glow = new THREE.Mesh(workerGlowGeometry, new THREE.MeshBasicMaterial({ color: WORKER_COLOR, transparent: true, opacity: 0.18 }));
    glow.position.copy(basePos);
    scene.add(glow);
    workers.push({ mesh, glow, basePos, phase: Math.random() * Math.PI * 2 });
  }
  // Connect masters to each other (full mesh)
  for (let i = 0; i < MASTER_COUNT; i++) {
    for (let j = i + 1; j < MASTER_COUNT; j++) {
      const lineGeom = new THREE.BufferGeometry().setFromPoints([
        masters[i].mesh.position,
        masters[j].mesh.position
      ])
      const lineMat = new THREE.LineBasicMaterial({ color: CONNECTION_COLOR, transparent: true, opacity: 0.22 })
      const line = new THREE.Line(lineGeom, lineMat)
      scene.add(line)
      connections.push({ line, nodes: [masters[i].mesh, masters[j].mesh] })
    }
  }
  // Connect each worker to all masters
  for (let i = 0; i < WORKER_COUNT; i++) {
    for (let j = 0; j < MASTER_COUNT; j++) {
      const lineGeom = new THREE.BufferGeometry().setFromPoints([
        workers[i].mesh.position,
        masters[j].mesh.position
      ])
      const lineMat = new THREE.LineBasicMaterial({ color: PRIMARY_COLOR, transparent: true, opacity: 0.13 })
      const line = new THREE.Line(lineGeom, lineMat)
      scene.add(line)
      workerConnections.push({ line, nodes: [workers[i].mesh, masters[j].mesh] })
    }
  }
  // Animate
  function animate(time = 0) {
    animationId = requestAnimationFrame(animate)
    const t = time * 0.001

    // Auto-rotate like globe
    if (autoRotate) {
      rotationY += 0.0005
    }

    // Animate masters with calmer floating motion
    masters.forEach((master, idx) => {
      master.mesh.position.x = master.basePos.x + Math.sin(t * 0.7 + master.phase + idx * 0.3) * 0.04
      master.mesh.position.y = master.basePos.y + Math.cos(t * 0.9 + master.phase + idx * 0.4) * 0.04
      master.mesh.position.z = master.basePos.z + Math.sin(t * 0.5 + master.phase + idx * 0.2) * 0.03
      master.glow.position.copy(master.mesh.position)
      if (master.crown) master.crown.position.copy(master.mesh.position)
      master.mesh.rotation.y = rotationY * 0.2 + idx * 0.6
    })

    // Animate workers with calmer floating motion
    workers.forEach((worker, idx) => {
      worker.mesh.position.x = worker.basePos.x + Math.sin(t * 0.7 + worker.phase + idx * 0.2) * 0.03
      worker.mesh.position.y = worker.basePos.y + Math.cos(t * 0.9 + worker.phase + idx * 0.3) * 0.03
      worker.mesh.position.z = worker.basePos.z + Math.sin(t * 0.5 + worker.phase + idx * 0.1) * 0.02
      worker.glow.position.copy(worker.mesh.position)
      worker.mesh.rotation.y = rotationY * 0.15 + idx * 0.4
    })

    // Update connections with enhanced opacity
    connections.forEach((conn, idx) => {
      conn.line.geometry.setFromPoints([
        conn.nodes[0].position,
        conn.nodes[1].position
      ])
      // Enhanced opacity animation like globe
      const material = conn.line.material as THREE.LineBasicMaterial
      material.opacity = 0.18 + Math.sin(t * 2 + idx * 0.3) * 0.08
    })

    workerConnections.forEach((conn, idx) => {
      conn.line.geometry.setFromPoints([
        conn.nodes[0].position,
        conn.nodes[1].position
      ])
      // Enhanced opacity animation like globe
      const material = conn.line.material as THREE.LineBasicMaterial
      material.opacity = 0.12 + Math.sin(t * 2 + idx * 0.3) * 0.06
    })

    // Leader election animation
    leaderChangeCooldown--
    if (leaderChangeCooldown <= 0) {
      prevLeaderIdx = leaderIdx
      leaderIdx = (leaderIdx + 1) % MASTER_COUNT
      leaderTransition = 0
      leaderChangeCooldown = 180 + Math.random() * 80
      // Move crown
      if (masters[prevLeaderIdx].crown && scene) {
        scene.remove(masters[prevLeaderIdx].crown!)
      }
      masters[leaderIdx].crown = createCrownMesh()
      if (masters[leaderIdx].crown) {
        masters[leaderIdx].crown!.position.copy(masters[leaderIdx].mesh.position)
        if (scene) {
          scene.add(masters[leaderIdx].crown!)
        }
      }
    }

    // Animate crown scale for leader
    masters.forEach((master, idx) => {
      if (master.crown) {
        let scale = idx === leaderIdx ? 1 + 0.13 * Math.sin(t * 2.5) : 1
        master.crown.scale.set(scale, scale, scale)
      }
      // Animate leader glow with enhanced effects
      if (idx === leaderIdx) {
        let glowMat = master.glow.material
        if (Array.isArray(glowMat)) glowMat = glowMat[0]
        if (glowMat && 'opacity' in glowMat) {
          glowMat.opacity = 0.35 + 0.12 * Math.abs(Math.sin(t * 2.5))
        }
        let meshMat = master.mesh.material
        if (Array.isArray(meshMat)) meshMat = meshMat[0]
        if (meshMat && 'color' in meshMat && meshMat.color instanceof THREE.Color) {
          meshMat.color.set(LEADER_COLOR)
        }
      } else {
        let glowMat = master.glow.material
        if (Array.isArray(glowMat)) glowMat = glowMat[0]
        if (glowMat && 'opacity' in glowMat) {
          glowMat.opacity = 0.18 + 0.05 * Math.sin(t * 1.5 + idx)
        }
        let meshMat = master.mesh.material
        if (Array.isArray(meshMat)) meshMat = meshMat[0]
        if (meshMat && 'color' in meshMat && meshMat.color instanceof THREE.Color) {
          meshMat.color.set(MASTER_COLOR)
        }
      }
    })

    // Generate master→worker and worker→master communication pulses
    if (mwCommPulses.length === 0) {
      mwCommCooldown--;
      if (mwCommCooldown <= 0) {
        const masterIdx = Math.floor(Math.random() * MASTER_COUNT)
        const workerIdx = Math.floor(Math.random() * WORKER_COUNT)
        mwCommPulses.push({
          mesh: createCommPulseMesh('m2w'),
          from: masters[masterIdx].mesh.position.clone(),
          to: workers[workerIdx].mesh.position.clone(),
          t: 0,
          active: true,
          direction: 'm2w',
          workerIdx
        })
        if (Math.random() < 0.5) {
          mwCommPulses.push({
            mesh: createCommPulseMesh('w2m'),
            from: workers[workerIdx].mesh.position.clone(),
            to: masters[masterIdx].mesh.position.clone(),
            t: 0,
            active: true,
            direction: 'w2m',
            workerIdx
          })
        }
        mwCommCooldown = 200 + Math.random() * 120
      }
    }
    // Animate communication pulses
    mwCommPulses = mwCommPulses.filter(pulse => {
      pulse.t += 0.025
      if (pulse.t >= 1) {
        if (scene) scene.remove(pulse.mesh)
        // Highlight worker node if pulse was master→worker
        if (pulse.direction === 'm2w') {
          workerHighlightTimers[pulse.workerIdx] = 12 // frames to highlight
        }
        return false
      }
      pulse.mesh.position.lerpVectors(pulse.from, pulse.to, pulse.t)
      const material = pulse.mesh.material as THREE.MeshBasicMaterial
      material.opacity = 0.7 * (1 - pulse.t)
      return true
    })
    // Animate worker highlight
    workers.forEach((worker, idx) => {
      if (workerHighlightTimers[idx] > 0) {
        workerHighlightTimers[idx]--
        let mat = worker.mesh.material
        if (Array.isArray(mat)) mat = mat[0]
        if (mat && 'color' in mat && mat.color instanceof THREE.Color) {
          mat.color.lerp(new THREE.Color(WORKER_COLOR), 0.15)
          mat.color.offsetHSL(0, 0, 0.08)
        }
        let glowMat = worker.glow.material
        if (Array.isArray(glowMat)) glowMat = glowMat[0]
        if (glowMat && 'opacity' in glowMat) {
          glowMat.opacity = 0.18 + 0.17 * (workerHighlightTimers[idx] / 12)
        }
      } else {
        let mat = worker.mesh.material
        if (Array.isArray(mat)) mat = mat[0]
        if (mat && 'color' in mat && mat.color instanceof THREE.Color) {
          mat.color.set(WORKER_COLOR)
        }
        let glowMat = worker.glow.material
        if (Array.isArray(glowMat)) glowMat = glowMat[0]
        if (glowMat && 'opacity' in glowMat) {
          glowMat.opacity = 0.18
        }
      }
    })

    renderer!.render(scene!, camera!)
  }

  function createCrownMesh() {
    // Sophisticated leader indicator: subtle ring with glow, but as a single mesh
    const ringGeometry = new THREE.TorusGeometry(0.25, 0.03, 16, 32)
    const ringMaterial = new THREE.MeshBasicMaterial({ 
      color: LEADER_COLOR, 
      transparent: true, 
      opacity: 0.6
    })
    const ring = new THREE.Mesh(ringGeometry, ringMaterial)
    ring.position.z = 0.15
    return ring
  }

  // Helper to create a communication pulse mesh
  function createCommPulseMesh(direction: 'm2w' | 'w2m') {
    const color = direction === 'm2w' ? 0x60a5fa : 0x34d399 // blue for master→worker, green for worker→master
    const geometry = new THREE.SphereGeometry(0.05, 10, 10)
    const material = new THREE.MeshBasicMaterial({ color, transparent: true, opacity: 0.7 })
    const mesh = new THREE.Mesh(geometry, material)
    if (scene) scene.add(mesh)
    return mesh
  }

  animate()
})

onBeforeUnmount(() => {
  if (animationId) cancelAnimationFrame(animationId)
  if (renderer) {
    renderer.dispose()
    renderer = null
  }
  if (handleResize) {
    window.removeEventListener('resize', handleResize)
  }
})
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
  justify-content: center;
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
  border-radius: 16px;
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
  font-size: 0.9rem;
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