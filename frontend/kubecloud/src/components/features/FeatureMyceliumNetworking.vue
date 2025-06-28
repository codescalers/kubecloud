<template>
  <section class="feature-panel mycelium-networking">
    <div class="feature-content-row">
      <div class="feature-content feature-content-overlay">
        <h2 class="feature-title">Mycelium-Native Networking</h2>
        <p class="feature-description">
          Every pod gets a private, end-to-end encrypted Mycelium IP address. Use standard Kubernetes Network Policies to create whitelists and secure your workloads.
        </p>
        <div class="feature-benefits">
          <v-chip class="ma-1" color="white" variant="outlined" size="small">End-to-End Encryption</v-chip>
          <v-chip class="ma-1" color="white" variant="outlined" size="small">Network Policies</v-chip>
          <v-chip class="ma-1" color="white" variant="outlined" size="small">Private IPs</v-chip>
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

// Mycelium network constants
const POD_COUNT = 12
const NAMESPACE_COUNT = 3
const POLICY_COUNT = 4
const ENCRYPTION_PULSE_RATE = 0.003
const NETWORK_RADIUS = 2.8
const NAMESPACE_RADIUS = 1.8

// Sophisticated color palette matching globe
const PRIMARY_COLOR = 0x60a5fa // KubeCloud blue
const ENCRYPTION_COLOR = 0x34d399 // Green for encryption
const POLICY_COLOR = 0xfbbf24 // Yellow for policies
const NAMESPACE_COLOR = 0x8ecfff // Light blue for namespaces
const CONNECTION_COLOR = 0x38bdf8 // Cyan for connections

// Network elements
const pods: { mesh: THREE.Mesh, glow: THREE.Mesh, basePos: THREE.Vector3, phase: number, namespace: number, encrypted: boolean }[] = []
const namespaces: { mesh: THREE.Mesh, glow: THREE.Mesh, basePos: THREE.Vector3, phase: number }[] = []
const policies: { mesh: THREE.Mesh, glow: THREE.Mesh, basePos: THREE.Vector3, phase: number }[] = []
const connections: { line: THREE.Line, nodes: [THREE.Mesh, THREE.Mesh], encrypted: boolean }[] = []

// Animation states
let encryptionPulses: { mesh: THREE.Mesh, from: THREE.Vector3, to: THREE.Vector3, t: number, active: boolean }[] = []
let policyPulses: { mesh: THREE.Mesh, from: THREE.Vector3, to: THREE.Vector3, t: number, active: boolean }[] = []
let pulseCooldown = 0
let autoRotate = true
let rotationY = 0

// Hover state
const hoveredNode = shallowRef<{ mesh: THREE.Mesh, type: string, pos: { x: number, y: number } } | null>(null)

const nodeLabelStyle = computed(() => {
  if (!hoveredNode.value) return {}
  let color = '#60a5fa', borderColor = '#60a5fa33'
  if (hoveredNode.value.type === 'Pod') { color = '#34d399'; borderColor = '#34d39933' }
  if (hoveredNode.value.type === 'Namespace') { color = '#8ecfff'; borderColor = '#8ecfff33' }
  if (hoveredNode.value.type === 'Policy') { color = '#fbbf24'; borderColor = '#fbbf2433' }
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
    ...pods.map(p => ({ mesh: p.mesh, type: 'Pod' })),
    ...namespaces.map(n => ({ mesh: n.mesh, type: 'Namespace' })),
    ...policies.map(p => ({ mesh: p.mesh, type: 'Policy' }))
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

function createEncryptionPulse(from: THREE.Vector3, to: THREE.Vector3) {
  if (!scene) return null
  const geometry = new THREE.SphereGeometry(0.08, 16, 16)
  const material = new THREE.MeshBasicMaterial({
    color: ENCRYPTION_COLOR,
    transparent: true,
    opacity: 0.8
  })
  const mesh = new THREE.Mesh(geometry, material)
  mesh.position.copy(from)
  scene.add(mesh)
  return { mesh, from: from.clone(), to: to.clone(), t: 0, active: true }
}

function createPolicyPulse(from: THREE.Vector3, to: THREE.Vector3) {
  if (!scene) return null
  const geometry = new THREE.SphereGeometry(0.06, 12, 12)
  const material = new THREE.MeshBasicMaterial({
    color: POLICY_COLOR,
    transparent: true,
    opacity: 0.7
  })
  const mesh = new THREE.Mesh(geometry, material)
  mesh.position.copy(from)
  scene.add(mesh)
  return { mesh, from: from.clone(), to: to.clone(), t: 0, active: true }
}

onMounted(() => {
  if (!threeCanvas.value) return

  // Initialize Three.js with globe-compatible settings
  renderer = new THREE.WebGLRenderer({
    canvas: threeCanvas.value,
    alpha: true,
    antialias: true,
    powerPreference: "high-performance"
  })
  renderer.setSize(threeCanvas.value.clientWidth, threeCanvas.value.clientHeight, false)
  renderer.setClearColor(0x000000, 0) // Transparent background like globe

  scene = new THREE.Scene()
  camera = new THREE.PerspectiveCamera(
    60,
    threeCanvas.value.clientWidth / threeCanvas.value.clientHeight,
    0.1,
    1000
  )
  camera.position.z = 6

  // Create namespace nodes (inner ring)
  for (let i = 0; i < NAMESPACE_COUNT; i++) {
    const angle = (i / NAMESPACE_COUNT) * Math.PI * 2
    const basePos = new THREE.Vector3(
      NAMESPACE_RADIUS * Math.cos(angle),
      NAMESPACE_RADIUS * Math.sin(angle),
      (Math.random() - 0.5) * 0.3
    )

    const mesh = new THREE.Mesh(
      new THREE.SphereGeometry(0.14, 20, 20),
      new THREE.MeshBasicMaterial({
        color: NAMESPACE_COLOR,
        transparent: true,
        opacity: 0.9
      })
    )
    mesh.position.copy(basePos)
    scene.add(mesh)

    // Enhanced glow like globe
    const glowGeometry = new THREE.SphereGeometry(0.24, 20, 20)
    const glow = new THREE.Mesh(glowGeometry, new THREE.MeshBasicMaterial({
      color: NAMESPACE_COLOR,
      transparent: true,
      opacity: 0.2
    }))
    glow.position.copy(basePos)
    scene.add(glow)

    namespaces.push({
      mesh,
      glow,
      basePos: basePos.clone(),
      phase: Math.random() * Math.PI * 2
    })
  }

  // Create pod nodes (outer ring) - distributed across namespaces
  for (let i = 0; i < POD_COUNT; i++) {
    const namespaceIdx = i % NAMESPACE_COUNT
    const namespaceAngle = (namespaceIdx / NAMESPACE_COUNT) * Math.PI * 2
    const podAngle = (i / POD_COUNT) * Math.PI * 2
    const radius = NETWORK_RADIUS + (Math.random() - 0.5) * 0.4
    
    const basePos = new THREE.Vector3(
      radius * Math.cos(podAngle),
      radius * Math.sin(podAngle),
      (Math.random() - 0.5) * 0.4
    )

    const mesh = new THREE.Mesh(
      new THREE.SphereGeometry(0.08, 16, 16),
      new THREE.MeshBasicMaterial({
        color: ENCRYPTION_COLOR,
        transparent: true,
        opacity: 0.85
      })
    )
    mesh.position.copy(basePos)
    scene.add(mesh)

    // Subtle glow
    const glowGeometry = new THREE.SphereGeometry(0.16, 16, 16)
    const glow = new THREE.Mesh(glowGeometry, new THREE.MeshBasicMaterial({
      color: ENCRYPTION_COLOR,
      transparent: true,
      opacity: 0.15
    }))
    glow.position.copy(basePos)
    scene.add(glow)

    pods.push({
      mesh,
      glow,
      basePos: basePos.clone(),
      phase: Math.random() * Math.PI * 2,
      namespace: namespaceIdx,
      encrypted: Math.random() > 0.3 // 70% encrypted
    })
  }

  // Create policy nodes (strategic positions)
  for (let i = 0; i < POLICY_COUNT; i++) {
    const angle = (i / POLICY_COUNT) * Math.PI * 2 + Math.PI / 4
    const radius = 1.2
    const basePos = new THREE.Vector3(
      radius * Math.cos(angle),
      radius * Math.sin(angle),
      (Math.random() - 0.5) * 0.2
    )

    const mesh = new THREE.Mesh(
      new THREE.BoxGeometry(0.12, 0.12, 0.12),
      new THREE.MeshBasicMaterial({
        color: POLICY_COLOR,
        transparent: true,
        opacity: 0.8
      })
    )
    mesh.position.copy(basePos)
    scene.add(mesh)

    // Glow
    const glowGeometry = new THREE.BoxGeometry(0.18, 0.18, 0.18)
    const glow = new THREE.Mesh(glowGeometry, new THREE.MeshBasicMaterial({
      color: POLICY_COLOR,
      transparent: true,
      opacity: 0.15
    }))
    glow.position.copy(basePos)
    scene.add(glow)

    policies.push({
      mesh,
      glow,
      basePos: basePos.clone(),
      phase: Math.random() * Math.PI * 2
    })
  }

  // Create encrypted connections between pods
  for (let i = 0; i < POD_COUNT; i++) {
    for (let j = i + 1; j < POD_COUNT; j++) {
      // Connect pods within same namespace or with encryption
      if (pods[i].namespace === pods[j].namespace || (pods[i].encrypted && pods[j].encrypted)) {
        const lineGeom = new THREE.BufferGeometry().setFromPoints([
          pods[i].mesh.position,
          pods[j].mesh.position
        ])
        const lineMat = new THREE.LineBasicMaterial({
          color: CONNECTION_COLOR,
          transparent: true,
          opacity: 0.18
        })
        const line = new THREE.Line(lineGeom, lineMat)
        scene.add(line)
        connections.push({
          line,
          nodes: [pods[i].mesh, pods[j].mesh],
          encrypted: pods[i].encrypted && pods[j].encrypted
        })
      }
    }
  }

  // Connect pods to their namespaces
  pods.forEach(pod => {
    const namespace = namespaces[pod.namespace]
    const lineGeom = new THREE.BufferGeometry().setFromPoints([
      pod.mesh.position,
      namespace.mesh.position
    ])
    const lineMat = new THREE.LineBasicMaterial({
      color: NAMESPACE_COLOR,
      transparent: true,
      opacity: 0.12
    })
    const line = new THREE.Line(lineGeom, lineMat)
    if (scene) {
      scene.add(line)
    }
    connections.push({
      line,
      nodes: [pod.mesh, namespace.mesh],
      encrypted: false
    })
  })

  // Animation loop
  function animate(time = 0) {
    animationId = requestAnimationFrame(animate)
    const t = time * 0.001

    // Auto-rotate like globe
    if (autoRotate) {
      rotationY += 0.0015
    }

    // Animate namespaces
    namespaces.forEach((namespace, idx) => {
      namespace.mesh.position.x = namespace.basePos.x + Math.sin(t * 0.6 + namespace.phase) * 0.05
      namespace.mesh.position.y = namespace.basePos.y + Math.cos(t * 0.8 + namespace.phase) * 0.05
      namespace.mesh.position.z = namespace.basePos.z + Math.sin(t * 0.4 + namespace.phase) * 0.03
      namespace.glow.position.copy(namespace.mesh.position)
      
      // Rotate around center
      namespace.mesh.rotation.y = rotationY + idx * 0.5
    })

    // Animate pods with globe-like floating motion
    pods.forEach((pod, idx) => {
      pod.mesh.position.x = pod.basePos.x + Math.sin(t * 0.7 + pod.phase + idx * 0.3) * 0.08
      pod.mesh.position.y = pod.basePos.y + Math.cos(t * 0.9 + pod.phase + idx * 0.4) * 0.08
      pod.mesh.position.z = pod.basePos.z + Math.sin(t * 0.5 + pod.phase + idx * 0.2) * 0.05
      pod.glow.position.copy(pod.mesh.position)
      
      // Subtle rotation
      pod.mesh.rotation.y = rotationY * 0.5 + idx * 0.2
    })

    // Animate policies
    policies.forEach((policy, idx) => {
      policy.mesh.position.x = policy.basePos.x + Math.sin(t * 0.5 + policy.phase) * 0.04
      policy.mesh.position.y = policy.basePos.y + Math.cos(t * 0.7 + policy.phase) * 0.04
      policy.mesh.position.z = policy.basePos.z + Math.sin(t * 0.3 + policy.phase) * 0.02
      policy.glow.position.copy(policy.mesh.position)
      
      // Rotate policies
      policy.mesh.rotation.y = rotationY * 0.3 + idx * 0.8
    })

    // Update connections
    connections.forEach(conn => {
      conn.line.geometry.setFromPoints([conn.nodes[0].position, conn.nodes[1].position])
    })

    // Generate encryption pulses
    pulseCooldown -= 1
    if (pulseCooldown <= 0 && Math.random() < 0.02) {
      const encryptedPods = pods.filter(p => p.encrypted)
      if (encryptedPods.length >= 2) {
        const from = encryptedPods[Math.floor(Math.random() * encryptedPods.length)]
        const to = encryptedPods[Math.floor(Math.random() * encryptedPods.length)]
        if (from !== to) {
          const pulse = createEncryptionPulse(from.mesh.position, to.mesh.position)
          if (pulse) {
            encryptionPulses.push(pulse)
            pulseCooldown = 60
          }
        }
      }
    }

    // Generate policy pulses
    if (Math.random() < 0.01) {
      const policy = policies[Math.floor(Math.random() * policies.length)]
      const pod = pods[Math.floor(Math.random() * pods.length)]
      const pulse = createPolicyPulse(policy.mesh.position, pod.mesh.position)
      if (pulse) {
        policyPulses.push(pulse)
      }
    }

    // Animate encryption pulses
    encryptionPulses = encryptionPulses.filter(pulse => {
      pulse.t += ENCRYPTION_PULSE_RATE
      if (pulse.t >= 1) {
        if (scene) {
          scene.remove(pulse.mesh)
        }
        return false
      }
      pulse.mesh.position.lerpVectors(pulse.from, pulse.to, pulse.t)
      const material = pulse.mesh.material as THREE.MeshBasicMaterial
      material.opacity = 0.8 * (1 - pulse.t)
      return true
    })

    // Animate policy pulses
    policyPulses = policyPulses.filter(pulse => {
      pulse.t += 0.008
      if (pulse.t >= 1) {
        if (scene) {
          scene.remove(pulse.mesh)
        }
        return false
      }
      pulse.mesh.position.lerpVectors(pulse.from, pulse.to, pulse.t)
      const material = pulse.mesh.material as THREE.MeshBasicMaterial
      material.opacity = 0.7 * (1 - pulse.t)
      return true
    })

    renderer!.render(scene!, camera!)
  }

  animate()
})

onBeforeUnmount(() => {
  if (animationId) {
    cancelAnimationFrame(animationId)
  }
  if (renderer) {
    renderer.dispose()
  }
})
</script>

<style scoped>
.feature-panel {
  position: relative;
  padding: 2rem;
  border-radius: 16px;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.1) 0%, rgba(147, 197, 253, 0.05) 100%);
  border: 1px solid rgba(59, 130, 246, 0.2);
  backdrop-filter: blur(10px);
  margin-bottom: 2rem;
}

.feature-content-row {
  display: flex;
  align-items: center;
  gap: 3rem;
}

.feature-content {
  flex: 1;
  z-index: 2;
}

.feature-title {
  font-size: 2rem;
  font-weight: 700;
  color: #ffffff;
  margin-bottom: 1rem;
  background: linear-gradient(135deg, #60a5fa 0%, #34d399 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.feature-description {
  font-size: 1.1rem;
  color: #e5e7eb;
  line-height: 1.6;
  margin-bottom: 1.5rem;
}

.feature-benefits {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.feature-animation-with-glow {
  position: relative;
  width: 480px;
  height: 480px;
  flex-shrink: 0;
}

.feature-animation-glow {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, rgba(59, 130, 246, 0.15) 0%, rgba(59, 130, 246, 0.05) 50%, transparent 100%);
  border-radius: 50%;
  z-index: 1;
}

.feature-animation {
  position: relative;
  width: 100%;
  height: 100%;
  z-index: 2;
}

.three-canvas {
  width: 100%;
  height: 100%;
  border-radius: 12px;
}

.node-label {
  position: absolute;
  background: rgba(0, 0, 0, 0.8);
  color: #ffffff;
  padding: 0.5rem 0.75rem;
  border-radius: 6px;
  font-size: 0.85rem;
  font-weight: 500;
  pointer-events: none;
  z-index: 10;
  border: 1px solid;
  backdrop-filter: blur(8px);
  transform: translate(-50%, -100%);
  margin-top: -8px;
}

@media (max-width: 768px) {
  .feature-content-row {
    flex-direction: column;
    gap: 2rem;
  }
  
  .feature-animation-with-glow {
    width: 100%;
    height: 360px;
  }
}
</style>
