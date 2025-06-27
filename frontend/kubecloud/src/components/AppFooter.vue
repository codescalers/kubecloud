<template>
  <footer class="app-footer">
    <div class="footer-content">
      <div class="footer-brand">
        <router-link to="/" class="footer-logo" aria-label="KubeCloud logo">
          <svg width="32" height="32" viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
            <ellipse cx="24" cy="28" rx="16" ry="10" fill="#181A20" stroke="#3B82F6" stroke-width="2.5"/>
            <ellipse cx="18" cy="22" rx="7" ry="6" fill="#181A20" stroke="#60A5FA" stroke-width="2.5"/>
            <ellipse cx="30" cy="20" rx="6" ry="5" fill="#181A20" stroke="#93C5FD" stroke-width="2.5"/>
            <ellipse cx="24" cy="28" rx="10" ry="7" fill="#181A20" stroke="#3B82F6" stroke-width="2.5"/>
            <ellipse cx="24" cy="28" rx="5" ry="3.5" fill="#3B82F6" fill-opacity="0.7"/>
          </svg>
          <span class="footer-title">KubeCloud</span>
        </router-link>
      </div>
      <nav class="footer-links">
        <router-link to="/docs" class="footer-link">Docs</router-link>
        <router-link to="/pricing" class="footer-link">Pricing</router-link>
        <a href="https://github.com/neverhack/kubecloud" target="_blank" rel="noopener" class="footer-link">GitHub</a>
        <router-link to="/privacy" class="footer-link">Privacy</router-link>
      </nav>
    </div>
    <div class="footer-bottom">
      <span class="footer-copyright">© {{ new Date().getFullYear() }} KubeCloud. All rights reserved.</span>
    </div>
  </footer>
</template>

<script setup lang="ts">
// No logic needed
</script>

<style scoped>
.app-footer {
  width: 100%;
  background: rgba(10, 25, 47, 0.65); /* semi-transparent, matches nav */
  color: #e0e7ef;
  border: none;
  box-shadow: 0 2px 16px 0 rgba(33, 150, 243, 0.10); /* soft shadow for readability */
  padding: 2.5rem 0 1.2rem 0;
  margin-top: auto;
  backdrop-filter: blur(8px);
  transition: background 0.3s;
}

.footer-content {
  max-width: 1300px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 2rem;
  padding: 0 2.5rem;
}

.footer-brand {
  display: flex;
  align-items: center;
}

.footer-logo {
  display: flex;
  align-items: center;
  text-decoration: none;
  color: #60a5fa;
  font-size: 1.2rem;
  font-weight: 500;
  letter-spacing: 0.01em;
  transition: color 0.2s;
}

.footer-logo:hover {
  color: #38bdf8;
}

.footer-title {
  color: inherit;
  font-weight: inherit;
  letter-spacing: inherit;
  margin-left: 0.5rem;
  font-size: inherit;
  transition: color 0.2s;
}

.footer-links {
  display: flex;
  gap: 1.2rem;
  align-items: center;
}

.footer-link {
  color: #e0e7ef;
  text-decoration: none;
  font-size: 1.05rem;
  font-weight: 500;
  letter-spacing: 0.02em;
  transition: color 0.2s;
  position: relative;
  padding: 0.2rem 0;
  min-width: 0;
}

.footer-link:hover {
  color: #60a5fa;
}

.footer-link::after {
  content: '';
  display: block;
  height: 2px;
  width: 0;
  background: linear-gradient(90deg, #60a5fa 0%, #38bdf8 100%);
  transition: width 0.3s;
  position: absolute;
  left: 0;
  bottom: -2px;
}

.footer-link:hover::after {
  width: 100%;
}

.footer-bottom {
  width: 100%;
  text-align: center;
  margin-top: 2rem;
  color: #60a5fa;
  font-size: 0.95rem;
  opacity: 0.85;
}

.footer-copyright {
  opacity: 0.8;
}

/* Responsive Design */
@media (max-width: 900px) {
  .footer-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 1.5rem;
    padding: 0 1.2rem;
  }
  
  .footer-links {
    gap: 0.7rem;
  }
  
  .footer-title {
    font-size: 1rem;
  }
}

@media (max-width: 600px) {
  .app-footer {
    padding: 1.5rem 0 0.7rem 0;
  }
  
  .footer-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
    padding: 0 0.5rem;
  }
  
  .footer-links {
    gap: 0.7rem;
    flex-wrap: wrap;
  }
  
  .footer-title {
    font-size: 1rem;
  }
  
  .footer-bottom {
    margin-top: 1.5rem;
  }
}
</style>
