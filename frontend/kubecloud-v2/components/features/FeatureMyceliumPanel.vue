<template>
  <FeatureSectionLayout
    title="Mycelium Networking"
    description="Ultra-fast, decentralized networking inspired by nature. Mycelium Networking forms a resilient, adaptive mesh that routes around failures and optimizes for speed and security."
    :tags="['End-to-end encrypted', 'Nature-inspired']"
    :hovered-node="hoveredNodeLabel ? { label: 'Peer', pos: hoveredNodeLabel } : undefined"
    #="{ canvasProps }"
  >
    <canvas
      v-bind="canvasProps"
      ref="threeCanvas"
      @mousemove="handleMouseMove"
      @mouseleave="handleMouseLeave"
      @click="handleClick"
    />
  </FeatureSectionLayout>
</template>

<script setup lang="ts">
import type { Line, Mesh, Vector3, WebGLRenderer } from "three"

const THREE = ensureThreeGlobal()

const threeCanvas = ref<HTMLCanvasElement | null>(null)
let renderer: WebGLRenderer | null = null
let animationId: number | null = null

const NODE_COUNT = 32
const CONNECTION_DISTANCE = 2.2
const nodes: { mesh: Mesh, basePos: Vector3, phase: number, glow: Mesh }[] = []
const activeConnections: { i: number, j: number, line: Line, t: number, direction: 1 | -1, opacity: number, fadingOut: boolean, maxOpacity: number, fadeInSpeed: number, fadeOutSpeed: number, lifetime: number, fadeInDelay: number, fadeInWait: number }[] = []

// Interactivity state
const hoveredNodeIdx = shallowRef<number | null>(null)
let pulse = 0
let pulseStart = 0
const raycaster = new THREE.Raycaster()
const mouse = new THREE.Vector2()

// Restore original lines array and organic growth logic
const lines: { line: Line, i: number, j: number, progress: number, growing: boolean }[] = []

// Add hover label state
const hoveredNodeLabel = shallowRef<{ x: number, y: number } | null>(null)

function handleMouseMove(event: MouseEvent) {
  if (!threeCanvas.value || !renderer)
    return
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

onMounted(() => {
  if (!threeCanvas.value)
    return
  renderer = new THREE.WebGLRenderer({ canvas: threeCanvas.value, alpha: true, antialias: true })
  renderer.setSize(threeCanvas.value.clientWidth / 1, threeCanvas.value.clientHeight / 1, false)
  renderer.setClearColor(0x000000, 0)
  const scene = new THREE.Scene()
  const camera = new THREE.PerspectiveCamera(
    60,
    threeCanvas.value.clientWidth / threeCanvas.value.clientHeight,
    0.1,
    1000,
  )
  camera.position.z = 6

  // Create nodes (glowing points)
  const nodeGeometry = new THREE.OctahedronGeometry(0.08, 0) // Diamond/octahedron shape for peers
  const nodeMaterial = new THREE.MeshBasicMaterial({ color: 0x60A5FA })
  for (let i = 0; i < NODE_COUNT; i++) {
    const phi = Math.acos(2 * Math.random() - 1)
    const theta = 2 * Math.PI * Math.random()
    const r = 1.7 + Math.random() * 1.1
    const basePos = new THREE.Vector3(
      r * Math.sin(phi) * Math.cos(theta),
      r * Math.sin(phi) * Math.sin(theta),
      r * Math.cos(phi),
    )
    const mesh = new THREE.Mesh(nodeGeometry, nodeMaterial.clone())
    mesh.position.copy(basePos)
    scene.add(mesh)
    // Add glow - also diamond shaped but larger
    const glowMat = new THREE.MeshBasicMaterial({ color: 0x60A5FA, transparent: true, opacity: 0.18 })
    const glow = new THREE.Mesh(new THREE.OctahedronGeometry(0.18, 0), glowMat)
    glow.position.copy(basePos)
    scene.add(glow)
    nodes.push({ mesh, basePos, phase: Math.random() * Math.PI * 2, glow })
  }

  // Create lines (connections)
  for (let i = 0; i < NODE_COUNT; i++) {
    const nodeI = nodes[i]
    for (let j = i + 1; j < NODE_COUNT; j++) {
      const nodeJ = nodes[j]
      if (nodeI && nodeJ && nodeI.basePos.distanceTo(nodeJ.basePos) < CONNECTION_DISTANCE) {
        const points = [nodeI.mesh.position, nodeJ.mesh.position]
        const lineGeom = new THREE.BufferGeometry().setFromPoints(points)
        const lineMat = new THREE.LineBasicMaterial({ color: 0x60A5FA, transparent: true, opacity: 0.22 })
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
    const line = lines[nextLineToGrow]
    if (line && nextLineToGrow < lines.length) {
      line.growing = true
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
      const nodeI = nodes[i]
      for (let j = i + 1; j < NODE_COUNT; j++) {
        const nodeJ = nodes[j]
        if (nodeI && nodeJ && nodeI.basePos.distanceTo(nodeJ.basePos) < CONNECTION_DISTANCE) {
          const lineObj = lines[lineIdx++]
          if (!lineObj) {
            continue
          }

          const { line, i: ni, j: nj, growing } = lineObj
          const nodeNI = nodes[ni]
          const nodeNJ = nodes[nj]
          if (!nodeNI || !nodeNJ) {
            continue
          }
          line.geometry.setFromPoints([
            nodeNI.mesh.position,
            nodeNJ.mesh.position,
          ])
          // Animate growth
          if (growing && lineObj.progress < 1) {
            lineObj.progress += 0.045 + Math.random() * 0.02
            if (lineObj.progress > 1)
              lineObj.progress = 1
            line.geometry.setDrawRange(0, Math.floor(2 * lineObj.progress))
          }
          else if (growing) {
            line.geometry.setDrawRange(0, 2)
          }
          else {
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
    const intersect = intersects[0]
    if (intersect) {
      hoveredNodeIdx.value = nodes.findIndex(n => n.mesh === intersect.object)
      const node = nodes[hoveredNodeIdx.value]
      if (!node) {
        return
      }
      // Project 3D position to 2D overlay for label
      const pos = node.mesh.position.clone().project(camera)
      const rect = threeCanvas.value!.getBoundingClientRect()
      const x = ((pos.x + 1) / 2) * rect.width
      const y = ((-pos.y + 1) / 2) * rect.height
      hoveredNodeLabel.value = { x, y }
    }
    // Highlight hovered node and its glow
    nodes.forEach((node, idx) => {
      let glowMat = node.glow.material
      if (Array.isArray(glowMat))
        glowMat = glowMat[0] ?? []
      let meshMat = node.mesh.material
      if (Array.isArray(meshMat))
        meshMat = meshMat[0] ?? []
      // Gentle pulse for all nodes
      const basePulse = 1 + Math.sin(time * 0.002 + idx) * 0.08
      // If hovered, brighten and color
      if (idx === hoveredNodeIdx.value) {
        if (glowMat && "opacity" in glowMat)
          glowMat.opacity = 0.45
        if (meshMat && "color" in meshMat && meshMat.color instanceof THREE.Color)
          meshMat.color.set(0x8ECFFF)
        node.mesh.scale.set(basePulse * 1.12, basePulse * 1.12, basePulse * 1.12)
        node.glow.scale.set(basePulse * 1.12, basePulse * 1.12, basePulse * 1.12)
      }
      else {
        if (glowMat && "opacity" in glowMat)
          glowMat.opacity = 0.18
        if (meshMat && "color" in meshMat && meshMat.color instanceof THREE.Color)
          meshMat.color.set(0x60A5FA)
        node.mesh.scale.set(basePulse, basePulse, basePulse)
        node.glow.scale.set(basePulse, basePulse, basePulse)
      }
    })
    // Optionally, highlight lines connected to hovered node
    activeConnections.forEach((conn) => {
      let lineMat = conn.line.material
      if (Array.isArray(lineMat))
        lineMat = lineMat[0] ?? []
      if (lineMat && "opacity" in lineMat) {
        if (hoveredNodeIdx.value !== null && (conn.i === hoveredNodeIdx.value || conn.j === hoveredNodeIdx.value)) {
          lineMat.opacity = 0.45
          if (lineMat && "color" in lineMat && lineMat.color instanceof THREE.Color)
            lineMat.color.set(0x8ECFFF)
        }
        else {
          lineMat.opacity = 0.22
          if (lineMat && "color" in lineMat && lineMat.color instanceof THREE.Color)
            lineMat.color.set(0x60A5FA)
        }
      }
    })
    // Pulse effect on click
    if (pulse > 0 && hoveredNodeIdx.value !== null) {
      const elapsed = (performance.now() - pulseStart) / 400
      const scale = 1 + Math.sin(Math.PI * Math.min(elapsed, 1)) * 0.7
      const node = nodes[hoveredNodeIdx.value]
      if (!node) {
        return
      }

      node.mesh.scale.set(scale, scale, scale)
      node.glow.scale.set(scale, scale, scale)
      if (elapsed >= 1) {
        pulse = 0
        node.mesh.scale.set(1, 1, 1)
        node.glow.scale.set(1, 1, 1)
      }
    }
    renderer!.render(scene, camera)
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
})
</script>
