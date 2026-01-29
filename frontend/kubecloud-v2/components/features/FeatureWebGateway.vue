<template>
  <FeatureSectionLayout
    title="Simple Web Gateway Access"
    description="Expose any service to the public web with a simple Kubernetes resource. No complex Ingress controllers. Domain and prefix-based routing is built-in."
    :tags="['Simple configuration', 'Built-in routing', 'No ingress controllers']"
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
import type { Line, LineBasicMaterial, Mesh, PerspectiveCamera, Scene, Vector3, WebGLRenderer } from "three"

const THREE = ensureThreeGlobal()

const threeCanvas = ref<HTMLCanvasElement | null>(null)
let renderer: WebGLRenderer | null = null
let animationId: number | null = null
let scene: Scene | null = null
let camera: PerspectiveCamera | null = null

// Network topology constants
const CLIENT_COUNT = 5
const SERVER_COUNT = 4
const GATEWAY_NODES = 3
const GATEWAY_RADIUS = 0.8

// Sophisticated color palette
const PRIMARY_COLOR = 0x60A5FA
const ACCENT_COLOR = 0x8ECFFF
const SUCCESS_COLOR = 0x34D399

// Network nodes and connections
const clients: { mesh: Mesh, glow: Mesh, basePos: Vector3, phase: number }[] = []
const servers: { mesh: Mesh, glow: Mesh, basePos: Vector3, phase: number, type: "public" | "private" }[] = []
const gatewayNodes: { mesh: Mesh, glow: Mesh, basePos: Vector3, phase: number }[] = []
const connections: { line: Line, nodes: [Mesh, Mesh] }[] = []

// Add a ref to track gateway pulse state
const gatewayPulse = Array.from({ length: GATEWAY_NODES }).map(() => 0)

// Add hover state for node labels
const hoveredNode = shallowRef<{ mesh: Mesh, type: string, pos: { x: number, y: number } } | null>(null)

// Add at the top:
let pulses: { mesh: Mesh, path: Vector3[], t: number, stage: 0 | 1, active: boolean }[] = []
let pulseCooldown = 0

// Add mouse event handlers for hover detection
function onCanvasMouseMove(event: MouseEvent) {
  if (!threeCanvas.value || !renderer || !camera)
    return
  const rect = threeCanvas.value.getBoundingClientRect()
  const mouse = new THREE.Vector2(
    ((event.clientX - rect.left) / rect.width) * 2 - 1,
    -((event.clientY - rect.top) / rect.height) * 2 + 1,
  )
  // Raycast against all nodes
  const allMeshes = [
    ...clients.map(c => ({ mesh: c.mesh, type: "Client" })),
    ...gatewayNodes.map(g => ({ mesh: g.mesh, type: "Gateway" })),
    ...servers.map(s => ({ mesh: s.mesh, type: s.type === "public" ? "Public Server" : "Private Server" })),
  ]
  const raycaster = new THREE.Raycaster()
  raycaster.setFromCamera(mouse, camera)
  const intersects = raycaster.intersectObjects(allMeshes.map(n => n.mesh))
  const intersect = intersects[0]
  if (intersect) {
    const found = allMeshes.find(n => n.mesh === intersect.object)
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
  if (!threeCanvas.value)
    return

  // Initialize Three.js with sophisticated settings
  renderer = new THREE.WebGLRenderer({
    canvas: threeCanvas.value,
    alpha: true,
    antialias: true,
    powerPreference: "high-performance",
  })
  renderer.setSize(threeCanvas.value.clientWidth, threeCanvas.value.clientHeight, false)
  renderer.setClearColor(0x000000, 0) // Transparent background like Mycelium

  scene = new THREE.Scene()
  camera = new THREE.PerspectiveCamera(
    60,
    threeCanvas.value.clientWidth / threeCanvas.value.clientHeight,
    0.1,
    1000,
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
      (Math.random() - 0.5) * 0.3,
    )

    const mesh = new THREE.Mesh(clientGeometry, new THREE.MeshBasicMaterial({
      color: PRIMARY_COLOR,
      transparent: true,
      opacity: 0.9,
    }))
    mesh.position.copy(basePos)
    scene.add(mesh)

    // Subtle glow
    const glowGeometry = new THREE.SphereGeometry(0.14, 16, 16)
    const glow = new THREE.Mesh(glowGeometry, new THREE.MeshBasicMaterial({
      color: PRIMARY_COLOR,
      transparent: true,
      opacity: 0.15,
    }))
    glow.position.copy(basePos)
    scene.add(glow)

    clients.push({
      mesh,
      glow,
      basePos: basePos.clone(),
      phase: Math.random() * Math.PI * 2,
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
      (Math.random() - 0.5) * 0.2,
    )

    const mesh = new THREE.Mesh(gatewayGeometry, new THREE.MeshBasicMaterial({
      color: ACCENT_COLOR,
      transparent: true,
      opacity: 0.95,
    }))
    mesh.position.copy(basePos)
    scene.add(mesh)

    // Enhanced glow for gateway nodes (sphere glow)
    const glowGeometry = new THREE.SphereGeometry(0.15, 18, 18)
    const glow = new THREE.Mesh(glowGeometry, new THREE.MeshBasicMaterial({
      color: ACCENT_COLOR,
      transparent: true,
      opacity: 0.2,
    }))
    glow.position.copy(basePos)
    scene.add(glow)

    gatewayNodes.push({
      mesh,
      glow,
      basePos: basePos.clone(),
      phase: Math.random() * Math.PI * 2,
    })
  }

  // Create server nodes (right side) - elegant endpoints
  const serverGeometry = new THREE.BoxGeometry(0.18, 0.11, 0.09)
  // Randomly assign half as private, half as public
  const serverTypes: ("public" | "private")[] = Array.from({ length: SERVER_COUNT }).fill("public").map((v, i) => i < Math.floor(SERVER_COUNT / 2) ? "private" : "public")
  // Shuffle for randomness
  for (let i = serverTypes.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1))
    const t1 = serverTypes[i]
    const t2 = serverTypes[j]
    if (t1 && t2) {
      serverTypes[i] = t2
      serverTypes[j] = t1
    }
  }
  for (let i = 0; i < SERVER_COUNT; i++) {
    // Place servers in a rightward arc, further out
    const angle = (i / SERVER_COUNT) * Math.PI * 1.2 - Math.PI * 0.1 // More rightward
    const radius = 2.6 + Math.random() * 0.3 // Increased from 2.4 + Math.random() * 0.3
    const basePos = new THREE.Vector3(
      radius * Math.cos(angle),
      radius * Math.sin(angle),
      (Math.random() - 0.5) * 0.2,
    )
    const type = serverTypes[i]
    if (!type) {
      return
    }

    const mesh = new THREE.Mesh(serverGeometry, new THREE.MeshBasicMaterial({
      color: SUCCESS_COLOR,
      transparent: true,
      opacity: 0.85,
    }))
    mesh.position.copy(basePos)
    if (scene)
      scene.add(mesh)
    // Glow: thicker and brighter for private
    const glowGeometry = new THREE.BoxGeometry(type === "private" ? 0.28 : 0.21, type === "private" ? 0.17 : 0.13, type === "private" ? 0.13 : 0.09)
    const glow = new THREE.Mesh(glowGeometry, new THREE.MeshBasicMaterial({
      color: SUCCESS_COLOR,
      transparent: true,
      opacity: type === "private" ? 0.28 : 0.12,
    }))
    glow.position.copy(basePos)
    if (scene)
      scene.add(glow)
    servers.push({
      mesh,
      glow,
      basePos: basePos.clone(),
      phase: Math.random() * Math.PI * 2,
      type,
    })
  }

  // Connect clients <-> gateway, gateway <-> servers (all), clients <-> public servers (direct)
  // 1. Clients <-> Gateway
  clients.forEach((client) => {
    gatewayNodes.forEach((gateway) => {
      const lineMaterial = new THREE.LineBasicMaterial({
        color: PRIMARY_COLOR,
        transparent: true,
        opacity: 0.15,
      })
      const p1 = client.mesh.position.clone()
      const p2 = gateway.mesh.position.clone()
      const lineGeom = new THREE.BufferGeometry().setFromPoints([p1, p2])
      const line = new THREE.Line(lineGeom, lineMaterial)
      if (scene)
        scene.add(line)
      connections.push({
        line,
        nodes: [client.mesh, gateway.mesh],
      })
    })
  })
  // 2. Gateway <-> All servers
  gatewayNodes.forEach((gateway) => {
    servers.forEach((server) => {
      const lineMaterial = new THREE.LineBasicMaterial({
        color: PRIMARY_COLOR,
        transparent: true,
        opacity: 0.15,
      })
      const p1b = gateway.mesh.position.clone()
      const p2b = server.mesh.position.clone()
      const lineGeomb = new THREE.BufferGeometry().setFromPoints([p1b, p2b])
      const lineb = new THREE.Line(lineGeomb, lineMaterial)
      if (scene)
        scene.add(lineb)
      connections.push({
        line: lineb,
        nodes: [gateway.mesh, server.mesh],
      })
    })
  })
  // 3. Clients <-> Public servers (direct)
  clients.forEach((client) => {
    servers.forEach((server) => {
      if (server.type === "public") {
        const lineMaterial = new THREE.LineBasicMaterial({
          color: PRIMARY_COLOR,
          transparent: true,
          opacity: 0.15,
        })
        const p1c = client.mesh.position.clone()
        const p2c = server.mesh.position.clone()
        const lineGeomc = new THREE.BufferGeometry().setFromPoints([p1c, p2c])
        const linec = new THREE.Line(lineGeomc, lineMaterial)
        if (scene)
          scene.add(linec)
        connections.push({
          line: linec,
          nodes: [client.mesh, server.mesh],
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
      const gwPulse = gatewayPulse[idx]
      if (!gwPulse) {
        return
      }

      if (gwPulse > 0) {
        pulse += 0.13 * gwPulse
        gatewayPulse[idx]! -= 0.06
        if (gwPulse < 0)
          gatewayPulse[idx] = 0
      }
      gateway.mesh.scale.set(pulse, pulse, pulse)
      gateway.glow.scale.set(pulse * 1.18, pulse * 1.18, pulse * 1.18)
      let glowMat = gateway.glow.material
      if (Array.isArray(glowMat))
        glowMat = glowMat[0] ?? []
      if (glowMat && "opacity" in glowMat) {
        glowMat.opacity = 0.22 + 0.10 * Math.abs(Math.sin(t * 2.5 + idx)) + 0.13 * (gwPulse > 0 ? gwPulse : 0)
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
        const nodeI = allNodes[i]
        const nodeJ = allNodes[j]
        if (!nodeI || !nodeJ) {
          continue
        }
        if (nodeI.basePos.distanceTo(nodeJ.basePos) < connectionDistance) {
          if (connectionIdx < connections.length) {
            const conn = connections[connectionIdx]
            if (!conn) {
              continue
            }

            // Always use current mesh positions for curve endpoints
            const p1 = nodeI.mesh.position
            const p2 = nodeJ.mesh.position
            conn.line.geometry.setFromPoints([p1, p2])
            // Subtle opacity animation
            const material = conn.line.material as LineBasicMaterial
            material.opacity = 0.12 + Math.sin(t * 2 + connectionIdx * 0.3) * 0.08
          }
          connectionIdx++
        }
      }
    }

    // Animate pulses
    pulseCooldown--
    if (pulseCooldown <= 0) {
      // Pick a random client, gateway, and server
      const clientIdx = Math.floor(Math.random() * clients.length)
      const gatewayIdx = Math.floor(Math.random() * gatewayNodes.length)
      const serverIdx = Math.floor(Math.random() * servers.length)

      const c = clients[clientIdx]
      const g = gatewayNodes[gatewayIdx]
      const s = servers[serverIdx]
      if (!c || !g || !s) {
        return
      }

      const path = [c.mesh.position.clone(), g.mesh.position.clone(), s.mesh.position.clone()]
      pulses.push({ mesh: createPulseMesh(), path, t: 0, stage: 0, active: true })
      pulseCooldown = 30 + Math.random() * 40
    }
    pulses.forEach((pulse) => {
      if (!pulse.active)
        return
      pulse.t += 0.012
      let pos: Vector3
      const [p0, p1, p2] = pulse.path
      if (pulse.stage === 0) {
        if (!p0 || !p1) {
          return
        }

        // Client to Gateway
        pos = new THREE.Vector3().lerpVectors(p0, p1, pulse.t * 2)
        if (pulse.t >= 0.5) {
          pulse.stage = 1
          pulse.t = 0
        }
      }
      else {
        if (!p1 || !p2) {
          return
        }
        // Gateway to Server
        pos = new THREE.Vector3().lerpVectors(p1, p2, pulse.t * 2)
        if (pulse.t >= 0.5) {
          pulse.active = false
          scene!.remove(pulse.mesh)
        }
      }
      pulse.mesh.position.copy(pos)
      let mat = pulse.mesh.material
      if (Array.isArray(mat))
        mat = mat[0] ?? []
      if (mat && "opacity" in mat)
        mat.opacity = 0.7 + 0.25 * Math.sin(Math.PI * pulse.t)
    })
    pulses = pulses.filter(p => p.active)

    renderer?.render(scene!, camera!)
  }
  animate()
})

onBeforeUnmount(() => {
  if (animationId)
    cancelAnimationFrame(animationId)

  // Clean up Three.js objects
  if (scene) {
    // Dispose of geometries and materials
    scene.traverse((object) => {
      if (object instanceof THREE.Mesh) {
        if (object.geometry)
          object.geometry.dispose()
        if (object.material) {
          if (Array.isArray(object.material)) {
            object.material.forEach(material => material.dispose())
          }
          else {
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
  const pulseGeom = new THREE.SphereGeometry(0.06, 16, 16)
  const pulseMat = new THREE.MeshBasicMaterial({ color: 0xFBBF24, transparent: true, opacity: 0.95 })
  const mesh = new THREE.Mesh(pulseGeom, pulseMat)
  scene!.add(mesh)
  return mesh
}
</script>
