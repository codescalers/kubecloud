<template>
  <section class="feature-panel gpu-on-demand">
    <div class="feature-content-row">
      <div class="feature-animation-with-glow">
        <div class="feature-animation-glow"></div>
        <div class="feature-animation">
          <canvas ref="threeCanvas" class="three-canvas">
            <div class="gpu-label">GPU</div>
          </canvas>
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
import { onMounted, onBeforeUnmount, ref } from 'vue'
import * as THREE from 'three'

const threeCanvas = ref<HTMLCanvasElement | null>(null)
let renderer: THREE.WebGLRenderer | null = null
let animationId: number | null = null
let scene: THREE.Scene | null = null
let camera: THREE.PerspectiveCamera | null = null

// Parameters
const K8S_COUNT = 5
const PODS_PER_K8S = 4
const GPU_COLOR = 0x60a5fa
const K8S_COLOR = 0x8ecfff
const POD_COLOR = 0x34d399
const PULSE_COLOR = 0xfff200
const CIRCUIT_COLOR = 0x38bdf8
const PARTICLE_COLOR = 0x8ecfff

const k8sNodes: { mesh: THREE.Mesh, glow: THREE.Mesh, basePos: THREE.Vector3, pods: { mesh: THREE.Mesh, glow: THREE.Mesh, basePos: THREE.Vector3 }[] }[] = []
let gpuChip: THREE.Mesh, gpuGlow: THREE.Mesh
let circuitLines: THREE.Line[] = []
let particles: THREE.Points | null = null
let pulses: { mesh: THREE.Mesh, path: THREE.Vector3[], t: number, active: boolean }[] = []
let pulseCooldown = 0
const connectionLines: THREE.Line[] = []
const podConnectionLines: THREE.Line[] = []

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

  // GPU chip in center
  gpuChip = new THREE.Mesh(
    new THREE.BoxGeometry(0.62, 0.36, 0.13),
    new THREE.MeshBasicMaterial({ color: GPU_COLOR })
  )
  gpuChip.position.set(0, 0, 0)
  scene.add(gpuChip)
  gpuGlow = new THREE.Mesh(
    new THREE.BoxGeometry(1.1, 0.65, 0.22),
    new THREE.MeshBasicMaterial({ color: GPU_COLOR, transparent: true, opacity: 0.32 })
  )
  gpuGlow.position.set(0, 0, 0)
  scene.add(gpuGlow)

  // Attach GPU label as child
  const gpuLabel = createTextLabel('GPU', 0.13)
  gpuLabel.position.set(0, -0.28, 0.09)
  gpuChip.add(gpuLabel)

  // Circuit lines on GPU chip
  for (let i = 0; i < 6; i++) {
    const y = -0.11 + i * 0.044
    const points = [
      new THREE.Vector3(-0.22, y, 0.06),
      new THREE.Vector3(-0.13, y, 0.09),
      new THREE.Vector3(0.13, y, 0.09),
      new THREE.Vector3(0.22, y, 0.06)
    ]
    const geom = new THREE.BufferGeometry().setFromPoints(points)
    const mat = new THREE.LineBasicMaterial({ color: CIRCUIT_COLOR, transparent: true, opacity: 0.5 })
    const line = new THREE.Line(geom, mat)
    gpuChip.add(line)
    circuitLines.push(line)
  }

  // K8s nodes in a ring
  const K8S_RADIUS = 2.1
  for (let i = 0; i < K8S_COUNT; i++) {
    const angle = (i / K8S_COUNT) * Math.PI * 2
    const basePos = new THREE.Vector3(
      K8S_RADIUS * Math.cos(angle),
      K8S_RADIUS * Math.sin(angle),
      0
    )
    const mesh = new THREE.Mesh(
      new THREE.SphereGeometry(0.17, 20, 20),
      new THREE.MeshBasicMaterial({ color: K8S_COLOR })
    )
    mesh.position.copy(basePos)
    scene.add(mesh)
    const glow = new THREE.Mesh(
      new THREE.SphereGeometry(0.28, 20, 20),
      new THREE.MeshBasicMaterial({ color: K8S_COLOR, transparent: true, opacity: 0.22 })
    )
    glow.position.copy(basePos)
    scene.add(glow)
    // Attach K8s label as child
    const k8sLabel = createTextLabel('K8s', 0.09)
    k8sLabel.position.set(0, -0.22, 0.13)
    mesh.add(k8sLabel)
    // Connection line from GPU to K8s
    const connGeom = new THREE.BufferGeometry().setFromPoints([gpuChip.position, mesh.position])
    const connMat = new THREE.LineBasicMaterial({ color: K8S_COLOR, transparent: true, opacity: 0.22 })
    const connLine = new THREE.Line(connGeom, connMat)
    scene.add(connLine)
    connectionLines.push(connLine)
    // Pods for this K8s node (in an arc)
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
        new THREE.BoxGeometry(0.22, 0.12, 0.07),
        new THREE.MeshBasicMaterial({ color: POD_COLOR, transparent: true, opacity: 1 })
      )
      podMesh.position.copy(podPos)
      scene.add(podMesh)
      const podGlow = new THREE.Mesh(
        new THREE.BoxGeometry(0.28, 0.16, 0.09),
        new THREE.MeshBasicMaterial({ color: POD_COLOR, transparent: true, opacity: 0.22 })
      )
      podGlow.position.copy(podPos)
      scene.add(podGlow)
      // Attach Pod label as child
      const podLabel = createTextLabel('Pod', 0.07)
      podLabel.position.set(0, -0.11, 0.06)
      podMesh.add(podLabel)
      // Connection line from K8s to pod
      const podConnGeom = new THREE.BufferGeometry().setFromPoints([mesh.position, podMesh.position])
      const podConnMat = new THREE.LineBasicMaterial({ color: POD_COLOR, transparent: true, opacity: 0.18 })
      const podConnLine = new THREE.Line(podConnGeom, podConnMat)
      scene.add(podConnLine)
      podConnectionLines.push(podConnLine)
      pods.push({ mesh: podMesh, glow: podGlow, basePos: podPos })
    }
    k8sNodes.push({ mesh, glow, basePos, pods })
  }

  // Particle field around GPU
  const PARTICLE_COUNT = 80
  const particleGeom = new THREE.BufferGeometry()
  const positions = new Float32Array(PARTICLE_COUNT * 3)
  for (let i = 0; i < PARTICLE_COUNT; i++) {
    const r = 1.1 + Math.random() * 1.2
    const theta = Math.random() * Math.PI * 2
    const phi = Math.acos(2 * Math.random() - 1)
    positions[i * 3] = r * Math.sin(phi) * Math.cos(theta)
    positions[i * 3 + 1] = r * Math.sin(phi) * Math.sin(theta)
    positions[i * 3 + 2] = r * Math.cos(phi)
  }
  particleGeom.setAttribute('position', new THREE.BufferAttribute(positions, 3))
  const particleMat = new THREE.PointsMaterial({ color: PARTICLE_COLOR, size: 0.06, transparent: true, opacity: 0.22 })
  particles = new THREE.Points(particleGeom, particleMat)
  scene.add(particles)

  // Animate
  function animate(time = 0) {
    animationId = requestAnimationFrame(animate)
    const t = time * 0.001
    // Animate GPU glow
    gpuGlow.scale.set(1 + 0.13 * Math.sin(t * 1.3), 1 + 0.13 * Math.sin(t * 1.3), 1)
    // Animate circuit lines
    circuitLines.forEach((line, idx) => {
      let mat = line.material
      if (Array.isArray(mat)) mat = mat[0]
      if (mat && 'opacity' in mat) mat.opacity = 0.35 + 0.25 * Math.sin(t * 1.3 + idx)
    })
    // Animate K8s nodes
    k8sNodes.forEach((k8s, i) => {
      const phase = t + i
      k8s.mesh.position.x = k8s.basePos.x + Math.sin(phase * 0.45) * 0.09
      k8s.mesh.position.y = k8s.basePos.y + Math.cos(phase * 0.55) * 0.09
      k8s.mesh.position.z = k8s.basePos.z + Math.sin(phase * 0.3) * 0.07
      k8s.glow.position.copy(k8s.mesh.position)
      // Animate pods
      k8s.pods.forEach((pod, j) => {
        const podPhase = t + i * 0.5 + j
        pod.mesh.position.x = pod.basePos.x + Math.sin(podPhase * 0.45) * 0.06
        pod.mesh.position.y = pod.basePos.y + Math.cos(podPhase * 0.55) * 0.06
        pod.mesh.position.z = pod.basePos.z + Math.sin(podPhase * 0.3) * 0.04
        pod.glow.position.copy(pod.mesh.position)
      })
    })
    // Animate connection lines
    connectionLines.forEach((line, i) => {
      line.geometry.setFromPoints([
        gpuChip.position,
        k8sNodes[i].mesh.position
      ])
    })
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
    // Animate pulses
    pulseCooldown--
    if (pulseCooldown <= 0) {
      // Launch a new pulse from GPU to random K8s node, then to a random pod
      const k8sIdx = Math.floor(Math.random() * K8S_COUNT)
      const podIdx = Math.floor(Math.random() * PODS_PER_K8S)
      const path = [
        gpuChip.position.clone(),
        k8sNodes[k8sIdx].mesh.position.clone(),
        k8sNodes[k8sIdx].pods[podIdx].mesh.position.clone()
      ]
      pulses.push({ mesh: createPulseMesh(), path, t: 0, active: true })
      pulseCooldown = 30 + Math.random() * 40
    }
    pulses.forEach((pulse, idx) => {
      if (!pulse.active) return
      pulse.t += 0.012
      let pos: THREE.Vector3
      if (pulse.t < 0.5) {
        pos = new THREE.Vector3().lerpVectors(pulse.path[0], pulse.path[1], pulse.t * 2)
      } else {
        pos = new THREE.Vector3().lerpVectors(pulse.path[1], pulse.path[2], (pulse.t - 0.5) * 2)
      }
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
  const pulseGeom = new THREE.SphereGeometry(0.06, 16, 16)
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
</style>
