<template>
  <div class="unified-background" :class="theme">
    <!-- Floating Particles -->
    <div class="background-particles">
      <div class="particle particle-1"></div>
      <div class="particle particle-2"></div>
      <div class="particle particle-3"></div>
      <div class="particle particle-4"></div>
      <div class="particle particle-5"></div>
      <div class="particle particle-6"></div>
      <div class="particle particle-7"></div>
      <div class="particle particle-8"></div>
    </div>
    
    <!-- Grid Lines -->
    <div class="background-grid-lines"></div>
    
    <!-- Network Connections -->
    <div class="background-connections">
      <div class="connection connection-1"></div>
      <div class="connection connection-2"></div>
      <div class="connection connection-3"></div>
      <div class="connection connection-4"></div>
    </div>
    
    <!-- Energy Waves -->
    <div class="background-energy-waves">
      <div class="energy-wave wave-1"></div>
      <div class="energy-wave wave-2"></div>
      <div class="energy-wave wave-3"></div>
    </div>
    
    <!-- Ambient Lights -->
    <div class="background-ambient-lights">
      <div class="light-spot light-1"></div>
      <div class="light-spot light-2"></div>
      <div class="light-spot light-3"></div>
      <div class="light-spot light-4"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  theme?: 'default' | 'home' | 'features' | 'pricing' | 'use-cases' | 'docs' | 'reserve' | 'dashboard'
}

withDefaults(defineProps<Props>(), {
  theme: 'default'
})
</script>

<style scoped>
.unified-background {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 0;
  pointer-events: none;
  background: linear-gradient(135deg, #0F172A 0%, #1E293B 50%, #334155 100%);
}

/* Floating Particles */
.background-particles {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
}

.particle {
  position: absolute;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(79, 70, 229, 0.3) 0%, rgba(79, 70, 229, 0.1) 50%, transparent 100%);
  animation: float-particle 20s ease-in-out infinite;
}

.particle-1 {
  width: 120px;
  height: 120px;
  top: 10%;
  left: 10%;
  animation-delay: 0s;
}

.particle-2 {
  width: 80px;
  height: 80px;
  top: 20%;
  right: 15%;
  background: radial-gradient(circle, rgba(234, 88, 12, 0.2) 0%, rgba(234, 88, 12, 0.1) 50%, transparent 100%);
  animation-delay: -3s;
}

.particle-3 {
  width: 100px;
  height: 100px;
  bottom: 20%;
  left: 20%;
  background: radial-gradient(circle, rgba(8, 145, 178, 0.25) 0%, rgba(8, 145, 178, 0.1) 50%, transparent 100%);
  animation-delay: -6s;
}

.particle-4 {
  width: 60px;
  height: 60px;
  top: 60%;
  right: 25%;
  background: radial-gradient(circle, rgba(79, 70, 229, 0.2) 0%, rgba(79, 70, 229, 0.05) 50%, transparent 100%);
  animation-delay: -9s;
}

.particle-5 {
  width: 90px;
  height: 90px;
  bottom: 10%;
  right: 10%;
  background: radial-gradient(circle, rgba(234, 88, 12, 0.15) 0%, rgba(234, 88, 12, 0.05) 50%, transparent 100%);
  animation-delay: -12s;
}

.particle-6 {
  width: 70px;
  height: 70px;
  top: 40%;
  left: 5%;
  background: radial-gradient(circle, rgba(8, 145, 178, 0.2) 0%, rgba(8, 145, 178, 0.05) 50%, transparent 100%);
  animation-delay: -15s;
}

.particle-7 {
  width: 85px;
  height: 85px;
  top: 70%;
  left: 60%;
  background: radial-gradient(circle, rgba(124, 58, 237, 0.18) 0%, rgba(124, 58, 237, 0.06) 50%, transparent 100%);
  animation-delay: -18s;
}

.particle-8 {
  width: 55px;
  height: 55px;
  top: 15%;
  left: 70%;
  background: radial-gradient(circle, rgba(244, 114, 182, 0.16) 0%, rgba(244, 114, 182, 0.05) 50%, transparent 100%);
  animation-delay: -21s;
}

@keyframes float-particle {
  0%, 100% { 
    transform: translateY(0px) translateX(0px) scale(1);
    opacity: 0.3;
  }
  25% { 
    transform: translateY(-20px) translateX(10px) scale(1.1);
    opacity: 0.5;
  }
  50% { 
    transform: translateY(-10px) translateX(-15px) scale(0.9);
    opacity: 0.4;
  }
  75% { 
    transform: translateY(-30px) translateX(5px) scale(1.05);
    opacity: 0.6;
  }
}

/* Grid Lines */
.background-grid-lines {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-image: 
    linear-gradient(rgba(79, 70, 229, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(79, 70, 229, 0.03) 1px, transparent 1px);
  background-size: 50px 50px;
  animation: grid-move 25s linear infinite;
}

@keyframes grid-move {
  0% { transform: translate(0, 0); }
  100% { transform: translate(50px, 50px); }
}

/* Network Connections */
.background-connections {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
}

.connection {
  position: absolute;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(79, 70, 229, 0.2), transparent);
  animation: connection-pulse 10s ease-in-out infinite;
}

.connection-1 {
  top: 30%;
  left: 20%;
  width: 200px;
  transform: rotate(45deg);
  animation-delay: 0s;
}

.connection-2 {
  top: 60%;
  right: 30%;
  width: 150px;
  transform: rotate(-30deg);
  animation-delay: -3s;
}

.connection-3 {
  bottom: 40%;
  left: 50%;
  width: 180px;
  transform: rotate(15deg);
  animation-delay: -6s;
}

.connection-4 {
  top: 45%;
  right: 10%;
  width: 120px;
  transform: rotate(60deg);
  animation-delay: -9s;
}

@keyframes connection-pulse {
  0%, 100% { 
    opacity: 0.1;
    transform: scaleX(0.8);
  }
  50% { 
    opacity: 0.4;
    transform: scaleX(1.2);
  }
}

/* Energy Waves */
.background-energy-waves {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
}

.energy-wave {
  position: absolute;
  height: 2px;
  background: linear-gradient(90deg, transparent, rgba(124, 58, 237, 0.3), transparent);
  animation: wave-flow 15s ease-in-out infinite;
}

.wave-1 {
  top: 25%;
  left: 0;
  right: 0;
  animation-delay: 0s;
}

.wave-2 {
  top: 55%;
  left: 0;
  right: 0;
  animation-delay: -5s;
}

.wave-3 {
  top: 85%;
  left: 0;
  right: 0;
  animation-delay: -10s;
}

@keyframes wave-flow {
  0%, 100% { 
    opacity: 0.1;
    transform: scaleX(0.5);
  }
  50% { 
    opacity: 0.3;
    transform: scaleX(1);
  }
}

/* Ambient Lights */
.background-ambient-lights {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
}

.light-spot {
  position: absolute;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(124, 58, 237, 0.1) 0%, transparent 70%);
  animation: light-pulse 8s ease-in-out infinite;
}

.light-1 {
  width: 300px;
  height: 300px;
  top: 10%;
  right: 10%;
  animation-delay: 0s;
}

.light-2 {
  width: 200px;
  height: 200px;
  bottom: 20%;
  left: 15%;
  background: radial-gradient(circle, rgba(234, 88, 12, 0.08) 0%, transparent 70%);
  animation-delay: -2s;
}

.light-3 {
  width: 250px;
  height: 250px;
  top: 50%;
  left: 5%;
  background: radial-gradient(circle, rgba(8, 145, 178, 0.06) 0%, transparent 70%);
  animation-delay: -4s;
}

.light-4 {
  width: 180px;
  height: 180px;
  bottom: 10%;
  right: 20%;
  background: radial-gradient(circle, rgba(244, 114, 182, 0.07) 0%, transparent 70%);
  animation-delay: -6s;
}

@keyframes light-pulse {
  0%, 100% { 
    opacity: 0.3;
    transform: scale(1);
  }
  50% { 
    opacity: 0.6;
    transform: scale(1.1);
  }
}

/* Theme Variations */
.unified-background.home {
  background: linear-gradient(135deg, #0F172A 0%, #1E293B 50%, #334155 100%);
}

.unified-background.features {
  background: linear-gradient(135deg, #0F172A 0%, #1E293B 50%, #334155 100%);
}

.unified-background.pricing {
  background: linear-gradient(135deg, #0F172A 0%, #1E293B 50%, #334155 100%);
}

.unified-background.use-cases {
  background: linear-gradient(135deg, #0F172A 0%, #1E293B 50%, #334155 100%);
}

.unified-background.docs {
  background: linear-gradient(135deg, #0F172A 0%, #1E293B 50%, #334155 100%);
}

.unified-background.reserve {
  background: linear-gradient(135deg, #0F172A 0%, #1E293B 50%, #334155 100%);
}

.unified-background.dashboard {
  background: linear-gradient(135deg, #0F172A 0%, #1E293B 50%, #334155 100%);
}

/* Responsive Design */
@media (max-width: 768px) {
  .particle {
    animation-duration: 15s;
  }
  
  .background-grid-lines {
    background-size: 30px 30px;
    animation-duration: 20s;
  }
  
  .connection {
    animation-duration: 8s;
  }
  
  .energy-wave {
    animation-duration: 12s;
  }
  
  .light-spot {
    animation-duration: 6s;
  }
}

/* Accessibility */
@media (prefers-reduced-motion: reduce) {
  .particle,
  .background-grid-lines,
  .connection,
  .energy-wave,
  .light-spot {
    animation: none;
    opacity: 0.1;
  }
}
</style> 