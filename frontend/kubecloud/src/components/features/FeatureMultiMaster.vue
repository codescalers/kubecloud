<template>
  <section class="feature-panel multi-master">
    <div class="feature-content-row">
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
      <div class="feature-animation-with-glow">
        <div class="feature-animation-glow"></div>
        <div class="feature-animation">
          <canvas ref="threeCanvas" class="three-canvas" @mousemove="onCanvasMouseMove" @mouseleave="onCanvasMouseLeave"></canvas>
          <div v-if="hoveredNode" class="node-label" :style="nodeLabelStyle">
            {{ hoveredNode.type }}
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script lang="ts">
export default {}
</script>

<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, shallowRef, computed } from 'vue'
import * as THREE from 'three'

const threeCanvas = ref<HTMLCanvasElement | null>(null)
let renderer: THREE.WebGLRenderer | null = null
let animationId: number | null = null
let scene: THREE.Scene | null = null
let camera: THREE.PerspectiveCamera | null = null

const MASTER_COUNT = 5
const WORKER_COUNT = 7
const RING_RADIUS = 1.6
const WORKER_RADIUS = 2.7
const PRIMARY_COLOR = 0x60a5fa
const MASTER_COLOR = 0xfbbf24 // gold/yellow for master
const LEADER_COLOR = 0xfff200 // bright yellow for leader
const WORKER_COLOR = 0x34d399
const CONNECTION_COLOR = 0x8ecfff

const masters: { mesh: THREE.Mesh, glow: THREE.Mesh, basePos: THREE.Vector3, phase: number, crown?: THREE.Mesh }[] = []
const workers: { mesh: THREE.Mesh, glow: THREE.Mesh, basePos: THREE.Vector3, phase: number }[] = []
const connections: { line: THREE.Line, nodes: [THREE.Mesh, THREE.Mesh] }[] = []
const workerConnections: { line: THREE.Line, nodes: [THREE.Mesh, THREE.Mesh] }[] = []

// Consensus pulses
let pulses: { mesh: THREE.Mesh, from: number, to: number, t: number, active: boolean }[] = []
let pulseCooldown = 0

// Leader state
let leaderIdx = 0
let leaderChangeCooldown = 0
let leaderTransition = 0 // 0..1 for smooth transition
let prevLeaderIdx = 0

const hoveredNode = shallowRef<{ mesh: THREE.Mesh, type: string, pos: { x: number, y: number } } | null>(null)

const WORKER_LABEL_COLOR = '#34d399'
const MASTER_LABEL_COLOR = '#fbbf24'

const nodeLabelStyle = computed(() => {
  if (!hoveredNode.value) return {}
  return {
    left: hoveredNode.value.pos.x + 'px',
    top: hoveredNode.value.pos.y + 'px',
    color: hoveredNode.value.type === 'Worker' ? WORKER_LABEL_COLOR : MASTER_LABEL_COLOR,
    borderColor: hoveredNode.value.type === 'Worker' ? '#34d39933' : '#fbbf2433',
  }
})

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
  renderer = new THREE.WebGLRenderer({
    canvas: threeCanvas.value,
    alpha: true,
    antialias: true,
    powerPreference: 'high-performance'
  })
  renderer.setSize(threeCanvas.value.clientWidth, threeCanvas.value.clientHeight, false)
  renderer.setClearColor(0x000000, 0)
  scene = new THREE.Scene()
  camera = new THREE.PerspectiveCamera(
    60,
    threeCanvas.value.clientWidth / threeCanvas.value.clientHeight,
    0.1,
    1000
  )
  camera.position.z = 5.5

  // Place master nodes in a ring
  for (let i = 0; i < MASTER_COUNT; i++) {
    const angle = (i / MASTER_COUNT) * Math.PI * 2 - Math.PI / 2
    const basePos = new THREE.Vector3(
      RING_RADIUS * Math.cos(angle),
      RING_RADIUS * Math.sin(angle),
      0
    )
    const mesh = new THREE.Mesh(
      new THREE.SphereGeometry(0.18, 24, 24),
      new THREE.MeshBasicMaterial({ color: MASTER_COLOR })
    )
    mesh.position.copy(basePos)
    scene.add(mesh)
    // Glow
    const glow = new THREE.Mesh(
      new THREE.SphereGeometry(0.32, 24, 24),
      new THREE.MeshBasicMaterial({ color: MASTER_COLOR, transparent: true, opacity: 0.18 })
    )
    glow.position.copy(basePos)
    scene.add(glow)
    // Crown (for leader)
    let crown: THREE.Mesh | undefined
    if (i === leaderIdx) {
      crown = createCrownMesh()
      crown.position.copy(basePos)
      scene.add(crown)
    }
    masters.push({ mesh, glow, basePos, phase: Math.random() * Math.PI * 2, crown })
  }
  // Place worker nodes in a ring outside
  for (let i = 0; i < WORKER_COUNT; i++) {
    const angle = (i / WORKER_COUNT) * Math.PI * 2
    const basePos = new THREE.Vector3(
      WORKER_RADIUS * Math.cos(angle),
      WORKER_RADIUS * Math.sin(angle),
      0
    )
    const mesh = new THREE.Mesh(
      new THREE.SphereGeometry(0.12, 18, 18),
      new THREE.MeshBasicMaterial({ color: WORKER_COLOR })
    )
    mesh.position.copy(basePos)
    scene.add(mesh)
    // Glow
    const glow = new THREE.Mesh(
      new THREE.SphereGeometry(0.21, 18, 18),
      new THREE.MeshBasicMaterial({ color: WORKER_COLOR, transparent: true, opacity: 0.18 })
    )
    glow.position.copy(basePos)
    scene.add(glow)
    workers.push({ mesh, glow, basePos, phase: Math.random() * Math.PI * 2 })
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
    // Animate masters (gentle organic motion)
    masters.forEach((master, idx) => {
      master.mesh.position.x = master.basePos.x + Math.sin(t * 0.7 + master.phase + idx) * 0.09
      master.mesh.position.y = master.basePos.y + Math.cos(t * 0.9 + master.phase + idx * 1.2) * 0.09
      master.mesh.position.z = master.basePos.z + Math.sin(t * 0.5 + master.phase + idx * 0.7) * 0.07
      master.glow.position.copy(master.mesh.position)
      if (master.crown) master.crown.position.copy(master.mesh.position)
    })
    // Animate workers (gentle organic motion)
    workers.forEach((worker, idx) => {
      worker.mesh.position.x = worker.basePos.x + Math.sin(t * 0.7 + worker.phase + idx) * 0.07
      worker.mesh.position.y = worker.basePos.y + Math.cos(t * 0.9 + worker.phase + idx * 1.2) * 0.07
      worker.mesh.position.z = worker.basePos.z + Math.sin(t * 0.5 + worker.phase + idx * 0.7) * 0.05
      worker.glow.position.copy(worker.mesh.position)
    })
    // Update connections
    connections.forEach(conn => {
      conn.line.geometry.setFromPoints([
        conn.nodes[0].position,
        conn.nodes[1].position
      ])
    })
    workerConnections.forEach(conn => {
      conn.line.geometry.setFromPoints([
        conn.nodes[0].position,
        conn.nodes[1].position
      ])
    })
    // Consensus pulses between masters
    pulseCooldown--
    if (pulseCooldown <= 0) {
      // Launch a new pulse between two random masters
      let from = Math.floor(Math.random() * MASTER_COUNT)
      let to = from
      while (to === from) to = Math.floor(Math.random() * MASTER_COUNT)
      pulses.push({
        mesh: createPulseMesh(),
        from,
        to,
        t: 0,
        active: true
      })
      pulseCooldown = 30 + Math.random() * 40
    }
    pulses.forEach((pulse, idx) => {
      if (!pulse.active) return
      pulse.t += 0.035
      const fromPos = masters[pulse.from].mesh.position
      const toPos = masters[pulse.to].mesh.position
      pulse.mesh.position.lerpVectors(fromPos, toPos, pulse.t)
      let mat = pulse.mesh.material
      if (Array.isArray(mat)) mat = mat[0]
      if (mat && 'opacity' in mat) mat.opacity = 0.5 + 0.45 * Math.sin(Math.PI * pulse.t)
      if (pulse.t >= 1) {
        pulse.active = false
        scene!.remove(pulse.mesh)
      }
    })
    pulses = pulses.filter(p => p.active)
    // Leader election animation
    leaderChangeCooldown--
    if (leaderChangeCooldown <= 0) {
      prevLeaderIdx = leaderIdx
      leaderIdx = (leaderIdx + 1) % MASTER_COUNT
      leaderTransition = 0
      leaderChangeCooldown = 180 + Math.random() * 80
      // Move crown
      if (masters[prevLeaderIdx].crown) scene!.remove(masters[prevLeaderIdx].crown!)
      masters[leaderIdx].crown = createCrownMesh()
      masters[leaderIdx].crown!.position.copy(masters[leaderIdx].mesh.position)
      scene!.add(masters[leaderIdx].crown!)
    }
    // Animate crown scale for leader
    masters.forEach((master, idx) => {
      if (master.crown) {
        let scale = idx === leaderIdx ? 1 + 0.13 * Math.sin(t * 2.5) : 1
        master.crown.scale.set(scale, scale, scale)
      }
      // Animate leader glow
      if (idx === leaderIdx) {
        let glowMat = master.glow.material
        if (Array.isArray(glowMat)) glowMat = glowMat[0]
        if (glowMat && 'opacity' in glowMat) {
          glowMat.opacity = 0.32 + 0.10 * Math.abs(Math.sin(t * 2.5))
        }
        let meshMat = master.mesh.material
        if (Array.isArray(meshMat)) meshMat = meshMat[0]
        if (meshMat && 'color' in meshMat && meshMat.color instanceof THREE.Color) meshMat.color.set(LEADER_COLOR)
      } else {
        let glowMat = master.glow.material
        if (Array.isArray(glowMat)) glowMat = glowMat[0]
        if (glowMat && 'opacity' in glowMat) {
          glowMat.opacity = 0.18
        }
        let meshMat = master.mesh.material
        if (Array.isArray(meshMat)) meshMat = meshMat[0]
        if (meshMat && 'color' in meshMat && meshMat.color instanceof THREE.Color) meshMat.color.set(MASTER_COLOR)
      }
    })
    renderer!.render(scene!, camera!)
  }
  animate()
})

function createPulseMesh() {
  const pulseGeom = new THREE.SphereGeometry(0.09, 16, 16)
  const pulseMat = new THREE.MeshBasicMaterial({ color: 0xffffff, transparent: true, opacity: 0.92 })
  const mesh = new THREE.Mesh(pulseGeom, pulseMat)
  scene!.add(mesh)
  return mesh
}
function createCrownMesh() {
  // Simple crown: 3-pointed polygon
  const shape = new THREE.Shape()
  shape.moveTo(-0.13, 0)
  shape.lineTo(-0.09, 0.13)
  shape.lineTo(0, 0.09)
  shape.lineTo(0.09, 0.13)
  shape.lineTo(0.13, 0)
  shape.lineTo(0, 0)
  const extrudeSettings = { depth: 0.04, bevelEnabled: false }
  const geometry = new THREE.ExtrudeGeometry(shape, extrudeSettings)
  const material = new THREE.MeshBasicMaterial({ color: 0xfff200, transparent: true, opacity: 0.92 })
  const mesh = new THREE.Mesh(geometry, material)
  mesh.position.z = 0.23
  return mesh
}

onBeforeUnmount(() => {
  if (animationId) cancelAnimationFrame(animationId)
  if (renderer) {
    renderer.dispose()
    renderer = null
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
  font-size: 1rem;
  font-weight: 500;
  padding: 0.25rem 0.7rem;
  border-radius: 8px;
  pointer-events: none;
  z-index: 20;
  white-space: nowrap;
  box-shadow: 0 2px 8px rgba(96, 165, 250, 0.12);
  border: 1px solid;
  transform: translate(-50%, -120%);
  user-select: none;
}
</style> 