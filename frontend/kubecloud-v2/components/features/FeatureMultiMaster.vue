<template>
  <FeatureSectionLayout
    title="Multi-Master Clusters"
    description="High-availability Kubernetes clusters with multiple control plane nodes. Automatic failover, leader election, and zero-downtime upgrades built-in."
    :tags="['HA Control Plane', 'Automatic Failover', 'Zero-downtime Upgrades']"
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
import * as THREE from "three"
import { onBeforeUnmount, onMounted, ref, shallowRef } from "vue"

const threeCanvas = ref<HTMLCanvasElement | null>(null)
let renderer: THREE.WebGLRenderer | null = null
let animationId: number | null = null
let scene: THREE.Scene | null = null
let camera: THREE.PerspectiveCamera | null = null

const MASTER_COUNT = 5
const WORKER_COUNT = 7
const PRIMARY_COLOR = 0x60A5FA
const MASTER_COLOR = 0x8ECFFF // More sophisticated light blue instead of cartoonish yellow
const LEADER_COLOR = 0x60A5FA // Use primary blue for leader instead of bright yellow
const WORKER_COLOR = 0x34D399
const CONNECTION_COLOR = 0x38BDF8 // Cyan for connections

const MASTER_RING_RADIUS = 1.2
const WORKER_RING_RADIUS = 2.0

const masters: { mesh: THREE.Mesh, glow: THREE.Mesh, basePos: THREE.Vector3, phase: number, crown?: THREE.Mesh }[] = []
const workers: { mesh: THREE.Mesh, glow: THREE.Mesh, basePos: THREE.Vector3, phase: number }[] = []
const connections: { line: THREE.Line, nodes: [THREE.Mesh, THREE.Mesh] }[] = []
const workerConnections: { line: THREE.Line, nodes: [THREE.Mesh, THREE.Mesh] }[] = []

// Leader state
let leaderIdx = 0
let leaderChangeCooldown = 0
let prevLeaderIdx = 0

// Globe-compatible animation states
let rotationY = 0

// Add communication pulses between masters and workers
let mwCommPulses: { mesh: THREE.Mesh, from: THREE.Vector3, to: THREE.Vector3, t: number, active: boolean, direction: "m2w" | "w2m", workerIdx: number }[] = []
let mwCommCooldown = 0
const workerHighlightTimers: number[] = Array.from({ length: WORKER_COUNT }).map(() => 0)

const hoveredNode = shallowRef<{ mesh: THREE.Mesh, type: string, pos: { x: number, y: number } } | null>(null)

// Declare handleResize outside onMounted so it can be accessed in cleanup
let handleResize: (() => void) | null = null

function onCanvasMouseMove(event: MouseEvent) {
  if (!threeCanvas.value || !renderer || !camera)
    return
  const rect = threeCanvas.value.getBoundingClientRect()
  const mouse = new THREE.Vector2(
    ((event.clientX - rect.left) / rect.width) * 2 - 1,
    -((event.clientY - rect.top) / rect.height) * 2 + 1,
  )
  const allMeshes = [
    ...masters.map((m, i) => ({ mesh: m.mesh, type: i === leaderIdx ? "Leader" : "Master" })),
    ...workers.map(w => ({ mesh: w.mesh, type: "Worker" })),
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

  // Set initial canvas size
  const canvas = threeCanvas.value
  const rect = canvas.getBoundingClientRect()

  // Set explicit canvas size if dimensions are 0
  if (rect.width === 0 || rect.height === 0) {
    canvas.width = 800
    canvas.height = 600
  }

  renderer = new THREE.WebGLRenderer({
    canvas,
    alpha: true,
    antialias: true,
    powerPreference: "high-performance",
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
    1000,
  )
  camera.position.z = 5.5

  // Handle resize
  handleResize = () => {
    if (!threeCanvas.value || !renderer || !camera)
      return
    const rect = threeCanvas.value.getBoundingClientRect()
    const renderWidth = rect.width > 0 ? rect.width : 800
    const renderHeight = rect.height > 0 ? rect.height : 600
    renderer.setSize(renderWidth, renderHeight, false)
    camera.aspect = renderWidth / renderHeight
    camera.updateProjectionMatrix()
  }

  window.addEventListener("resize", handleResize)

  // Place master nodes in a ring
  const MASTER_RADIUS = 0.09 // Reduced from 0.18
  const MASTER_GLOW_RADIUS = 0.15 // Reduced from 0.32
  for (let i = 0; i < MASTER_COUNT; i++) {
    const angle = (i / MASTER_COUNT) * Math.PI * 2 - Math.PI / 2
    const basePos = new THREE.Vector3(
      MASTER_RING_RADIUS * Math.cos(angle),
      MASTER_RING_RADIUS * Math.sin(angle),
      0,
    )
    const masterGeometry = new THREE.SphereGeometry(MASTER_RADIUS, 18, 18)
    const masterGlowGeometry = new THREE.SphereGeometry(MASTER_GLOW_RADIUS, 18, 18)
    const mesh = new THREE.Mesh(masterGeometry, new THREE.MeshBasicMaterial({ color: MASTER_COLOR }))
    mesh.position.copy(basePos)
    scene.add(mesh)
    const glow = new THREE.Mesh(masterGlowGeometry, new THREE.MeshBasicMaterial({ color: MASTER_COLOR, transparent: true, opacity: 0.18 }))
    glow.position.copy(basePos)
    scene.add(glow)
    // Add leader highlight (ring) to the leader master
    let crown
    if (i === leaderIdx) {
      const ringGeometry = new THREE.TorusGeometry(MASTER_RADIUS + 0.03, 0.012, 16, 32)
      const ringMaterial = new THREE.MeshBasicMaterial({ color: 0xFBBF24, transparent: true, opacity: 0.7 })
      crown = new THREE.Mesh(ringGeometry, ringMaterial)
      crown.position.copy(basePos)
      crown.rotation.x = Math.PI / 2
      scene.add(crown)
    }
    masters.push({ mesh, glow, basePos, phase: Math.random() * Math.PI * 2, crown })
  }
  // Place worker nodes in a ring outside
  const WORKER_RADIUS = 0.07 // Reduced from 0.12
  const WORKER_GLOW_RADIUS = 0.12 // Reduced from 0.21
  for (let i = 0; i < WORKER_COUNT; i++) {
    const angle = (i / WORKER_COUNT) * Math.PI * 2
    const basePos = new THREE.Vector3(
      WORKER_RING_RADIUS * Math.cos(angle),
      WORKER_RING_RADIUS * Math.sin(angle),
      0,
    )
    const workerGeometry = new THREE.SphereGeometry(WORKER_RADIUS, 18, 18)
    const workerGlowGeometry = new THREE.SphereGeometry(WORKER_GLOW_RADIUS, 18, 18)
    const mesh = new THREE.Mesh(workerGeometry, new THREE.MeshBasicMaterial({ color: WORKER_COLOR }))
    mesh.position.copy(basePos)
    scene.add(mesh)
    const glow = new THREE.Mesh(workerGlowGeometry, new THREE.MeshBasicMaterial({ color: WORKER_COLOR, transparent: true, opacity: 0.18 }))
    glow.position.copy(basePos)
    scene.add(glow)
    workers.push({ mesh, glow, basePos, phase: Math.random() * Math.PI * 2 })
  }
  // Connect masters to each other (full mesh)
  for (let i = 0; i < MASTER_COUNT; i++) {
    for (let j = i + 1; j < MASTER_COUNT; j++) {
      const masterI = masters[i]
      const masterJ = masters[j]
      if (!masterI || !masterJ) {
        continue
      }

      const lineGeom = new THREE.BufferGeometry().setFromPoints([
        masterI.mesh.position,
        masterJ.mesh.position,
      ])
      const lineMat = new THREE.LineBasicMaterial({ color: CONNECTION_COLOR, transparent: true, opacity: 0.22 })
      const line = new THREE.Line(lineGeom, lineMat)
      scene.add(line)
      connections.push({ line, nodes: [masterI.mesh, masterJ.mesh] })
    }
  }
  // Connect each worker to all masters
  for (let i = 0; i < WORKER_COUNT; i++) {
    for (let j = 0; j < MASTER_COUNT; j++) {
      const workerI = workers[i]
      const masterJ = masters[j]
      if (!workerI || !masterJ) {
        continue
      }

      const lineGeom = new THREE.BufferGeometry().setFromPoints([
        workerI.mesh.position,
        masterJ.mesh.position,
      ])
      const lineMat = new THREE.LineBasicMaterial({ color: PRIMARY_COLOR, transparent: true, opacity: 0.13 })
      const line = new THREE.Line(lineGeom, lineMat)
      scene.add(line)
      workerConnections.push({ line, nodes: [workerI.mesh, masterJ.mesh] })
    }
  }
  // Animate
  function animate(time = 0) {
    animationId = requestAnimationFrame(animate)
    const t = time * 0.001

    // Auto-rotate like globe
    rotationY += 0.0005

    // Animate masters with calmer floating motion
    masters.forEach((master, idx) => {
      master.mesh.position.x = master.basePos.x + Math.sin(t * 0.7 + master.phase + idx * 0.3) * 0.04
      master.mesh.position.y = master.basePos.y + Math.cos(t * 0.9 + master.phase + idx * 0.4) * 0.04
      master.mesh.position.z = master.basePos.z + Math.sin(t * 0.5 + master.phase + idx * 0.2) * 0.03
      master.glow.position.copy(master.mesh.position)
      if (master.crown)
        master.crown.position.copy(master.mesh.position)
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
        conn.nodes[1].position,
      ])
      // Enhanced opacity animation like globe
      const material = conn.line.material as THREE.LineBasicMaterial
      material.opacity = 0.18 + Math.sin(t * 2 + idx * 0.3) * 0.08
    })

    workerConnections.forEach((conn, idx) => {
      conn.line.geometry.setFromPoints([
        conn.nodes[0].position,
        conn.nodes[1].position,
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
      leaderChangeCooldown = 180 + Math.random() * 80
      // Move crown
      const ms = masters[prevLeaderIdx]
      if (!ms) {
        return
      }

      if (ms.crown && scene) {
        scene.remove(ms.crown)
      }

      const lm = masters[leaderIdx]
      if (!lm) {
        return
      }

      lm.crown = createCrownMesh()
      if (lm.crown) {
        lm.crown!.position.copy(lm.mesh.position)
        if (scene) {
          scene.add(lm.crown!)
        }
      }
    }

    // Animate crown scale for leader
    masters.forEach((master, idx) => {
      if (master.crown) {
        const scale = idx === leaderIdx ? 1 + 0.13 * Math.sin(t * 2.5) : 1
        master.crown.scale.set(scale, scale, scale)
      }
      // Animate leader glow with enhanced effects
      if (idx === leaderIdx) {
        let glowMat = master.glow.material
        if (Array.isArray(glowMat))
          glowMat = glowMat[0] ?? []
        if (glowMat && "opacity" in glowMat) {
          glowMat.opacity = 0.35 + 0.12 * Math.abs(Math.sin(t * 2.5))
        }
        let meshMat = master.mesh.material
        if (Array.isArray(meshMat))
          meshMat = meshMat[0] ?? []
        if (meshMat && "color" in meshMat && meshMat.color instanceof THREE.Color) {
          meshMat.color.set(LEADER_COLOR)
        }
      }
      else {
        let glowMat = master.glow.material
        if (Array.isArray(glowMat))
          glowMat = glowMat[0] ?? []
        if (glowMat && "opacity" in glowMat) {
          glowMat.opacity = 0.18 + 0.05 * Math.sin(t * 1.5 + idx)
        }
        let meshMat = master.mesh.material
        if (Array.isArray(meshMat))
          meshMat = meshMat[0] ?? []
        if (meshMat && "color" in meshMat && meshMat.color instanceof THREE.Color) {
          meshMat.color.set(MASTER_COLOR)
        }
      }
    })

    // Generate master→worker and worker→master communication pulses
    if (mwCommPulses.length === 0) {
      mwCommCooldown--

      if (mwCommCooldown <= 0) {
        const masterIdx = Math.floor(Math.random() * MASTER_COUNT)
        const workerIdx = Math.floor(Math.random() * WORKER_COUNT)

        const ms = masters[masterIdx]
        const wr = workers[workerIdx]
        if (!ms || !wr) {
          return
        }

        mwCommPulses.push({
          mesh: createCommPulseMesh("m2w"),
          from: ms.mesh.position.clone(),
          to: wr.mesh.position.clone(),
          t: 0,
          active: true,
          direction: "m2w",
          workerIdx,
        })
        if (Math.random() < 0.5) {
          mwCommPulses.push({
            mesh: createCommPulseMesh("w2m"),
            from: wr.mesh.position.clone(),
            to: ms.mesh.position.clone(),
            t: 0,
            active: true,
            direction: "w2m",
            workerIdx,
          })
        }
        mwCommCooldown = 200 + Math.random() * 120
      }
    }
    // Animate communication pulses
    mwCommPulses = mwCommPulses.filter((pulse) => {
      pulse.t += 0.025
      if (pulse.t >= 1) {
        if (scene)
          scene.remove(pulse.mesh)
        // Highlight worker node if pulse was master→worker
        if (pulse.direction === "m2w") {
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
      const wh = workerHighlightTimers[idx]
      if (wh && wh > 0) {
        workerHighlightTimers[idx]! -= 1
        let mat = worker.mesh.material
        if (Array.isArray(mat))
          mat = mat[0] ?? []
        if (mat && "color" in mat && mat.color instanceof THREE.Color) {
          mat.color.lerp(new THREE.Color(WORKER_COLOR), 0.15)
          mat.color.offsetHSL(0, 0, 0.08)
        }
        let glowMat = worker.glow.material
        if (Array.isArray(glowMat))
          glowMat = glowMat[0] ?? []
        if (glowMat && "opacity" in glowMat) {
          glowMat.opacity = 0.18 + 0.17 * (wh / 12)
        }
      }
      else {
        let mat = worker.mesh.material
        if (Array.isArray(mat))
          mat = mat[0] ?? []
        if (mat && "color" in mat && mat.color instanceof THREE.Color) {
          mat.color.set(WORKER_COLOR)
        }
        let glowMat = worker.glow.material
        if (Array.isArray(glowMat))
          glowMat = glowMat[0] ?? []
        if (glowMat && "opacity" in glowMat) {
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
      opacity: 0.6,
    })
    const ring = new THREE.Mesh(ringGeometry, ringMaterial)
    ring.position.z = 0.15
    return ring
  }

  // Helper to create a communication pulse mesh
  function createCommPulseMesh(direction: "m2w" | "w2m") {
    const color = direction === "m2w" ? 0x60A5FA : 0x34D399 // blue for master→worker, green for worker→master
    const geometry = new THREE.SphereGeometry(0.05, 10, 10)
    const material = new THREE.MeshBasicMaterial({ color, transparent: true, opacity: 0.7 })
    const mesh = new THREE.Mesh(geometry, material)
    if (scene)
      scene.add(mesh)
    return mesh
  }

  animate()
})

onBeforeUnmount(() => {
  if (animationId)
    cancelAnimationFrame(animationId)
  if (renderer) {
    renderer.dispose()
    renderer = null
  }
  if (handleResize) {
    window.removeEventListener("resize", handleResize)
  }
})
</script>
