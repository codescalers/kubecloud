<template>
  <div
    ref="globeContainer"
    class="globe-canvas"
    :style="{ width: width + 'px', height: height + 'px', maxWidth: '100%' }"
  >
    <slot />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import * as THREE from 'three'

/**
 * Props for FeatureGlobe
 * @prop {number} width - Globe canvas width in px
 * @prop {number} height - Globe canvas height in px
 * @prop {number} pointCount - Number of points to render (default: 2500)
 * @prop {[number, number][]} nodes - Optional array of [lat, lng] for node positions
 * @prop {Array<{from: [number, number], to: [number, number]}>} arcs - Optional array of arcs
 */
const props = defineProps({
  width: { type: Number, default: 480 },
  height: { type: Number, default: 480 },
  pointCount: { type: Number, default: 2500 },
  nodes: { type: Array as () => [number, number][], default: undefined },
  arcs: { type: Array as () => Array<{from: [number, number], to: [number, number]}>, default: undefined },
})

const emit = defineEmits(['node-click', 'node-hover', 'arc-hover'])

const globeContainer = ref<HTMLElement | null>(null)
let renderer: THREE.WebGLRenderer | null = null
let scene: THREE.Scene | null = null
let camera: THREE.PerspectiveCamera | null = null
let animationId: number | null = null
let pointCloud: THREE.Points | null = null
let raycaster: THREE.Raycaster | null = null
let mouse = new THREE.Vector2()
let INTERSECTED: number | null = null
let isDragging = false
let lastMouseX = 0
let lastMouseY = 0
let rotationY = 0
let rotationX = 0
let autoRotate = true

const GLOBE_RADIUS = 2.9
const POINT_COLOR = '#60a5fa' // KubeCloud blue
const POINT_COLOR_HIGHLIGHT = '#fff'
const ARC_COLOR = '#38bdf8' // Cyan

type NodeIndex = number
import { ref as vueRef } from 'vue'
const pulsingNodes = vueRef<NodeIndex[]>([])
let pulseInterval: number | null = null

function randomLatLngPoints(count: number): [number, number][] {
  const points: [number, number][] = []
  for (let i = 0; i < count; i++) {
    const u = Math.random()
    const v = Math.random()
    const theta = 2 * Math.PI * u
    const phi = Math.acos(2 * v - 1)
    const lat = 90 - (phi * 180 / Math.PI)
    const lng = (theta * 180 / Math.PI) - 180
    points.push([lat, lng])
  }
  return points
}

function getNodePositions(): [number, number][] {
  return props.nodes && props.nodes.length > 0
    ? props.nodes
    : randomLatLngPoints(props.pointCount)
}

function createGlobe() {
  scene = new THREE.Scene()
  scene.background = null
  camera = new THREE.PerspectiveCamera(60, props.width / props.height, 0.1, 100)
  camera.position.z = 7

  // Point cloud geometry (darker blue palette)
  const nodePositions = getNodePositions()
  const geometry = new THREE.BufferGeometry()
  const positions = new Float32Array(nodePositions.length * 3)
  const colors = new Float32Array(nodePositions.length * 3)
  for (let i = 0; i < nodePositions.length; i++) {
    const [lat, lng] = nodePositions[i]
    const phi = (90 - lat) * (Math.PI / 180)
    const theta = (lng + 180) * (Math.PI / 180)
    const x = GLOBE_RADIUS * Math.sin(phi) * Math.cos(theta)
    const y = GLOBE_RADIUS * Math.cos(phi)
    const z = GLOBE_RADIUS * Math.sin(phi) * Math.sin(theta)
    positions[i * 3] = x
    positions[i * 3 + 1] = y
    positions[i * 3 + 2] = z
    // Darker blue gradient (less white, more blue)
    const c = 0.7 + 0.3 * (y / GLOBE_RADIUS)
    colors[i * 3] = 0.18 * c + 0.10 // less white, more blue
    colors[i * 3 + 1] = 0.40 * c + 0.10 // less white, more blue
    colors[i * 3 + 2] = 0.80 * c + 0.15 // more blue
  }
  geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3))
  geometry.setAttribute('color', new THREE.BufferAttribute(colors, 3))

  const material = new THREE.PointsMaterial({
    size: 0.08,
    vertexColors: true,
    transparent: true,
    opacity: 0.92
  })
  pointCloud = new THREE.Points(geometry, material)
  scene.add(pointCloud)

  // Lighting (minimal, just for points)
  const ambientLight = new THREE.AmbientLight('#ffffff', 0.5)
  scene.add(ambientLight)

  // Raycaster for hover/click
  raycaster = new THREE.Raycaster()
}

function animate() {
  if (!renderer || !scene || !camera || !pointCloud) return
  animationId = requestAnimationFrame(animate)
  // Auto-rotate unless dragging
  if (autoRotate && !isDragging) {
    rotationY += 0.002
  }
  if (pointCloud) {
    pointCloud.rotation.y = rotationY
    pointCloud.rotation.x = rotationX
    // Color pulse effect for pulsingNodes
    const geometry = pointCloud.geometry as THREE.BufferGeometry
    const colors = geometry.getAttribute('color') as THREE.BufferAttribute
    const baseColors = geometry.getAttribute('baseColor') as THREE.BufferAttribute | undefined
    const time = performance.now() * 0.002
    if (!baseColors) {
      // Store original colors for pulsing
      const orig = colors.array.slice()
      geometry.setAttribute('baseColor', new THREE.BufferAttribute(new Float32Array(orig), 3))
    }
    const base = (geometry.getAttribute('baseColor') as THREE.BufferAttribute)?.array
    if (base) {
      for (const idx of pulsingNodes.value) {
        // Animate color: blend between base and 'on' color (white)
        const t = 0.5 + 0.5 * Math.sin(time * 2 + idx)
        // Base color
        const r0 = base[idx * 3]
        const g0 = base[idx * 3 + 1]
        const b0 = base[idx * 3 + 2]
        // 'On' color (white)
        const r1 = 1.0, g1 = 1.0, b1 = 1.0
        colors.setX(idx, r0 * (1 - t) + r1 * t)
        colors.setY(idx, g0 * (1 - t) + g1 * t)
        colors.setZ(idx, b0 * (1 - t) + b1 * t)
      }
      // Reset non-pulsing nodes to base color
      for (let i = 0; i < colors.count; i++) {
        if (!pulsingNodes.value.includes(i)) {
          colors.setX(i, base[i * 3])
          colors.setY(i, base[i * 3 + 1])
          colors.setZ(i, base[i * 3 + 2])
        }
      }
      colors.needsUpdate = true
    }
  }
  // Hover effect
  if (raycaster && pointCloud && globeContainer.value) {
    raycaster.setFromCamera(mouse, camera!)
    const intersects = raycaster.intersectObject(pointCloud)
    const geometry = pointCloud.geometry as THREE.BufferGeometry
    const colors = geometry.getAttribute('color') as THREE.BufferAttribute
    if (INTERSECTED !== null) {
      // Restore previous color
      colors.setX(INTERSECTED * 3, 0.22)
      colors.setY(INTERSECTED * 3 + 1, 0.65)
      colors.setZ(INTERSECTED * 3 + 2, 0.98)
    }
    if (intersects.length > 0) {
      const index = intersects[0].index
      if (index !== undefined) {
        INTERSECTED = index
        // Highlight color
        colors.setX(index * 3, 1.0)
        colors.setY(index * 3 + 1, 1.0)
        colors.setZ(index * 3 + 2, 1.0)
        emit('node-hover', index)
      }
    } else {
      INTERSECTED = null
    }
    colors.needsUpdate = true
  }
  renderer.render(scene, camera)
}

function resizeRenderer() {
  if (!renderer || !camera || !globeContainer.value) return
  renderer.setSize(props.width, props.height, false)
  camera.aspect = props.width / props.height
  camera.updateProjectionMatrix()
}

function onPointerDown(event: MouseEvent) {
  isDragging = true
  autoRotate = false
  lastMouseX = event.clientX
  lastMouseY = event.clientY
}
function onPointerUp() {
  isDragging = false
  autoRotate = true
}
function onPointerMove(event: MouseEvent) {
  if (!globeContainer.value) return
  // Mouse for raycaster
  const rect = globeContainer.value.getBoundingClientRect()
  mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1
  mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1
  // Drag to rotate
  if (isDragging) {
    const deltaX = event.clientX - lastMouseX
    const deltaY = event.clientY - lastMouseY
    rotationY += deltaX * 0.01
    rotationX += deltaY * 0.01
    lastMouseX = event.clientX
    lastMouseY = event.clientY
  }
}
function onPointerClick(event: MouseEvent) {
  if (INTERSECTED !== null && pointCloud) {
    emit('node-click', INTERSECTED)
    // Pulse effect: temporarily enlarge the clicked point
    const geometry = pointCloud.geometry as THREE.BufferGeometry
    const positions = geometry.getAttribute('position') as THREE.BufferAttribute
    const i = INTERSECTED
    const origX = positions.getX(i * 3)
    const origY = positions.getY(i * 3 + 1)
    const origZ = positions.getZ(i * 3 + 2)
    // Animate pulse
    let t = 0
    function pulse() {
      t += 0.08
      const scale = 1 + 0.5 * Math.sin(Math.PI * t) * 0.25
      positions.setX(i * 3, origX * scale)
      positions.setY(i * 3 + 1, origY * scale)
      positions.setZ(i * 3 + 2, origZ * scale)
      positions.needsUpdate = true
      if (t < 1) {
        requestAnimationFrame(pulse)
      } else {
        positions.setX(i * 3, origX)
        positions.setY(i * 3 + 1, origY)
        positions.setZ(i * 3 + 2, origZ)
        positions.needsUpdate = true
      }
    }
    pulse()
  }
}

function pickRandomPulsingNodes(count: number, max: number): number[] {
  const indices = new Set<number>()
  while (indices.size < count && max > 0) {
    indices.add(Math.floor(Math.random() * max))
  }
  return Array.from(indices)
}

onMounted(() => {
  if (!globeContainer.value) return
  renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true })
  renderer.setClearColor(0x000000, 0) // transparent
  globeContainer.value.appendChild(renderer.domElement)
  createGlobe()
  resizeRenderer()
  window.addEventListener('resize', resizeRenderer)
  globeContainer.value.addEventListener('pointerdown', onPointerDown)
  globeContainer.value.addEventListener('pointerup', onPointerUp)
  globeContainer.value.addEventListener('pointermove', onPointerMove)
  globeContainer.value.addEventListener('click', onPointerClick)
  // Initialize pulsing nodes
  const nodeCount = getNodePositions().length
  pulsingNodes.value = pickRandomPulsingNodes(12, nodeCount)
  pulseInterval = window.setInterval(() => {
    pulsingNodes.value = pickRandomPulsingNodes(12, nodeCount)
  }, 2500)
  animate()
})

onBeforeUnmount(() => {
  if (animationId) cancelAnimationFrame(animationId)
  window.removeEventListener('resize', resizeRenderer)
  if (globeContainer.value) {
    globeContainer.value.removeEventListener('pointerdown', onPointerDown)
    globeContainer.value.removeEventListener('pointerup', onPointerUp)
    globeContainer.value.removeEventListener('pointermove', onPointerMove)
    globeContainer.value.removeEventListener('click', onPointerClick)
  }
  if (renderer && globeContainer.value) {
    globeContainer.value.removeChild(renderer.domElement)
  }
  if (pulseInterval) clearInterval(pulseInterval)
  renderer = null
  scene = null
  camera = null
  pointCloud = null
  raycaster = null
})

// Watch for prop changes (e.g., size, data)
watch(() => [props.width, props.height, props.pointCount, props.nodes, props.arcs], () => {
  if (scene && renderer && camera && globeContainer.value) {
    // Remove old scene
    while (scene.children.length > 0) {
      scene.remove(scene.children[0])
    }
    createGlobe()
    resizeRenderer()
    // Re-pick pulsing nodes for new geometry
    const nodeCount = getNodePositions().length
    pulsingNodes.value = pickRandomPulsingNodes(12, nodeCount)
  }
})
</script>

<style scoped>
.globe-canvas {
  min-width: 200px;
  min-height: 180px;
  margin: 0 auto;
  position: relative;
  z-index: 3;
  background: none;
  border-radius: 50%;
  box-shadow: none;
  cursor: grab;
}
.globe-canvas:active {
  cursor: grabbing;
}
@media (max-width: 600px) {
  .globe-canvas {
    height: 260px;
    min-height: 180px;
  }
}
</style> 