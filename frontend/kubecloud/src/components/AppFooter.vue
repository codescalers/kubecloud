<template>
  <footer class="app-footer minimalist-footer">
    <div class="footer-content minimalist-footer-content">
      <div class="footer-brand minimalist-footer-brand">
        <router-link to="/" class="footer-logo minimalist-footer-logo" aria-label="KubeCloud logo">
          <svg width="32" height="32" viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
            <ellipse cx="24" cy="28" rx="16" ry="10" fill="#181A20" stroke="#3B82F6" stroke-width="2.5"/>
            <ellipse cx="18" cy="22" rx="7" ry="6" fill="#181A20" stroke="#60A5FA" stroke-width="2.5"/>
            <ellipse cx="30" cy="20" rx="6" ry="5" fill="#181A20" stroke="#93C5FD" stroke-width="2.5"/>
            <ellipse cx="24" cy="28" rx="10" ry="7" fill="#181A20" stroke="#3B82F6" stroke-width="2.5"/>
            <ellipse cx="24" cy="28" rx="5" ry="3.5" fill="#3B82F6" fill-opacity="0.7"/>
          </svg>
          <span class="footer-title minimalist-footer-title ms-2">KubeCloud</span>
        </router-link>
      </div>
      <nav class="footer-links minimalist-footer-links">
        <router-link to="/docs" class="footer-link minimalist-footer-link">Docs</router-link>
        <router-link to="/pricing" class="footer-link minimalist-footer-link">Pricing</router-link>
        <a href="https://github.com/neverhack/kubecloud" target="_blank" rel="noopener" class="footer-link minimalist-footer-link">GitHub</a>
        <router-link to="/privacy" class="footer-link minimalist-footer-link">Privacy</router-link>
      </nav>
    </div>
    <div class="footer-bottom minimalist-footer-bottom">
      <span class="footer-copyright minimalist-footer-copyright">&copy; {{ new Date().getFullYear() }} KubeCloud. All rights reserved.</span>
    </div>
  </footer>
</template>

<script setup lang="ts">
// Footer component logic
</script>

<style scoped>
.app-footer.minimalist-footer {
  background: #0F172A;
  border-top: 2px solid #3B82F6;
  color: #fff;
  font-family: 'Space Grotesk', 'Inter', sans-serif;
  padding: 2.5rem 0 1.2rem 0;
  margin-top: 0;
  box-shadow: none;
}

.footer-content.minimalist-footer-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 2rem;
}

.footer-brand.minimalist-footer-brand {
  display: flex;
  align-items: center;
}

.footer-logo.minimalist-footer-logo {
  display: flex;
  align-items: center;
  text-decoration: none;
  color: #fff;
  font-size: 1.2rem;
  font-weight: 700;
  letter-spacing: 0.01em;
}

.footer-title.minimalist-footer-title {
  color: #fff;
  font-weight: 700;
  letter-spacing: 0.01em;
  margin-left: 0.5rem;
}

.footer-links.minimalist-footer-links {
  display: flex;
  gap: 2.5rem;
}

.footer-link.minimalist-footer-link {
  color: #3B82F6;
  text-decoration: none;
  font-size: 1.05rem;
  font-family: 'Space Grotesk', 'Inter', sans-serif;
  font-weight: 500;
  letter-spacing: 0.02em;
  transition: color 0.2s;
}

.footer-link.minimalist-footer-link:hover {
  color: #60A5FA;
}

.footer-bottom.minimalist-footer-bottom {
  text-align: center;
  margin-top: 2rem;
  color: #fff;
  font-size: 0.95rem;
  opacity: 0.7;
}

@media (max-width: 900px) {
  .footer-content.minimalist-footer-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 1.5rem;
    padding: 0 1rem;
  }
  .footer-links.minimalist-footer-links {
    gap: 1.2rem;
  }
}
@media (max-width: 600px) {
  .app-footer.minimalist-footer {
    padding: 1.5rem 0 0.7rem 0;
  }
  .footer-content.minimalist-footer-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
    padding: 0 0.5rem;
  }
  .footer-links.minimalist-footer-links {
    gap: 0.7rem;
  }
  .footer-title.minimalist-footer-title {
    font-size: 1rem;
  }
}
</style>
