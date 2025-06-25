<template>
  <div ref="globeContainer" class="globe-canvas"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import * as THREE from 'three'

const globeContainer = ref<HTMLElement | null>(null)
let renderer: THREE.WebGLRenderer | null = null
let scene: THREE.Scene | null = null
let camera: THREE.PerspectiveCamera | null = null
let animationId: number | null = null
let pointCloud: THREE.Points | null = null
let arcMeshes: THREE.Line[] = []
let raycaster: THREE.Raycaster | null = null
let mouse = new THREE.Vector2()
let INTERSECTED: number | null = null
let isDragging = false
let lastMouseX = 0
let lastMouseY = 0
let rotationY = 0
let rotationX = 0
let autoRotate = true

const GLOBE_RADIUS = 2.3
const POINT_COLOR = '#60a5fa' // KubeCloud blue
const POINT_COLOR_HIGHLIGHT = '#fff'
const ARC_COLOR = '#38bdf8' // Cyan
const POINT_COUNT = 2500

// Generate fake point cloud data (lat, lng)
function randomLatLngPoints(count: number): [number, number][] {
  const points: [number, number][] = []
  for (let i = 0; i < count; i++) {
    // Uniformly distributed points on a sphere
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
const nodePositions: [number, number][] = randomLatLngPoints(POINT_COUNT)

function createGlobe() {
  scene = new THREE.Scene()
  scene.background = null
  camera = new THREE.PerspectiveCamera(60, 1, 0.1, 100)
  camera.position.z = 7

  // Point cloud geometry
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
    // Blue gradient: brighter at the top
    const c = 0.7 + 0.3 * (y / GLOBE_RADIUS)
    colors[i * 3] = 0.22 * c
    colors[i * 3 + 1] = 0.65 * c
    colors[i * 3 + 2] = 0.98 * c
  }
  geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3))
  geometry.setAttribute('color', new THREE.BufferAttribute(colors, 3))

  const material = new THREE.PointsMaterial({
    size: 0.08,
    vertexColors: true,
    transparent: true,
    opacity: 0.95
  })
  pointCloud = new THREE.Points(geometry, material)
  scene.add(pointCloud)

  // Animated arcs between random node pairs
  arcMeshes = []
  for (let i = 0; i < 10; i++) {
    const idxA = Math.floor(Math.random() * nodePositions.length)
    let idxB = Math.floor(Math.random() * nodePositions.length)
    while (idxB === idxA) idxB = Math.floor(Math.random() * nodePositions.length)
    const arc = createArc(nodePositions[idxA], nodePositions[idxB])
    if (arc) {
      scene.add(arc)
      arcMeshes.push(arc)
    }
  }

  // Lighting (minimal, just for arcs)
  const ambientLight = new THREE.AmbientLight('#ffffff', 0.5)
  scene.add(ambientLight)

  // Raycaster for hover/click
  raycaster = new THREE.Raycaster()
}

// Create a curved arc between two lat/lng points
function createArc(start: [number, number], end: [number, number]) {
  const [lat1, lng1] = start
  const [lat2, lng2] = end
  const phi1 = (90 - lat1) * (Math.PI / 180)
  const theta1 = (lng1 + 180) * (Math.PI / 180)
  const phi2 = (90 - lat2) * (Math.PI / 180)
  const theta2 = (lng2 + 180) * (Math.PI / 180)
  const p1 = new THREE.Vector3(
    GLOBE_RADIUS * Math.sin(phi1) * Math.cos(theta1),
    GLOBE_RADIUS * Math.cos(phi1),
    GLOBE_RADIUS * Math.sin(phi1) * Math.sin(theta1)
  )
  const p2 = new THREE.Vector3(
    GLOBE_RADIUS * Math.sin(phi2) * Math.cos(theta2),
    GLOBE_RADIUS * Math.cos(phi2),
    GLOBE_RADIUS * Math.sin(phi2) * Math.sin(theta2)
  )
  // Control point for curve (midpoint, raised above globe)
  const mid = p1.clone().lerp(p2, 0.5)
  mid.normalize().multiplyScalar(GLOBE_RADIUS * 1.25)
  const curve = new THREE.QuadraticBezierCurve3(p1, mid, p2)
  const points = curve.getPoints(50)
  const geometry = new THREE.BufferGeometry().setFromPoints(points)
  const material = new THREE.LineBasicMaterial({
    color: ARC_COLOR,
    linewidth: 2,
    transparent: true,
    opacity: 0.7
  })
  return new THREE.Line(geometry, material)
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
  }
  arcMeshes.forEach((arc, i) => {
    if (arc && arc.material && 'opacity' in arc.material) {
      (arc.material as THREE.LineBasicMaterial).opacity = 0.5 + 0.3 * Math.sin(Date.now() * 0.001 + i)
      arc.rotation.y = rotationY
      arc.rotation.x = rotationX
    }
  })
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
  const width = globeContainer.value.clientWidth
  const height = globeContainer.value.clientHeight
  renderer.setSize(width, height, false)
  camera.aspect = width / height
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
      const scale = 1 + 0.5 * Math.sin(Math.PI * t)
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
  renderer = null
  scene = null
  camera = null
  pointCloud = null
  arcMeshes = []
  raycaster = null
})
</script>

<style scoped>
.globe-canvas {
  width: 100%;
  height: 480px;
  min-height: 320px;
  max-width: 600px;
  margin: 0 auto;
  position: relative;
  z-index: 3;
  background: none;
  border-radius: 50%;
  box-shadow: none;
  /* No pointer-events: none, we want interactivity */
}
@media (max-width: 600px) {
  .globe-canvas {
    height: 260px;
    min-height: 180px;
  }
}
</style> 