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
import { createGlobeScene } from '../../utils/globeUtils'

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
  labels: { type: Array as () => string[] | undefined, default: undefined },
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
let nodeLabels: string[] = []

async function createGlobe() {
  if (!globeContainer.value || !renderer) return
  const globeScene = await createGlobeScene({
    width: props.width,
    height: props.height,
    nodes: props.nodes,
    labels: props.labels
  })
  scene = globeScene.scene
  camera = globeScene.camera
  pointCloud = globeScene.pointCloud
  nodeLabels = globeScene.nodeLabels
  raycaster = new THREE.Raycaster()
}

function animate() {
  if (!renderer || !scene || !camera) return
  animationId = requestAnimationFrame(animate)
  if (autoRotate && !isDragging) rotationY += 0.002
  scene.children.forEach(child => {
    if (child instanceof THREE.Points) {
      child.rotation.y = rotationY
      child.rotation.x = rotationX
    }
  })
  if (raycaster && pointCloud && globeContainer.value) {
    raycaster.setFromCamera(mouse, camera!)
    const intersects = raycaster.intersectObject(pointCloud)
    const geometry = pointCloud.geometry as THREE.BufferGeometry
    const colors = geometry.getAttribute('color') as THREE.BufferAttribute
    if (INTERSECTED !== null && colors) {
      const baseColors = geometry.getAttribute('baseColor') as THREE.BufferAttribute
      if (baseColors) {
        colors.setX(INTERSECTED, baseColors.array[INTERSECTED * 3])
        colors.setY(INTERSECTED, baseColors.array[INTERSECTED * 3 + 1])
        colors.setZ(INTERSECTED, baseColors.array[INTERSECTED * 3 + 2])
      }
    }
    if (intersects.length > 0) {
      const index = intersects[0].index
      if (index !== undefined && colors) {
        INTERSECTED = index
        colors.setX(index, 1.0)
        colors.setY(index, 1.0)
        colors.setZ(index, 1.0)
        emit('node-hover', index)
      }
    } else {
      INTERSECTED = null
    }
    if (colors) colors.needsUpdate = true
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
  const rect = globeContainer.value.getBoundingClientRect()
  mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1
  mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1
  if (isDragging) {
    rotationY += (event.clientX - lastMouseX) * 0.01
    rotationX += (event.clientY - lastMouseY) * 0.01
    lastMouseX = event.clientX
    lastMouseY = event.clientY
  }
}
function onPointerLeave() {
  INTERSECTED = null
}
function onPointerClick() {
  if (INTERSECTED !== null && pointCloud) {
    emit('node-click', INTERSECTED)
    const geometry = pointCloud.geometry as THREE.BufferGeometry
    const colors = geometry.getAttribute('color') as THREE.BufferAttribute
    const baseColors = geometry.getAttribute('baseColor') as THREE.BufferAttribute
    if (colors && baseColors) {
      const i = INTERSECTED
      let t = 0
      function flash() {
        t += 0.1
        const intensity = Math.max(0, 1 - t) * 2
        colors.setX(i, Math.min(1, baseColors.array[i * 3] + intensity))
        colors.setY(i, Math.min(1, baseColors.array[i * 3 + 1] + intensity))
        colors.setZ(i, Math.min(1, baseColors.array[i * 3 + 2] + intensity))
        colors.needsUpdate = true
        if (t < 1) {
          requestAnimationFrame(flash)
        } else {
          colors.setX(i, baseColors.array[i * 3])
          colors.setY(i, baseColors.array[i * 3 + 1])
          colors.setZ(i, baseColors.array[i * 3 + 2])
          colors.needsUpdate = true
        }
      }
      flash()
    }
  }
}



onMounted(async () => {
  if (!globeContainer.value) return
  renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true })
  renderer.setClearColor(0x000000, 0)
  renderer.setSize(props.width, props.height)
  globeContainer.value.appendChild(renderer.domElement)
  await createGlobe()
  resizeRenderer()
  window.addEventListener('resize', resizeRenderer)
  globeContainer.value.addEventListener('pointerdown', onPointerDown)
  globeContainer.value.addEventListener('pointerup', onPointerUp)
  globeContainer.value.addEventListener('pointermove', onPointerMove)
  globeContainer.value.addEventListener('pointerleave', onPointerLeave)
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
    globeContainer.value.removeEventListener('pointerleave', onPointerLeave)
    globeContainer.value.removeEventListener('click', onPointerClick)
  }
  if (renderer && globeContainer.value) globeContainer.value.removeChild(renderer.domElement)
  renderer = null
  scene = null
  camera = null
  pointCloud = null
  raycaster = null
})

watch(() => [props.width, props.height], (newVal, oldVal) => {
  if (oldVal && (newVal[0] !== oldVal[0] || newVal[1] !== oldVal[1])) resizeRenderer()
})
let nodeUpdateTimeout: number | null = null
watch(() => props.nodes, (newNodes, oldNodes) => {
  if (nodeUpdateTimeout) clearTimeout(nodeUpdateTimeout)
  nodeUpdateTimeout = window.setTimeout(async () => {
    if (scene && renderer && camera && globeContainer.value) {
      const oldLength = Array.isArray(oldNodes) ? oldNodes.length : 0
      const newLength = Array.isArray(newNodes) ? newNodes.length : 0
      if (oldLength !== newLength) {
        while (scene.children.length > 0) scene.remove(scene.children[0])
        await createGlobe()
      }
    }
  }, 100)
}, { deep: true })
</script>

<style scoped>
.globe-canvas {
  min-width: 300px;
  min-height: 300px;
  margin: 0 auto;
  position: relative;
  z-index: 3;
  background: none;
  border-radius: 50%;
  cursor: grab;
  overflow: visible;
  display: flex;
  align-items: center;
  justify-content: center;
}

.globe-canvas canvas {
  display: block;
  margin: 0 auto;
}
.globe-canvas:active {
  cursor: grabbing;
}
@media (max-width: 600px) {
  .globe-canvas {
    height: 400px;
    min-height: 300px;
  }
}
</style>
