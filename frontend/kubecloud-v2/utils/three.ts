import * as THREE from "three"

export function ensureThreeGlobal() {
  if (!window.THREE) {
    window.THREE = THREE
  }

  return THREE
}
