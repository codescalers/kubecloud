<template>
  <FeatureSectionLayout
    title="Effortless Load Balancing & Scaling"
    description="Mycelium Cloud automatically balances traffic and scales your services up or down based on demand. Enjoy high availability and optimal performance with zero manual intervention."
    :tags="['Auto-scaling', 'Built-in load balancing', 'High availability']"
    reversed
    :hovered-node="hoveredNode ? { label: hoveredNode.type, pos: hoveredNode.pos } : undefined"
    #="{ canvasProps }"
  >
    <canvas
      v-bind="canvasProps"
      ref="threeCanvas"
      @mousemove="onCanvasMouseMove"
      @mouseleave="onCanvasMouseLeave"
    />
  </FeatureSectionLayout>
</template>

<script setup lang="ts">
import type { Line, Mesh, MeshBasicMaterial, PerspectiveCamera, Scene, Vector3, WebGLRenderer } from "three"

const THREE = ensureThreeGlobal()

const threeCanvas = ref<HTMLCanvasElement | null>(null)
let renderer: WebGLRenderer | null = null
let animationId: number | null = null
let scene: Scene | null = null
let camera: PerspectiveCamera | null = null

const CLIENT_COUNT = 5
const SERVICE_MIN = 3
const SERVICE_MAX = 8
const LOAD_BALANCER_COLOR = 0x60A5FA
const CLIENT_COLOR = 0x8ECFFF
const SERVICE_COLOR = 0x34D399
const TRAFFIC_COLOR = 0xFBBF24

const clients: { mesh: Mesh, glow: Mesh, basePos: Vector3, phase: number }[] = []
const services: { mesh: Mesh, glow: Mesh, basePos: Vector3, phase: number, active: boolean, scale: number, opacity: number }[] = []
let loadBalancer: { mesh: Mesh, glow: Mesh, basePos: Vector3 } | null = null
const clientLines: Line[] = []
const serviceLines: Line[] = []
let trafficPulses: { mesh: Mesh, from: Vector3, to: Vector3, t: number, active: boolean }[] = []
let trafficCooldown = 0

// Globe-compatible animation states
let rotationY = 0
let scalingPulses: { mesh: Mesh, from: Vector3, to: Vector3, t: number, active: boolean }[] = []
let scalingCooldown = 0

const hoveredNode = shallowRef<{ mesh: Mesh, type: string, pos: { x: number, y: number } } | null>(null)

function onCanvasMouseMove(event: MouseEvent) {
  if (!threeCanvas.value || !renderer || !camera)
    return
  const rect = threeCanvas.value.getBoundingClientRect()
  const mouse = new THREE.Vector2(
    ((event.clientX - rect.left) / rect.width) * 2 - 1,
    -((event.clientY - rect.top) / rect.height) * 2 + 1,
  )
  const allMeshes = [
    ...clients.map(c => ({ mesh: c.mesh, type: "Client" })),
    ...(loadBalancer ? [{ mesh: loadBalancer.mesh, type: "Load Balancer" }] : []),
    ...services.filter(s => s.active && s.opacity > 0.7).map(s => ({ mesh: s.mesh, type: "Service" })),
  ]
  const raycaster = new THREE.Raycaster()
  raycaster.setFromCamera(mouse, camera)
  const intersects = raycaster.intersectObjects(allMeshes.map(n => n.mesh))
  const intersect = intersects[0]
  if (intersect) {
    const found = allMeshes.find(n => n.mesh === intersect.object)
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
  if (!threeCanvas.value)
    return
  renderer = new THREE.WebGLRenderer({
    canvas: threeCanvas.value,
    alpha: true,
    antialias: true,
    powerPreference: "high-performance",
  })
  renderer.setSize(threeCanvas.value.clientWidth, threeCanvas.value.clientHeight, false)
  renderer.setClearColor(0x000000, 0)
  scene = new THREE.Scene()
  camera = new THREE.PerspectiveCamera(
    60,
    threeCanvas.value.clientWidth / threeCanvas.value.clientHeight,
    0.1,
    1000,
  )
  camera.position.z = 5.5

  // Place load balancer in center FIRST
  const lbPos = new THREE.Vector3(0, 0, 0)
  const lbMesh = new THREE.Mesh(
    new THREE.BoxGeometry(0.6, 0.32, 0.08),
    new THREE.MeshBasicMaterial({ color: LOAD_BALANCER_COLOR }),
  )
  lbMesh.position.copy(lbPos)
  scene.add(lbMesh)
  const lbGlow = new THREE.Mesh(
    new THREE.BoxGeometry(0.8, 0.45, 0.12),
    new THREE.MeshBasicMaterial({ color: LOAD_BALANCER_COLOR, transparent: true, opacity: 0.18 }),
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
      new THREE.MeshBasicMaterial({ color: CLIENT_COLOR }),
    )
    mesh.position.copy(basePos)
    scene.add(mesh)
    const glow = new THREE.Mesh(
      new THREE.SphereGeometry(0.19, 18, 18),
      new THREE.MeshBasicMaterial({ color: CLIENT_COLOR, transparent: true, opacity: 0.18 }),
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
      new THREE.MeshBasicMaterial({ color: SERVICE_COLOR, transparent: true, opacity: 0 }),
    )
    mesh.position.copy(basePos)
    scene.add(mesh)
    const glow = new THREE.Mesh(
      new THREE.BoxGeometry(0.2, 0.1, 0.05),
      new THREE.MeshBasicMaterial({ color: SERVICE_COLOR, transparent: true, opacity: 0 }),
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

    // Auto-rotate like globe
    rotationY += 0.001

    // Animate clients with globe-like floating motion
    clients.forEach((client, idx) => {
      client.mesh.position.x = client.basePos.x + Math.sin(t * 0.7 + client.phase + idx * 0.3) * 0.06
      client.mesh.position.y = client.basePos.y + Math.cos(t * 0.9 + client.phase + idx * 0.4) * 0.06
      client.mesh.position.z = client.basePos.z + Math.sin(t * 0.5 + client.phase + idx * 0.2) * 0.04
      client.glow.position.copy(client.mesh.position)
      clientLines[idx]?.geometry.setFromPoints([client.mesh.position, loadBalancer!.mesh.position])

      // Rotate around center like globe
      client.mesh.rotation.y = rotationY * 0.3 + idx * 0.2
    })

    // Animate load balancer with enhanced effects
    if (loadBalancer) {
      loadBalancer.mesh.position.x = loadBalancer.basePos.x + Math.sin(t * 0.6) * 0.04
      loadBalancer.mesh.position.y = loadBalancer.basePos.y + Math.cos(t * 0.8) * 0.04
      loadBalancer.mesh.position.z = loadBalancer.basePos.z + Math.sin(t * 0.5) * 0.03
      loadBalancer.glow.position.copy(loadBalancer.mesh.position)

      // Rotate around center
      loadBalancer.mesh.rotation.y = rotationY * 0.5
    }

    // Generate scaling pulses
    scalingCooldown--
    if (scalingCooldown <= 0 && Math.random() < 0.015) {
      const activeServices = services.filter(s => s.active && s.opacity > 0.8)
      if (activeServices.length > 0) {
        const service = activeServices[Math.floor(Math.random() * activeServices.length)]
        if (!service) {
          return
        }

        const pulse = createScalingPulse(loadBalancer!.mesh.position, service.mesh.position)
        if (pulse) {
          scalingPulses.push(pulse)
          scalingCooldown = 50
        }
      }
    }

    // Animate scaling pulses
    scalingPulses = scalingPulses.filter((pulse) => {
      pulse.t += 0.006
      if (pulse.t >= 1) {
        if (scene) {
          scene.remove(pulse.mesh)
        }
        return false
      }
      pulse.mesh.position.lerpVectors(pulse.from, pulse.to, pulse.t)
      const material = pulse.mesh.material as MeshBasicMaterial
      material.opacity = 0.7 * (1 - pulse.t)
      return true
    })

    // Dynamic scaling: randomly activate/deactivate service nodes
    if (Math.random() < 0.012) {
      const actives = services.filter(s => s.active)
      const inactives = services.filter(s => !s.active)
      if (actives.length < SERVICE_MAX && Math.random() > 0.3) {
        // Scale up
        const s = inactives[Math.floor(Math.random() * inactives.length)]
        if (s)
          s.active = true
      }
      else if (actives.length > SERVICE_MIN) {
        // Scale down
        const s = actives[Math.floor(Math.random() * actives.length)]
        if (s)
          s.active = false
      }
    }

    // Animate service fade/scale with enhanced effects
    services.forEach((service, idx) => {
      const t2 = t + service.phase
      if (service.active && service.opacity < 1) {
        service.opacity += 0.04
        if (service.opacity > 1)
          service.opacity = 1
      }
      else if (!service.active && service.opacity > 0) {
        service.opacity -= 0.04
        if (service.opacity < 0)
          service.opacity = 0
      }
      // Animate scale
      if (service.active && service.scale < 1)
        service.scale += 0.04
      if (!service.active && service.scale > 0.7)
        service.scale -= 0.04
      if (service.scale < 0.7)
        service.scale = 0.7
      if (service.scale > 1)
        service.scale = 1
      service.mesh.scale.set(service.scale, service.scale, service.scale)
      service.glow.scale.set(service.scale, service.scale, service.scale)

      // Animate position with globe-like motion
      service.mesh.position.x = service.basePos.x + Math.sin(t2 * 0.7 + idx * 0.2) * 0.05
      service.mesh.position.y = service.basePos.y + Math.cos(t2 * 0.9 + idx * 0.3) * 0.05
      service.mesh.position.z = service.basePos.z + Math.sin(t2 * 0.5 + idx * 0.1) * 0.03
      service.glow.position.copy(service.mesh.position)

      // Rotate around center
      service.mesh.rotation.y = rotationY * 0.4 + idx * 0.6

      // Update material opacity with enhanced effects
      let mat = service.mesh.material
      if (Array.isArray(mat))
        mat = mat[0] ?? []
      if (mat && "opacity" in mat)
        mat.opacity = service.opacity
      let glowMat = service.glow.material
      if (Array.isArray(glowMat))
        glowMat = glowMat[0] ?? []
      if (glowMat && "opacity" in glowMat) {
        glowMat.opacity = (0.18 + 0.05 * Math.sin(t * 2 + idx)) * service.opacity
      }

      // Update service connection line with enhanced opacity
      const sl = serviceLines[idx]
      if (!sl) {
        return
      }

      sl.geometry.setFromPoints([loadBalancer!.mesh.position, service.mesh.position])
      let lineMat = sl.material
      if (Array.isArray(lineMat))
        lineMat = lineMat[0] ?? []
      if (lineMat && "opacity" in lineMat) {
        lineMat.opacity = (0.18 + 0.06 * Math.sin(t * 2 + idx * 0.3)) * service.opacity
      }
    })

    // Animate traffic pulses: from random client to load balancer, then to random active service
    trafficCooldown--
    if (trafficCooldown <= 0) {
      const activeServices = services.filter(s => s.active && s.opacity > 0.8)
      if (activeServices.length > 0) {
        const clientIdx = Math.floor(Math.random() * clients.length)

        const idx = Math.floor(Math.random() * activeServices.length)
        if (activeServices.length < idx) {
          return
        }

        const service = activeServices[idx]
        if (!service) {
          return
        }

        const serviceIdx = services.indexOf(service)
        const c = clients[clientIdx]
        const s = services[serviceIdx]
        if (!c || !s) {
          return
        }

        const from = c.mesh.position.clone()
        const mid = loadBalancer!.mesh.position.clone()
        const to = s.mesh.position.clone()
        trafficPulses.push({ mesh: createPulseMesh(), from, to: mid, t: 0, active: true })
        trafficPulses.push({ mesh: createPulseMesh(), from: mid, to, t: 0, active: true })
        trafficCooldown = 35 + Math.random() * 40
      }
    }
    trafficPulses.forEach((pulse) => {
      if (!pulse.active)
        return
      pulse.t += 0.018
      pulse.mesh.position.lerpVectors(pulse.from, pulse.to, pulse.t)
      let mat = pulse.mesh.material
      if (Array.isArray(mat))
        mat = mat[0] ?? []
      if (mat && "opacity" in mat)
        mat.opacity = 0.5 + 0.45 * Math.sin(Math.PI * pulse.t)
      if (pulse.t >= 1) {
        pulse.active = false
        if (scene) {
          scene.remove(pulse.mesh)
        }
      }
    })
    trafficPulses = trafficPulses.filter(p => p.active)
    renderer!.render(scene!, camera!)
  }

  function createScalingPulse(from: Vector3, to: Vector3) {
    if (!scene)
      return null
    const geometry = new THREE.SphereGeometry(0.06, 12, 12)
    const material = new THREE.MeshBasicMaterial({
      color: SERVICE_COLOR,
      transparent: true,
      opacity: 0.7,
    })
    const mesh = new THREE.Mesh(geometry, material)
    mesh.position.copy(from)
    scene.add(mesh)
    return { mesh, from: from.clone(), to: to.clone(), t: 0, active: true }
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
  if (animationId)
    cancelAnimationFrame(animationId)
  if (renderer) {
    renderer.dispose()
    renderer = null
  }
})
</script>

<style scoped>
.subtitle{
  font-size: 1.1rem
}
</style>
