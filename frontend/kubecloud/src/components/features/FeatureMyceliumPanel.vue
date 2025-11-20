<template>
  <v-container class="py-8 py-md-12" fluid>
    <v-row align="center" justify="center" class="mx-auto" style="min-height: 60vh; max-width: 1600px;">
      <v-col cols="12" md="7" class="d-flex align-center justify-center justify-md-end pa-4 pa-md-6">
        <div class="canvas-container w-100 position-relative" style="max-width: 900px; min-width: 0;">
          <div class="feature-animation-glow"></div>
          <div class="w-100 h-100 position-relative d-flex align-center justify-center" style="z-index: 1;">
            <canvas 
              ref="threeCanvas" 
              class="w-100 h-100" 
              style="display: block; background: transparent;"
              @mousemove="handleMouseMove"
              @mouseleave="handleMouseLeave"
              @click="handleClick"
            ></canvas>
            <v-chip
              v-if="hoveredNodeLabel"
              size="small"
              class="elevation-2"
              :style="{
                position: 'absolute',
                left: hoveredNodeLabel.x + 'px',
                top: hoveredNodeLabel.y + 'px',
                transform: 'translate(-50%, -120%)',
                'pointer-events': 'none',
                'user-select': 'none',
                color: '#8ecfff',
                'border-color': '#60a5fa33',
                'border-width': '1px',
                'border-style': 'solid',
                'background': 'rgba(30,41,59,0.92)'
              }"
            >
              Peer
            </v-chip>
          </div>
        </div>
      </v-col>
      <v-col cols="12" md="5" class="d-flex flex-column align-center align-md-start pa-6 text-white text-center text-md-start">
        <div>
          <h2 class="text-h4 text-md-h3 font-weight-medium mb-3">Mycelium Networking</h2>
          <div class="subtitle" style="color:#60a5fa">
            Ultra-fast, decentralized networking inspired by nature. Mycelium Networking forms a resilient, adaptive mesh that routes around failures and optimizes for speed and security.
          </div>
          <div class="d-flex flex-wrap justify-center justify-md-start mt-2" style="gap: 8px; color: #b6d6ff;">
            <v-chip class="ma-1" size="small">End-to-end encrypted</v-chip>
            <v-chip class="ma-1" size="small">Nature-inspired</v-chip>
          </div>
        </div>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, shallowRef } from 'vue'
import * as THREE from 'three'

const threeCanvas = ref<HTMLCanvasElement | null>(null)
let renderer: THREE.WebGLRenderer | null = null
let animationId: number | null = null

const NODE_COUNT = 32
const CONNECTION_DISTANCE = 2.2
const nodes: { mesh: THREE.Mesh, basePos: THREE.Vector3, phase: number, glow: THREE.Mesh }[] = []
const ACTIVE_CONNECTIONS = 12
const activeConnections: { i: number, j: number, line: THREE.Line, t: number, direction: 1 | -1, opacity: number, fadingOut: boolean, maxOpacity: number, fadeInSpeed: number, fadeOutSpeed: number, lifetime: number, fadeInDelay: number, fadeInWait: number }[] = []

// Interactivity state
const hoveredNodeIdx = shallowRef<number | null>(null)
let pulse = 0
let pulseStart = 0
const raycaster = new THREE.Raycaster()
const mouse = new THREE.Vector2()

// Restore original lines array and organic growth logic
const lines: { line: THREE.Line, i: number, j: number, progress: number, growing: boolean }[] = []

// Add hover label state
const hoveredNodeLabel = shallowRef<{ x: number, y: number } | null>(null)

function handleMouseMove(event: MouseEvent) {
  if (!threeCanvas.value || !renderer) return
  const rect = threeCanvas.value.getBoundingClientRect()
  mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1
  mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1
  // No need to access camera here; label projection is handled in animate()
}
function handleMouseLeave() {
  hoveredNodeIdx.value = null
}
function handleClick() {
  if (hoveredNodeIdx.value !== null) {
    pulse = 1
    pulseStart = performance.now()
  }
}

function createPulseMesh() {
  const geometry = new THREE.SphereGeometry(0.11, 12, 12)
  const material = new THREE.MeshBasicMaterial({ color: 0xffffff, transparent: true, opacity: 0.95 })
  return new THREE.Mesh(geometry, material)
}

function addRandomConnection(scene: THREE.Scene) {
  // Pick two distinct nodes not already connected
  let tries = 0
  while (tries < 30) {
    const i = Math.floor(Math.random() * NODE_COUNT)
    let j = Math.floor(Math.random() * NODE_COUNT)
    while (j === i) j = Math.floor(Math.random() * NODE_COUNT)
    // Check not already connected
    if (!activeConnections.some(c => (c.i === i && c.j === j) || (c.i === j && c.j === i))) {
      // Create line
      const points = [nodes[i].mesh.position, nodes[j].mesh.position]
      const lineGeom = new THREE.BufferGeometry().setFromPoints(points)
      const lineMat = new THREE.LineBasicMaterial({ color: 0x60a5fa, transparent: true, opacity: 0 }) // Start invisible
      const line = new THREE.Line(lineGeom, lineMat)
      scene.add(line)
      // Random direction
      const direction = Math.random() > 0.5 ? 1 : -1
      // Random max opacity, fade speeds, lifetime, and fade-in delay
      const maxOpacity = 0.22 + Math.random() * 0.16 // 0.22–0.38
      const fadeInSpeed = 0.02 + Math.random() * 0.03 // 0.02–0.05
      const fadeOutSpeed = 0.02 + Math.random() * 0.03 // 0.02–0.05
      const lifetime = 1.2 + Math.random() * 2.3 // 1.2–3.5s
      const fadeInDelay = Math.random() * 0.5 // 0–0.5s
      activeConnections.push({ i, j, line, t: 0, direction, opacity: 0, fadingOut: false, maxOpacity, fadeInSpeed, fadeOutSpeed, lifetime, fadeInDelay, fadeInWait: 0 })
      break
    }
    tries++
  }
}

function removeConnection(scene: THREE.Scene, idx: number) {
  const conn = activeConnections[idx]
  scene.remove(conn.line)
  activeConnections.splice(idx, 1)
}

onMounted(() => {
  if (!threeCanvas.value) return
  renderer = new THREE.WebGLRenderer({ canvas: threeCanvas.value, alpha: true, antialias: true })
  renderer.setSize(threeCanvas.value.clientWidth / 1, threeCanvas.value.clientHeight / 1, false)
  renderer.setClearColor(0x000000, 0)
  const scene = new THREE.Scene()
  const camera = new THREE.PerspectiveCamera(
    60,
    threeCanvas.value.clientWidth / threeCanvas.value.clientHeight,
    0.1,
    1000
  )
  camera.position.z = 6

  // Create nodes (glowing points)
  const nodeGeometry = new THREE.OctahedronGeometry(0.08, 0) // Diamond/octahedron shape for peers
  const nodeMaterial = new THREE.MeshBasicMaterial({ color: 0x60a5fa })
  for (let i = 0; i < NODE_COUNT; i++) {
    const phi = Math.acos(2 * Math.random() - 1)
    const theta = 2 * Math.PI * Math.random()
    const r = 1.7 + Math.random() * 1.1
    const basePos = new THREE.Vector3(
      r * Math.sin(phi) * Math.cos(theta),
      r * Math.sin(phi) * Math.sin(theta),
      r * Math.cos(phi)
    )
    const mesh = new THREE.Mesh(nodeGeometry, nodeMaterial.clone())
    mesh.position.copy(basePos)
    scene.add(mesh)
    // Add glow - also diamond shaped but larger
    const glowMat = new THREE.MeshBasicMaterial({ color: 0x60a5fa, transparent: true, opacity: 0.18 })
    const glow = new THREE.Mesh(new THREE.OctahedronGeometry(0.18, 0), glowMat)
    glow.position.copy(basePos)
    scene.add(glow)
    nodes.push({ mesh, basePos, phase: Math.random() * Math.PI * 2, glow })
  }

  // Create lines (connections)
  for (let i = 0; i < NODE_COUNT; i++) {
    for (let j = i + 1; j < NODE_COUNT; j++) {
      if (nodes[i].basePos.distanceTo(nodes[j].basePos) < CONNECTION_DISTANCE) {
        const points = [nodes[i].mesh.position, nodes[j].mesh.position]
        const lineGeom = new THREE.BufferGeometry().setFromPoints(points)
        const lineMat = new THREE.LineBasicMaterial({ color: 0x60a5fa, transparent: true, opacity: 0.22 })
        const line = new THREE.Line(lineGeom, lineMat)
        lineGeom.setDrawRange(0, 0) // Start hidden
        scene.add(line)
        lines.push({ line, i, j, progress: 0, growing: false })
      }
    }
  }

  // Animate line growth (organic spreading)
  let nextLineToGrow = 0
  function growNextLine() {
    if (nextLineToGrow < lines.length) {
      lines[nextLineToGrow].growing = true
      nextLineToGrow++
      setTimeout(growNextLine, 120 + Math.random() * 180)
    }
  }
  growNextLine()

  // Animate nodes with organic motion and interactivity
  function animate(time = 0) {
    animationId = requestAnimationFrame(animate)
    // Animate node positions
    nodes.forEach((node, idx) => {
      const t = time * 0.001 + node.phase
      node.mesh.position.x = node.basePos.x + Math.sin(t * 0.7 + idx) * 0.18
      node.mesh.position.y = node.basePos.y + Math.cos(t * 0.9 + idx * 1.2) * 0.18
      node.mesh.position.z = node.basePos.z + Math.sin(t * 0.5 + idx * 0.7) * 0.18
      node.glow.position.copy(node.mesh.position)
    })
    // Update line geometry and animate growth
    let lineIdx = 0
    for (let i = 0; i < NODE_COUNT; i++) {
      for (let j = i + 1; j < NODE_COUNT; j++) {
        if (nodes[i].basePos.distanceTo(nodes[j].basePos) < CONNECTION_DISTANCE) {
          const lineObj = lines[lineIdx++]
          const { line, i: ni, j: nj, growing } = lineObj
          line.geometry.setFromPoints([
            nodes[ni].mesh.position,
            nodes[nj].mesh.position
          ])
          // Animate growth
          if (growing && lineObj.progress < 1) {
            lineObj.progress += 0.045 + Math.random() * 0.02
            if (lineObj.progress > 1) lineObj.progress = 1
            line.geometry.setDrawRange(0, Math.floor(2 * lineObj.progress))
          } else if (growing) {
            line.geometry.setDrawRange(0, 2)
          } else {
            line.geometry.setDrawRange(0, 0)
          }
        }
      }
    }
    // Raycasting for node hover
    hoveredNodeIdx.value = null
    hoveredNodeLabel.value = null
    raycaster.setFromCamera(mouse, camera)
    const intersects = raycaster.intersectObjects(nodes.map(n => n.mesh))
    if (intersects.length > 0) {
      hoveredNodeIdx.value = nodes.findIndex(n => n.mesh === intersects[0].object)
      // Project 3D position to 2D overlay for label
      const pos = nodes[hoveredNodeIdx.value].mesh.position.clone().project(camera)
      const rect = threeCanvas.value!.getBoundingClientRect()
      const x = ((pos.x + 1) / 2) * rect.width
      const y = ((-pos.y + 1) / 2) * rect.height
      hoveredNodeLabel.value = { x, y }
    }
    // Highlight hovered node and its glow
    nodes.forEach((node, idx) => {
      let glowMat = node.glow.material
      if (Array.isArray(glowMat)) glowMat = glowMat[0]
      let meshMat = node.mesh.material
      if (Array.isArray(meshMat)) meshMat = meshMat[0]
      // Gentle pulse for all nodes
      const basePulse = 1 + Math.sin(time * 0.002 + idx) * 0.08
      // If hovered, brighten and color
      if (idx === hoveredNodeIdx.value) {
        if (glowMat && 'opacity' in glowMat) glowMat.opacity = 0.45
        if (meshMat && 'color' in meshMat && meshMat.color instanceof THREE.Color) meshMat.color.set(0x8ecfff)
        node.mesh.scale.set(basePulse * 1.12, basePulse * 1.12, basePulse * 1.12)
        node.glow.scale.set(basePulse * 1.12, basePulse * 1.12, basePulse * 1.12)
      } else {
        if (glowMat && 'opacity' in glowMat) glowMat.opacity = 0.18
        if (meshMat && 'color' in meshMat && meshMat.color instanceof THREE.Color) meshMat.color.set(0x60a5fa)
        node.mesh.scale.set(basePulse, basePulse, basePulse)
        node.glow.scale.set(basePulse, basePulse, basePulse)
      }
    })
    // Optionally, highlight lines connected to hovered node
    activeConnections.forEach((conn, idx) => {
      let lineMat = conn.line.material
      if (Array.isArray(lineMat)) lineMat = lineMat[0]
      if (lineMat && 'opacity' in lineMat) {
        if (hoveredNodeIdx.value !== null && (conn.i === hoveredNodeIdx.value || conn.j === hoveredNodeIdx.value)) {
          lineMat.opacity = 0.45
          if (lineMat && 'color' in lineMat && lineMat.color instanceof THREE.Color) lineMat.color.set(0x8ecfff)
        } else {
          lineMat.opacity = 0.22
          if (lineMat && 'color' in lineMat && lineMat.color instanceof THREE.Color) lineMat.color.set(0x60a5fa)
        }
      }
    })
    // Pulse effect on click
    if (pulse > 0 && hoveredNodeIdx.value !== null) {
      const elapsed = (performance.now() - pulseStart) / 400
      const scale = 1 + Math.sin(Math.PI * Math.min(elapsed, 1)) * 0.7
      nodes[hoveredNodeIdx.value].mesh.scale.set(scale, scale, scale)
      nodes[hoveredNodeIdx.value].glow.scale.set(scale, scale, scale)
      if (elapsed >= 1) {
        pulse = 0
        nodes[hoveredNodeIdx.value].mesh.scale.set(1, 1, 1)
        nodes[hoveredNodeIdx.value].glow.scale.set(1, 1, 1)
      }
    }
    renderer!.render(scene, camera)
  }
  animate()
})

onBeforeUnmount(() => {
  if (animationId) cancelAnimationFrame(animationId)
  if (renderer) {
    renderer.dispose()
    renderer = null
  }
})
</script>

<style scoped>
.canvas-container {
  height: 600px;
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

.subtitle {
  font-size: 1.1rem;
}

@media (max-width: 960px) {
  .canvas-container {
    height: 400px;
  }
  
  .feature-animation-glow {
    width: 100%;
    height: 100%;
  }
}

@media (max-width: 600px) {
  .canvas-container {
    height: 320px;
  }
}
</style> 