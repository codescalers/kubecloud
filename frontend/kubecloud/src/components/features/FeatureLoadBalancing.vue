<template>
  <section class="feature-panel load-balancing">
    <div class="feature-content-row">
      <div class="feature-content feature-content-overlay">
        <h2 class="feature-title">Effortless Load Balancing & Scaling</h2>
        <p class="feature-description">
          KubeCloud automatically balances traffic and scales your services up or down based on demand. Enjoy high availability and optimal performance with zero manual intervention.
        </p>
        <div class="feature-benefits">
          <v-chip class="ma-1" color="white" variant="outlined" size="small">Auto-scaling</v-chip>
          <v-chip class="ma-1" color="white" variant="outlined" size="small">Built-in load balancing</v-chip>
          <v-chip class="ma-1" color="white" variant="outlined" size="small">High availability</v-chip>
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

<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, shallowRef, computed } from 'vue'
import * as THREE from 'three'

const threeCanvas = ref<HTMLCanvasElement | null>(null)
let renderer: THREE.WebGLRenderer | null = null
let animationId: number | null = null
let scene: THREE.Scene | null = null
let camera: THREE.PerspectiveCamera | null = null

const CLIENT_COUNT = 5
const SERVICE_MIN = 3
const SERVICE_MAX = 8
const LOAD_BALANCER_COLOR = 0x60a5fa
const CLIENT_COLOR = 0x8ecfff
const SERVICE_COLOR = 0x34d399
const TRAFFIC_COLOR = 0xfbbf24

const clients: { mesh: THREE.Mesh, glow: THREE.Mesh, basePos: THREE.Vector3, phase: number }[] = []
const services: { mesh: THREE.Mesh, glow: THREE.Mesh, basePos: THREE.Vector3, phase: number, active: boolean, scale: number, opacity: number }[] = []
let loadBalancer: { mesh: THREE.Mesh, glow: THREE.Mesh, basePos: THREE.Vector3 } | null = null
const clientLines: THREE.Line[] = []
const serviceLines: THREE.Line[] = []
let trafficPulses: { mesh: THREE.Mesh, from: THREE.Vector3, to: THREE.Vector3, t: number, active: boolean }[] = []
let trafficCooldown = 0

const hoveredNode = shallowRef<{ mesh: THREE.Mesh, type: string, pos: { x: number, y: number } } | null>(null)

const nodeLabelStyle = computed(() => {
  if (!hoveredNode.value) return {}
  let color = '#60a5fa', borderColor = '#60a5fa33'
  if (hoveredNode.value.type === 'Client') { color = '#8ecfff'; borderColor = '#8ecfff33' }
  if (hoveredNode.value.type === 'Service') { color = '#34d399'; borderColor = '#34d39933' }
  return {
    left: hoveredNode.value.pos.x + 'px',
    top: hoveredNode.value.pos.y + 'px',
    color,
    borderColor
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
    ...clients.map(c => ({ mesh: c.mesh, type: 'Client' })),
    ...(loadBalancer ? [{ mesh: loadBalancer.mesh, type: 'Load Balancer' }] : []),
    ...services.filter(s => s.active && s.opacity > 0.7).map(s => ({ mesh: s.mesh, type: 'Service' }))
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

  // Place load balancer in center FIRST
  const lbPos = new THREE.Vector3(0, 0, 0)
  const lbMesh = new THREE.Mesh(
    new THREE.BoxGeometry(0.6, 0.32, 0.08),
    new THREE.MeshBasicMaterial({ color: LOAD_BALANCER_COLOR })
  )
  lbMesh.position.copy(lbPos)
  scene.add(lbMesh)
  const lbGlow = new THREE.Mesh(
    new THREE.BoxGeometry(0.8, 0.45, 0.12),
    new THREE.MeshBasicMaterial({ color: LOAD_BALANCER_COLOR, transparent: true, opacity: 0.18 })
  )
  lbGlow.position.copy(lbPos)
  scene.add(lbGlow)
  loadBalancer = { mesh: lbMesh, glow: lbGlow, basePos: lbPos }
  // Place clients in a tighter arc above
  for (let i = 0; i < CLIENT_COUNT; i++) {
    const angle = -Math.PI / 4 + (i / (CLIENT_COUNT - 1)) * (Math.PI / 2)
    const r = 2.8
    const x = r * Math.cos(angle)
    const y = r * Math.sin(angle)
    const basePos = new THREE.Vector3(x, y, 0)
    const mesh = new THREE.Mesh(
      new THREE.SphereGeometry(0.11, 18, 18),
      new THREE.MeshBasicMaterial({ color: CLIENT_COLOR })
    )
    mesh.position.copy(basePos)
    scene.add(mesh)
    const glow = new THREE.Mesh(
      new THREE.SphereGeometry(0.19, 18, 18),
      new THREE.MeshBasicMaterial({ color: CLIENT_COLOR, transparent: true, opacity: 0.18 })
    )
    glow.position.copy(basePos)
    clients.push({ mesh, glow, basePos, phase: Math.random() * Math.PI * 2 })
    // Line to load balancer
    const lineGeom = new THREE.BufferGeometry().setFromPoints([mesh.position, loadBalancer!.mesh.position])
    const lineMat = new THREE.LineBasicMaterial({ color: CLIENT_COLOR, transparent: true, opacity: 0.13 })
    const line = new THREE.Line(lineGeom, lineMat)
    scene.add(line)
    clientLines.push(line)
  }
  // Place service nodes in a tighter arc below
  for (let i = 0; i < SERVICE_MAX; i++) {
    const angle = (3 * Math.PI) / 4 + (i / (SERVICE_MAX - 1)) * (Math.PI / 2)
    const r = 2.8
    const x = r * Math.cos(angle)
    const y = r * Math.sin(angle)
    const basePos = new THREE.Vector3(x, y, 0)
    const mesh = new THREE.Mesh(
      new THREE.BoxGeometry(0.2, 0.1, 0.05),
      new THREE.MeshBasicMaterial({ color: SERVICE_COLOR, transparent: true, opacity: 0 })
    )
    mesh.position.copy(basePos)
    scene.add(mesh)
    const glow = new THREE.Mesh(
      new THREE.BoxGeometry(0.2, 0.1, 0.05),
      new THREE.MeshBasicMaterial({ color: SERVICE_COLOR, transparent: true, opacity: 0 })
    )
    glow.position.copy(basePos)
    services.push({ mesh, glow, basePos, phase: Math.random() * Math.PI * 2, active: i < SERVICE_MIN, scale: 0.7, opacity: i < SERVICE_MIN ? 1 : 0 })
    // Line to load balancer
    const lineGeom = new THREE.BufferGeometry().setFromPoints([lbMesh.position, mesh.position])
    const lineMat = new THREE.LineBasicMaterial({ color: SERVICE_COLOR, transparent: true, opacity: 0 })
    const line = new THREE.Line(lineGeom, lineMat)
    scene.add(line)
    serviceLines.push(line)
  }
  // Animate
  function animate(time = 0) {
    animationId = requestAnimationFrame(animate)
    const t = time * 0.001
    // Animate clients
    clients.forEach((client, idx) => {
      client.mesh.position.x = client.basePos.x + Math.sin(t * 0.7 + client.phase + idx) * 0.07
      client.mesh.position.y = client.basePos.y + Math.cos(t * 0.9 + client.phase + idx * 1.2) * 0.07
      client.mesh.position.z = client.basePos.z + Math.sin(t * 0.5 + client.phase + idx * 0.7) * 0.05
      client.glow.position.copy(client.mesh.position)
      clientLines[idx].geometry.setFromPoints([client.mesh.position, loadBalancer!.mesh.position])
    })
    // Animate load balancer
    if (loadBalancer) {
      loadBalancer.mesh.position.x = loadBalancer.basePos.x + Math.sin(t * 0.6) * 0.04
      loadBalancer.mesh.position.y = loadBalancer.basePos.y + Math.cos(t * 0.8) * 0.04
      loadBalancer.mesh.position.z = loadBalancer.basePos.z + Math.sin(t * 0.5) * 0.03
      loadBalancer.glow.position.copy(loadBalancer.mesh.position)
    }
    // Dynamic scaling: randomly activate/deactivate service nodes
    if (Math.random() < 0.012) {
      const actives = services.filter(s => s.active)
      const inactives = services.filter(s => !s.active)
      if (actives.length < SERVICE_MAX && Math.random() > 0.3) {
        // Scale up
        const s = inactives[Math.floor(Math.random() * inactives.length)]
        if (s) s.active = true
      } else if (actives.length > SERVICE_MIN) {
        // Scale down
        const s = actives[Math.floor(Math.random() * actives.length)]
        if (s) s.active = false
      }
    }
    // Animate service fade/scale
    services.forEach((service, idx) => {
      const t2 = t + service.phase
      if (service.active && service.opacity < 1) {
        service.opacity += 0.04
        if (service.opacity > 1) service.opacity = 1
      } else if (!service.active && service.opacity > 0) {
        service.opacity -= 0.04
        if (service.opacity < 0) service.opacity = 0
      }
      // Animate scale
      if (service.active && service.scale < 1) service.scale += 0.04
      if (!service.active && service.scale > 0.7) service.scale -= 0.04
      if (service.scale < 0.7) service.scale = 0.7
      if (service.scale > 1) service.scale = 1
      service.mesh.scale.set(service.scale, service.scale, service.scale)
      service.glow.scale.set(service.scale, service.scale, service.scale)
      // Animate position
      service.mesh.position.x = service.basePos.x + Math.sin(t2 * 0.7 + idx) * 0.06
      service.mesh.position.y = service.basePos.y + Math.cos(t2 * 0.9 + idx * 1.2) * 0.06
      service.mesh.position.z = service.basePos.z + Math.sin(t2 * 0.5 + idx * 0.7) * 0.04
      service.glow.position.copy(service.mesh.position)
      // Update material opacity
      let mat = service.mesh.material
      if (Array.isArray(mat)) mat = mat[0]
      if (mat && 'opacity' in mat) mat.opacity = service.opacity
      let glowMat = service.glow.material
      if (Array.isArray(glowMat)) glowMat = glowMat[0]
      if (glowMat && 'opacity' in glowMat) glowMat.opacity = 0.18 * service.opacity
      // Update service connection line
      serviceLines[idx].geometry.setFromPoints([loadBalancer!.mesh.position, service.mesh.position])
      let lineMat = serviceLines[idx].material
      if (Array.isArray(lineMat)) lineMat = lineMat[0]
      if (lineMat && 'opacity' in lineMat) lineMat.opacity = 0.18 * service.opacity
    })
    // Animate traffic pulses: from random client to load balancer, then to random active service
    trafficCooldown--
    if (trafficCooldown <= 0) {
      const activeServices = services.filter(s => s.active && s.opacity > 0.8)
      if (activeServices.length > 0) {
        const clientIdx = Math.floor(Math.random() * clients.length)
        const serviceIdx = services.indexOf(activeServices[Math.floor(Math.random() * activeServices.length)])
        const from = clients[clientIdx].mesh.position.clone()
        const mid = loadBalancer!.mesh.position.clone()
        const to = services[serviceIdx].mesh.position.clone()
        trafficPulses.push({ mesh: createPulseMesh(), from, to: mid, t: 0, active: true })
        trafficPulses.push({ mesh: createPulseMesh(), from: mid, to, t: 0, active: true })
        trafficCooldown = 35 + Math.random() * 40
      }
    }
    trafficPulses.forEach((pulse, idx) => {
      if (!pulse.active) return
      pulse.t += 0.018
      pulse.mesh.position.lerpVectors(pulse.from, pulse.to, pulse.t)
      let mat = pulse.mesh.material
      if (Array.isArray(mat)) mat = mat[0]
      if (mat && 'opacity' in mat) mat.opacity = 0.5 + 0.45 * Math.sin(Math.PI * pulse.t)
      if (pulse.t >= 1) {
        pulse.active = false
        scene!.remove(pulse.mesh)
      }
    })
    trafficPulses = trafficPulses.filter(p => p.active)
    renderer!.render(scene!, camera!)
  }
  animate()
})

function createPulseMesh() {
  const pulseGeom = new THREE.SphereGeometry(0.08, 14, 14)
  const pulseMat = new THREE.MeshBasicMaterial({ color: TRAFFIC_COLOR, transparent: true, opacity: 0.92 })
  const mesh = new THREE.Mesh(pulseGeom, pulseMat)
  scene!.add(mesh)
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