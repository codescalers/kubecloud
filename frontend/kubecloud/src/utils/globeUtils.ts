/**
 * Shared utilities for globe functionality
 */
import * as THREE from 'three'

// Constants
export const GLOBE_RADIUS = 3.6
export const GLOBE_COLORS = {
  BASE_SPHERE: 0x0f1e3a,
  BORDER_POINT: { r: 1.0, g: 1.0, b: 1.0 },
  LAND_POINT_BASE: { r: 0.3, g: 0.5, b: 0.8 },
  NODE_BASE: { r: 1.0, g: 0.8, b: 0.3 },
  NODE_HOVER: { r: 1.0, g: 1.0, b: 1.0 }
}

// Global land mask data
let landAlphaData: ImageData | null = null
let cachedBorderPoints: [number, number][] | null = null
let cachedBorderStep: number | null = null

/**
 * Load land mask data from world atlas
 */
export async function loadLandMask(): Promise<void> {
  try {
    await loadRasterFromWorldAtlas(1024)
  } catch (_e) {
    // Fallback: tiny pre-baked alpha PNG
    await new Promise<void>((resolve) => {
      const img = new Image()
      img.crossOrigin = 'anonymous'
      img.src = 'https://unpkg.com/world-atlas-land@1.0.0/alpha-1024.png'
      img.onload = () => {
        const c = document.createElement('canvas')
        c.width = img.width
        c.height = img.height
        const ctx = c.getContext('2d')!
        ctx.drawImage(img, 0, 0)
        landAlphaData = ctx.getImageData(0, 0, img.width, img.height)
        resolve()
      }
      img.onerror = () => resolve()
    })
  }
}

/**
 * Fetch Natural Earth land TopoJSON and rasterize with d3-geo
 */
async function loadRasterFromWorldAtlas(width: number = 1024): Promise<void> {
  // @ts-ignore - external URL import without types
  const d3 = await import(/* @vite-ignore */ 'https://cdn.jsdelivr.net/npm/d3-geo@3/+esm') as any
  // @ts-ignore - external URL import without types
  const topojson = await import(/* @vite-ignore */ 'https://cdn.jsdelivr.net/npm/topojson-client@3/+esm') as any
  const topoUrl = 'https://cdn.jsdelivr.net/npm/world-atlas@2/land-110m.json'
  const resp = await fetch(topoUrl)
  if (!resp.ok) throw new Error('Failed to fetch world-atlas land')
  const topo = await resp.json()
  const land = topojson.feature(topo, topo.objects.land)

  const aspect = 2 // equirectangular 2:1
  const w = width
  const h = Math.round(width / aspect)
  const canvas = document.createElement('canvas')
  canvas.width = w
  canvas.height = h
  const ctx = canvas.getContext('2d')!
  ctx.fillStyle = '#000'
  ctx.fillRect(0, 0, w, h)
  ctx.fillStyle = '#fff'

  const projection = d3.geoEquirectangular().fitSize([w, h], land)
  const path = d3.geoPath(projection, ctx as unknown as CanvasRenderingContext2D)
  ctx.beginPath()
  path(land as any)
  ctx.fill()
  landAlphaData = ctx.getImageData(0, 0, w, h)
}

/**
 * Calculate optimal stride based on globe size to prevent dots from being too dense
 */
export function calculateOptimalStride(width: number, height: number): number {
  const baseStride = 6
  const baseSize = 480 // Reference size for base stride
  const currentSize = Math.min(width, height)

  // For smaller globes, increase stride to reduce density
  // For larger globes, keep or slightly reduce stride for more detail
  const sizeRatio = currentSize / baseSize
  const stride = Math.max(2, Math.round(baseStride / Math.sqrt(sizeRatio)))

  return stride
}

/**
 * Convert latitude/longitude to 3D coordinates on sphere
 */
export function latLngToSphere(lat: number, lng: number, radius: number = GLOBE_RADIUS): THREE.Vector3 {
  const phi = (90 - lat) * (Math.PI / 180)
  const theta = (-lng) * (Math.PI / 180)
  const x = radius * Math.sin(phi) * Math.cos(theta)
  const y = radius * Math.cos(phi)
  const z = radius * Math.sin(phi) * Math.sin(theta)
  return new THREE.Vector3(x, y, z)
}

/**
 * Validate and filter node coordinates
 */
export function validateNodes(nodes: [number, number][] | undefined): [number, number][] {
  try {
    if (nodes && Array.isArray(nodes) && nodes.length > 0) {
      return nodes.filter(([lat, lng]) =>
        typeof lat === 'number' &&
        typeof lng === 'number' &&
        !isNaN(lat) &&
        !isNaN(lng) &&
        Math.abs(lat) <= 90 &&
        Math.abs(lng) <= 180
      )
    }
  } catch (_) {}
  return []
}

/**
 * Generate land dots from the alpha land mask
 */
export function generateLandDotsFromMask(stride: number = 3): [number, number][] {
  const points: [number, number][] = []
  if (!landAlphaData) return points
  const w = landAlphaData.width
  const h = landAlphaData.height
  const threshold = 127
  for (let y = 0; y < h; y += stride) {
    for (let x = 0; x < w; x += stride) {
      const idx = (x + y * w) * 4
      if (landAlphaData.data[idx] > threshold) {
        const lng = (x / w) * 360 - 180
        const lat = 90 - (y / h) * 180
        points.push([lat, lng])
      }
    }
  }
  return points
}

/**
 * Get continent border points using marching squares algorithm
 */
export function getContinentBorders(step: number = 3): [number, number][] {
  if (cachedBorderPoints && cachedBorderStep === step) return cachedBorderPoints
  if (!landAlphaData) return []
  const w = landAlphaData.width
  const h = landAlphaData.height
  const threshold = 127
  const sample = (x: number, y: number) => {
    const ix = ((x % w + w) % w) + Math.max(0, Math.min(h - 1, y)) * w
    return landAlphaData!.data[ix * 4] > threshold ? 1 : 0
  }
  const points: [number, number][] = []
  const addPoint = (px: number, py: number) => {
    const lng = (px / w) * 360 - 180
    const lat = 90 - (py / h) * 180
    points.push([lat, lng])
  }
  // marching squares edge interpolation helper
  const interp = (x1: number, y1: number, x2: number, y2: number, v1: number, v2: number) => {
    const t = 0.5 // simple midpoint (binary mask), good enough for our halftone
    return { x: x1 + (x2 - x1) * t, y: y1 + (y2 - y1) * t }
  }
  for (let y = 0; y < h; y += step) {
    for (let x = 0; x < w; x += step) {
      const v0 = sample(x, y)
      const v1 = sample(x + step, y)
      const v2 = sample(x + step, y + step)
      const v3 = sample(x, y + step)
      const idx = (v0 << 3) | (v1 << 2) | (v2 << 1) | v3
      if (idx === 0 || idx === 15) continue
      const x0 = x, y0 = y, x1p = x + step, y1p = y + step
      // edges: a=top(x0->x1p,y0), b=right(x1p,y0->y1p), c=bottom(x1p->x0,y1p), d=left(x0,y1p->y0)
      const a = interp(x0, y0, x1p, y0, v0, v1)
      const b = interp(x1p, y0, x1p, y1p, v1, v2)
      const c = interp(x1p, y1p, x0, y1p, v2, v3)
      const d = interp(x0, y1p, x0, y0, v3, v0)
      // handle cases; add a few samples along each segment for curvature look
      switch (idx) {
        case 1: case 14: addPoint((d.x + c.x) * 0.5, (d.y + c.y) * 0.5); break
        case 2: case 13: addPoint((b.x + c.x) * 0.5, (b.y + c.y) * 0.5); break
        case 3: case 12: addPoint((a.x + c.x) * 0.5, (a.y + c.y) * 0.5); break
        case 4: case 11: addPoint((a.x + b.x) * 0.5, (a.y + b.y) * 0.5); break
        case 5: case 10: addPoint((a.x + d.x) * 0.5, (a.y + d.y) * 0.5); addPoint((b.x + c.x) * 0.5, (b.y + c.y) * 0.5); break
        case 6: case 9: addPoint((b.x + d.x) * 0.5, (b.y + d.y) * 0.5); break
        case 7: case 8: addPoint((a.x + d.x) * 0.5, (a.y + d.y) * 0.5); break
      }
    }
  }
  cachedBorderPoints = points
  cachedBorderStep = step
  return points
}

/**
 * Create base sphere for the globe
 */
export function createBaseSphere(): THREE.Mesh {
  const baseSphereGeometry = new THREE.SphereGeometry(GLOBE_RADIUS, 64, 64)
  const baseSphereMaterial = new THREE.MeshBasicMaterial({ color: GLOBE_COLORS.BASE_SPHERE })
  return new THREE.Mesh(baseSphereGeometry, baseSphereMaterial)
}

/**
 * Create lighting setup for the globe
 */
export function createGlobeLighting(): THREE.Light[] {
  const ambientLight = new THREE.AmbientLight('#ffffff', 0.8)
  const directionalLight = new THREE.DirectionalLight('#ffffff', 0.6)
  directionalLight.position.set(5, 5, 5)
  return [ambientLight, directionalLight]
}

/**
 * Create background point cloud shader material
 */
export function createBackgroundPointMaterial(): THREE.ShaderMaterial {
  return new THREE.ShaderMaterial({
    uniforms: {},
    vertexShader: `
      attribute vec3 color;
      varying vec3 vColor;
      void main() {
        vColor = color;
        vec4 mvPosition = modelViewMatrix * vec4(position, 1.0);
        gl_PointSize = 1.9;
        gl_Position = projectionMatrix * mvPosition;
      }
    `,
    fragmentShader: `
      varying vec3 vColor;
      void main() {
        float distance = length(gl_PointCoord - vec2(0.5));
        if (distance > 0.5) discard;

        // Create circular shape with soft edges
        float alpha = 1.0 - (distance * 2.0);
        alpha = pow(alpha, 1.0);

        // Add subtle glow to background points
        vec3 glowColor = vColor * 0.15;
        vec3 finalColor = vColor + glowColor;
        finalColor = min(finalColor, vec3(1.0));

        gl_FragColor = vec4(finalColor, alpha * 1.0);
      }
    `,
    transparent: true,
    blending: THREE.AdditiveBlending
  })
}

/**
 * Create node point cloud shader material
 */
export function createNodePointMaterial(): THREE.ShaderMaterial {
  return new THREE.ShaderMaterial({
    uniforms: {},
    vertexShader: `
      precision mediump float;
      attribute vec3 color;
      varying vec3 vColor;
      void main() {
        vColor = color;
        vec3 pos = position + normalize(position) * 0.015;
        vec4 mvPosition = modelViewMatrix * vec4(pos, 1.0);
        float depth = -mvPosition.z;
        float size = 4.5;
        float atten = 100.0 / max(60.0, depth);
        gl_PointSize = clamp(size * atten, 3.0, 7.0);
        gl_Position = projectionMatrix * mvPosition;
      }
    `,
    fragmentShader: `
      precision mediump float;
      varying vec3 vColor;
      void main() {
        vec2 uv = gl_PointCoord - vec2(0.5);
        float r = length(uv);
        if (r > 0.5) discard;
        float alpha = smoothstep(0.5, 0.3, r);
        float core = smoothstep(0.2, 0.0, r);
        vec3 color = vColor * 0.8 + vec3(1.0) * 0.2;
        float finalAlpha = alpha * 0.9 + core * 0.1;
        gl_FragColor = vec4(min(color, vec3(1.0)), finalAlpha);
      }
    `,
    transparent: true,
    blending: THREE.NormalBlending,
    depthWrite: true,
    depthTest: true
  })
}

/**
 * Create background points geometry and colors
 */
export function createBackgroundPoints(stride: number): {
  positions: Float32Array
  colors: Float32Array
  count: number
} {
  const landPoints = generateLandDotsFromMask(stride)
  const positions = new Float32Array(landPoints.length * 3)
  const colors = new Float32Array(landPoints.length * 3)

  // Get continent border points for different coloring
  const continentBorders = getContinentBorders(stride)
  const borderSet = new Set(continentBorders.map(p => `${p[0]},${p[1]}`))

  for (let i = 0; i < landPoints.length; i++) {
    const [lat, lng] = landPoints[i]
    const pos = latLngToSphere(lat, lng)
    positions[i * 3] = pos.x
    positions[i * 3 + 1] = pos.y
    positions[i * 3 + 2] = pos.z

    // Enhanced colors for continent borders vs land points
    const isBorder = borderSet.has(`${lat},${lng}`)
    if (isBorder) {
      // Bright white for continent borders - more prominent
      colors[i * 3] = GLOBE_COLORS.BORDER_POINT.r
      colors[i * 3 + 1] = GLOBE_COLORS.BORDER_POINT.g
      colors[i * 3 + 2] = GLOBE_COLORS.BORDER_POINT.b
    } else {
      // Enhanced blue gradient for land points - more visible
      const c = 0.6 + 0.3 * (pos.y / GLOBE_RADIUS)
      colors[i * 3] = GLOBE_COLORS.LAND_POINT_BASE.r * c + 0.15
      colors[i * 3 + 1] = GLOBE_COLORS.LAND_POINT_BASE.g * c + 0.25
      colors[i * 3 + 2] = GLOBE_COLORS.LAND_POINT_BASE.b * c + 0.4
    }
  }

  return { positions, colors, count: landPoints.length }
}

/**
 * Create node points geometry and colors
 */
export function createNodePoints(nodes: [number, number][]): {
  positions: Float32Array
  colors: Float32Array
  baseColors: Float32Array
  count: number
} {
  const positions = new Float32Array(nodes.length * 3)
  const colors = new Float32Array(nodes.length * 3)
  const baseColors = new Float32Array(nodes.length * 3)

  // Always use hero accent blue for node color
  const colorObj = { r: 0.376, g: 0.647, b: 0.980 } // #60a5fa
  for (let i = 0; i < nodes.length; i++) {
    const [lat, lng] = nodes[i]
    const pos = latLngToSphere(lat, lng)
    positions[i * 3] = pos.x
    positions[i * 3 + 1] = pos.y
    positions[i * 3 + 2] = pos.z
    colors[i * 3] = colorObj.r
    colors[i * 3 + 1] = colorObj.g
    colors[i * 3 + 2] = colorObj.b
    baseColors[i * 3] = colorObj.r
    baseColors[i * 3 + 1] = colorObj.g
    baseColors[i * 3 + 2] = colorObj.b
  }

  return { positions, colors, baseColors, count: nodes.length }
}

/**
 * Create a complete globe scene with background and nodes
 */
export interface GlobeSceneOptions {
  width: number
  height: number
  nodes?: [number, number][]
  labels?: string[]
}

export interface GlobeScene {
  scene: THREE.Scene
  camera: THREE.PerspectiveCamera
  pointCloud: THREE.Points | null
  nodeLabels: string[]
}

export async function createGlobeScene(options: GlobeSceneOptions): Promise<GlobeScene> {
  const { width, height, nodes, labels } = options

  // Ensure land mask is loaded
  if (!landAlphaData) {
    await loadLandMask()
  }

  const scene = new THREE.Scene()
  scene.background = null

  const camera = new THREE.PerspectiveCamera(75, width / height, 0.1, 1000)
  camera.position.set(0, 0, 9)
  camera.lookAt(0, 0, 0)

  // Add base sphere
  const baseSphere = createBaseSphere()
  scene.add(baseSphere)

  // Add lighting
  const lights = createGlobeLighting()
  lights.forEach(light => scene.add(light))

  // Validate nodes
  const validNodes = validateNodes(nodes)

  // Create background points
  const optimalStride = calculateOptimalStride(width, height)
  const bgData = createBackgroundPoints(optimalStride)

  const bgGeometry = new THREE.BufferGeometry()
  bgGeometry.setAttribute('position', new THREE.BufferAttribute(bgData.positions, 3))
  bgGeometry.setAttribute('color', new THREE.BufferAttribute(bgData.colors, 3))

  const bgMaterial = createBackgroundPointMaterial()
  const bgPointCloud = new THREE.Points(bgGeometry, bgMaterial)
  scene.add(bgPointCloud)

  let pointCloud: THREE.Points | null = null
  let nodeLabels: string[] = []

  // Create node points if available
  if (validNodes.length > 0) {
    if (labels && labels.length === validNodes.length) {
      nodeLabels = [...labels]
    } else {
      nodeLabels = validNodes.map(([lat, lng]) => `(${lat.toFixed(2)}, ${lng.toFixed(2)})`)
    }

  const nodeData = createNodePoints(validNodes)
    const nodeGeometry = new THREE.BufferGeometry()
    nodeGeometry.setAttribute('position', new THREE.BufferAttribute(nodeData.positions, 3))
    nodeGeometry.setAttribute('color', new THREE.BufferAttribute(nodeData.colors, 3))
    nodeGeometry.setAttribute('baseColor', new THREE.BufferAttribute(nodeData.baseColors, 3))

    const nodeMaterial = createNodePointMaterial()
    pointCloud = new THREE.Points(nodeGeometry, nodeMaterial)
    scene.add(pointCloud)
  } else {
    // Empty point cloud for interaction
    const emptyGeometry = new THREE.BufferGeometry()
    emptyGeometry.setAttribute('position', new THREE.BufferAttribute(new Float32Array(0), 3))
    emptyGeometry.setAttribute('color', new THREE.BufferAttribute(new Float32Array(0), 3))
    pointCloud = new THREE.Points(emptyGeometry, new THREE.PointsMaterial({ size: 0.08, vertexColors: true }))
    scene.add(pointCloud)
  }

  return { scene, camera, pointCloud, nodeLabels }
}

/**
 * Process grid nodes for globe display
 */
export interface GridNode {
  nodeId?: string | number
  location?: {
    latitude?: number
    longitude?: number
  }
}

export interface ProcessedGridNodes {
  nodes: [number, number][]
  labels: string[]
}

export function processGridNodesForGlobe(gridNodes: GridNode[]): ProcessedGridNodes {
  if (!gridNodes.length) {
    return { nodes: [], labels: [] }
  }

  // Filter and convert nodes with valid coordinates
  const validNodes = gridNodes
    .filter(node =>
      node?.location?.latitude &&
      node?.location?.longitude &&
      typeof node.location.latitude === 'number' &&
      typeof node.location.longitude === 'number' &&
      !isNaN(node.location.latitude) &&
      !isNaN(node.location.longitude) &&
      Math.abs(node.location.latitude) <= 90 &&
      Math.abs(node.location.longitude) <= 180
    )

  const nodes = validNodes.map(node => [
    node.location!.latitude!,
    node.location!.longitude!
  ] as [number, number])

  const labels = validNodes.map((node, idx) =>
    node?.nodeId?.toString?.() || `Node ${idx + 1}`
  )

  return { nodes, labels }
}
