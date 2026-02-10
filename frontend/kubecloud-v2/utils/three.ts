import * as THREE from "three"

export function ensureThreeGlobal() {
  if ($meta.client && !window.THREE) {
    window.THREE = THREE
  }

  return THREE
}

export function applyVanta(type: "dots" | "globe", el: HTMLElement) {
  if (!$meta.client) {
    return
  }

  if (type === "dots") {
    applyVantaDots(el)
  }
  else if (type === "globe") {
    applyVantaGlobe(el)
  }
}

function applyVantaDots(el: HTMLElement) {
  window?.VANTA?.DOTS?.({
    el,
    mouseControls: true,
    touchControls: true,
    gyroControls: false,
    minHeight: 200,
    minWidth: 200,
    scale: 1,
    scaleMobile: 1,
    color: 0x2B3951,
    color2: 0x2B3951,
    backgroundColor: 0x0A192F,
    size: 5,
    spacing: 70,
    showLines: true,
  })
}

function applyVantaGlobe(el: HTMLElement) {
  window?.VANTA?.GLOBE?.({
    el,
    mouseControls: true,
    touchControls: true,
    gyroControls: true,
    minHeight: 200,
    minWidth: 200,
    scale: 1,
    scaleMobile: 1,
    color: 0x2B3951,
    color2: 0x2B3951,
    size: 0.75,
    backgroundColor: 0x0A192F,
  })
}
