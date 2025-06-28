<template>
  <section class="feature-panel gpu-on-demand">
    <div class="feature-content-row">
      <div class="feature-animation-with-glow">
        <div class="feature-animation-glow"></div>
        <div class="feature-animation">
          <canvas ref="threeCanvas" class="three-canvas" @mousemove="onCanvasMouseMove" @mouseleave="onCanvasMouseLeave">
            <div class="gpu-label">GPU</div>
          </canvas>
          <div v-if="hoveredNode" class="node-label" :style="{ left: hoveredNode.pos.x + 'px', top: hoveredNode.pos.y + 'px' }">
            {{ hoveredNode.type }}
          </div>
        </div>
      </div>
      <div class="feature-content feature-content-overlay">
        <h2 class="feature-title">GPU on Demand</h2>
        <p class="feature-description">
          Get instant access to powerful computing for your projects. Run demanding apps easily and scale up when you need more power. Pay only for what you use.
        </p>
        <div class="feature-benefits">
          <v-chip class="ma-1" color="white" variant="outlined" size="small">Instant Access</v-chip>
          <v-chip class="ma-1" color="white" variant="outlined" size="small">Full Passthrough</v-chip>
          <v-chip class="ma-1" color="white" variant="outlined" size="small">Pay-per-Use</v-chip>
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

// Parameters - adjusted to be consistent with other animations
const K8S_COUNT = 5
const PODS_PER_K8S = 4
const GPU_COUNT = 3 // Number of nodes that have GPUs
const GPU_COLOR = 0x60a5fa
const K8S_COLOR = 0x8ecfff
const POD_COLOR = 0x34d399
const PULSE_COLOR = 0xfff200
const CIRCUIT_COLOR = 0x38bdf8
const PARTICLE_COLOR = 0x8ecfff

const k8sNodes: { mesh: THREE.Mesh, glow: THREE.Mesh, basePos: THREE.Vector3, pods: { mesh: THREE.Mesh, glow: THREE.Mesh, basePos: THREE.Vector3 }[], gpu?: THREE.Mesh, gpuGlow?: THREE.Mesh }[] = []
let circuitLines: THREE.Line[] = []
let particles: THREE.Points | null = null
let pulses: { mesh: THREE.Mesh, path: THREE.Vector3[], t: number, active: boolean }[] = []
let pulseCooldown = 0
const connectionLines: THREE.Line[] = []
const podConnectionLines: THREE.Line[] = []

// Add hover state for node labels
const hoveredNode = shallowRef<{ mesh: THREE.Mesh, type: string, pos: { x: number, y: number } } | null>(null)

// Add mouse event handlers for hover detection
function onCanvasMouseMove(event: MouseEvent) {
  if (!threeCanvas.value || !renderer || !camera) return
  const rect = threeCanvas.value.getBoundingClientRect()
  const mouse = new THREE.Vector2(
    ((event.clientX - rect.left) / rect.width) * 2 - 1,
    -((event.clientY - rect.top) / rect.height) * 2 + 1
  )
  
  // Raycast against all nodes
  const allMeshes: { mesh: THREE.Mesh, type: string }[] = []
  
  k8sNodes.forEach((k8s, idx) => {
    allMeshes.push({ mesh: k8s.mesh, type: 'Kubernetes Node' })
    allMeshes.push({ mesh: k8s.glow, type: 'Kubernetes Node' })
    if (k8s.gpu) {
      allMeshes.push({ mesh: k8s.gpu, type: 'GPU Chip' })
      allMeshes.push({ mesh: k8s.gpuGlow!, type: 'GPU Chip' })
    }
    k8s.pods.forEach(pod => {
      allMeshes.push({ mesh: pod.mesh, type: 'Application Pod' })
      allMeshes.push({ mesh: pod.glow, type: 'Application Pod' })
    })
  })
  
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
  renderer = new THREE.WebGLRenderer({ canvas: threeCanvas.value, alpha: true, antialias: true })
  renderer.setSize(threeCanvas.value.clientWidth, threeCanvas.value.clientHeight, false)
  renderer.setClearColor(0x000000, 0)
  scene = new THREE.Scene()
  camera = new THREE.PerspectiveCamera(
    60,
    threeCanvas.value.clientWidth / threeCanvas.value.clientHeight,
    0.1,
    1000
  )
  camera.position.z = 6

  // K8s nodes in a ring - resized to be consistent
  const K8S_RADIUS = 2.1
  for (let i = 0; i < K8S_COUNT; i++) {
    const angle = (i / K8S_COUNT) * Math.PI * 2
    const basePos = new THREE.Vector3(
      K8S_RADIUS * Math.cos(angle),
      K8S_RADIUS * Math.sin(angle),
      0
    )
    const mesh = new THREE.Mesh(
      new THREE.SphereGeometry(0.12, 18, 18), // Reduced from 0.17, 20, 20
      new THREE.MeshBasicMaterial({ color: K8S_COLOR })
    )
    mesh.position.copy(basePos)
    scene.add(mesh)
    const glow = new THREE.Mesh(
      new THREE.SphereGeometry(0.21, 18, 18), // Reduced from 0.28, 20, 20
      new THREE.MeshBasicMaterial({ color: K8S_COLOR, transparent: true, opacity: 0.22 })
    )
    glow.position.copy(basePos)
    scene.add(glow)
    
    // Add GPU chip to some nodes (GPU-enabled nodes)
    let gpu: THREE.Mesh | undefined
    let gpuGlow: THREE.Mesh | undefined
    if (i < GPU_COUNT) {
      gpu = new THREE.Mesh(
        new THREE.BoxGeometry(0.18, 0.11, 0.04), // Reduced from 0.25, 0.15, 0.05
        new THREE.MeshBasicMaterial({ color: GPU_COLOR })
      )
      // Position GPU chip on the node
      gpu.position.set(0.15, 0, 0.06) // Adjusted position for smaller size
      mesh.add(gpu) // Attach to node
      
      gpuGlow = new THREE.Mesh(
        new THREE.BoxGeometry(0.25, 0.16, 0.06), // Reduced from 0.35, 0.22, 0.08
        new THREE.MeshBasicMaterial({ color: GPU_COLOR, transparent: true, opacity: 0.32 })
      )
      gpuGlow.position.set(0.15, 0, 0.06) // Adjusted position for smaller size
      mesh.add(gpuGlow)
      
      // Add circuit lines to GPU chip - adjusted for smaller size
      for (let j = 0; j < 3; j++) { // Reduced from 4 lines to 3
        const y = -0.04 + j * 0.03 // Adjusted for smaller GPU
        const points = [
          new THREE.Vector3(-0.07, y, 0.02), // Adjusted coordinates for smaller size
          new THREE.Vector3(-0.035, y, 0.025),
          new THREE.Vector3(0.035, y, 0.025),
          new THREE.Vector3(0.07, y, 0.02)
        ]
        const geom = new THREE.BufferGeometry().setFromPoints(points)
        const mat = new THREE.LineBasicMaterial({ color: CIRCUIT_COLOR, transparent: true, opacity: 0.5 })
        const line = new THREE.Line(geom, mat)
        gpu.add(line)
        circuitLines.push(line)
      }
    }
    
    // Pods for this K8s node (in an arc) - resized to be consistent
    const pods: { mesh: THREE.Mesh, glow: THREE.Mesh, basePos: THREE.Vector3 }[] = []
    const POD_RADIUS = 1.1
    for (let j = 0; j < PODS_PER_K8S; j++) {
      const arcSpread = Math.PI / 2
      const podAngle = angle - arcSpread / 2 + (j / (PODS_PER_K8S - 1)) * arcSpread
      const podPos = new THREE.Vector3(
        basePos.x + POD_RADIUS * Math.cos(podAngle),
        basePos.y + POD_RADIUS * Math.sin(podAngle),
        0
      )
      const podMesh = new THREE.Mesh(
        new THREE.BoxGeometry(0.16, 0.09, 0.05), // Reduced from 0.22, 0.12, 0.07
        new THREE.MeshBasicMaterial({ color: POD_COLOR, transparent: true, opacity: 1 })
      )
      podMesh.position.copy(podPos)
      scene.add(podMesh)
      const podGlow = new THREE.Mesh(
        new THREE.BoxGeometry(0.2, 0.12, 0.07), // Reduced from 0.28, 0.16, 0.09
        new THREE.MeshBasicMaterial({ color: POD_COLOR, transparent: true, opacity: 0.22 })
      )
      podGlow.position.copy(podPos)
      scene.add(podGlow)
      
      // Connection line from K8s to pod
      const podConnGeom = new THREE.BufferGeometry().setFromPoints([mesh.position, podMesh.position])
      const podConnMat = new THREE.LineBasicMaterial({ color: POD_COLOR, transparent: true, opacity: 0.18 })
      const podConnLine = new THREE.Line(podConnGeom, podConnMat)
      scene.add(podConnLine)
      podConnectionLines.push(podConnLine)
      pods.push({ mesh: podMesh, glow: podGlow, basePos: podPos })
    }
    k8sNodes.push({ mesh, glow, basePos, pods, gpu, gpuGlow })
  }

  // Particle field around GPU-enabled nodes
  const PARTICLE_COUNT = 60
  const particleGeom = new THREE.BufferGeometry()
  const positions = new Float32Array(PARTICLE_COUNT * 3)
  for (let i = 0; i < PARTICLE_COUNT; i++) {
    const r = 1.8 + Math.random() * 1.0
    const theta = Math.random() * Math.PI * 2
    const phi = Math.acos(2 * Math.random() - 1)
    positions[i * 3] = r * Math.sin(phi) * Math.cos(theta)
    positions[i * 3 + 1] = r * Math.sin(phi) * Math.sin(theta)
    positions[i * 3 + 2] = r * Math.cos(phi)
  }
  particleGeom.setAttribute('position', new THREE.BufferAttribute(positions, 3))
  const particleMat = new THREE.PointsMaterial({ color: PARTICLE_COLOR, size: 0.04, transparent: true, opacity: 0.22 })
  particles = new THREE.Points(particleGeom, particleMat)
  scene.add(particles)

  // Animate
  function animate(time = 0) {
    animationId = requestAnimationFrame(animate)
    const t = time * 0.001
    
    // Animate circuit lines on GPU chips
    circuitLines.forEach((line, idx) => {
      let mat = line.material
      if (Array.isArray(mat)) mat = mat[0]
      if (mat && 'opacity' in mat) mat.opacity = 0.35 + 0.25 * Math.sin(t * 1.3 + idx)
    })
    
    // Animate K8s nodes with consistent motion
    k8sNodes.forEach((k8s, i) => {
      const phase = t + i
      k8s.mesh.position.x = k8s.basePos.x + Math.sin(phase * 0.7 + i * 0.3) * 0.06
      k8s.mesh.position.y = k8s.basePos.y + Math.cos(phase * 0.9 + i * 0.4) * 0.06
      k8s.mesh.position.z = k8s.basePos.z + Math.sin(phase * 0.5 + i * 0.2) * 0.04
      k8s.glow.position.copy(k8s.mesh.position)
      
      // Animate GPU glow if present
      if (k8s.gpuGlow) {
        k8s.gpuGlow.scale.set(1 + 0.13 * Math.sin(t * 1.3), 1 + 0.13 * Math.sin(t * 1.3), 1)
      }
      
      // Animate pods with consistent motion
      k8s.pods.forEach((pod, j) => {
        const podPhase = t + i * 0.5 + j
        pod.mesh.position.x = pod.basePos.x + Math.sin(podPhase * 0.7 + j * 0.2) * 0.04
        pod.mesh.position.y = pod.basePos.y + Math.cos(podPhase * 0.9 + j * 0.3) * 0.04
        pod.mesh.position.z = pod.basePos.z + Math.sin(podPhase * 0.5 + j * 0.1) * 0.03
        pod.glow.position.copy(pod.mesh.position)
      })
    })
    
    // Animate connection lines
    let podLineIdx = 0
    k8sNodes.forEach((k8s, i) => {
      k8s.pods.forEach((pod, j) => {
        podConnectionLines[podLineIdx].geometry.setFromPoints([
          k8s.mesh.position,
          pod.mesh.position
        ])
        podLineIdx++
      })
    })
    
    // Animate particles
    if (particles) {
      const pos = particles.geometry.getAttribute('position')
      for (let i = 0; i < pos.count; i++) {
        pos.setZ(i, pos.getZ(i) + Math.sin(t * 0.7 + i) * 0.002)
      }
      pos.needsUpdate = true
    }
    
    // Animate pulses - now from GPU-enabled nodes to their pods
    pulseCooldown--
    if (pulseCooldown <= 0) {
      // Find a GPU-enabled node
      const gpuNodes = k8sNodes.filter(node => node.gpu)
      if (gpuNodes.length > 0) {
        const gpuNode = gpuNodes[Math.floor(Math.random() * gpuNodes.length)]
        const podIdx = Math.floor(Math.random() * gpuNode.pods.length)
        const path = [
          gpuNode.mesh.position.clone(),
          gpuNode.pods[podIdx].mesh.position.clone()
        ]
        pulses.push({ mesh: createPulseMesh(), path, t: 0, active: true })
        pulseCooldown = 30 + Math.random() * 40
      }
    }
    
    pulses.forEach((pulse, idx) => {
      if (!pulse.active) return
      pulse.t += 0.012
      const pos = new THREE.Vector3().lerpVectors(pulse.path[0], pulse.path[1], pulse.t)
      pulse.mesh.position.copy(pos)
      let mat = pulse.mesh.material
      if (Array.isArray(mat)) mat = mat[0]
      if (mat && 'opacity' in mat) mat.opacity = 0.7 + 0.25 * Math.sin(Math.PI * pulse.t)
      if (pulse.t >= 1) {
        pulse.active = false
        scene!.remove(pulse.mesh)
      }
    })
    pulses = pulses.filter(p => p.active)
    renderer!.render(scene!, camera!)
  }
  animate()
})

function createPulseMesh() {
  const pulseGeom = new THREE.SphereGeometry(0.05, 14, 14) // Reduced from 0.06, 16, 16
  const pulseMat = new THREE.MeshBasicMaterial({ color: PULSE_COLOR, transparent: true, opacity: 0.95 })
  const mesh = new THREE.Mesh(pulseGeom, pulseMat)
  scene!.add(mesh)
  return mesh
}

function createTextLabel(text: string, size: number = 0.13): THREE.Sprite {
  const canvas = document.createElement('canvas')
  canvas.width = 256
  canvas.height = 64
  const ctx = canvas.getContext('2d')!
  ctx.font = 'bold 32px Arial'
  ctx.fillStyle = '#ffffff'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.shadowColor = '#60a5fa'
  ctx.shadowBlur = 8
  ctx.fillText(text, 128, 32)
  const texture = new THREE.CanvasTexture(canvas)
  const mat = new THREE.SpriteMaterial({ map: texture, transparent: true })
  const sprite = new THREE.Sprite(mat)
  sprite.scale.set(size * 3, size, 1)
  return sprite
}

onBeforeUnmount(() => {
  if (animationId) cancelAnimationFrame(animationId)
  if (renderer) {
    renderer.dispose()
    renderer = null
  }
})
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
    height: 180px;
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
.gpu-label {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  color: #fff;
  font-size: 1.2rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  pointer-events: none;
  text-shadow: 0 2px 12px #60a5fa, 0 0px 2px #000;
  z-index: 10;
  user-select: none;
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
