<template>
  <footer class="app-footer">
    <div class="footer-content container-padding">
      <div class="footer-main">
        <div class="footer-brand">
          <div class="footer-logo kubecloud-glow-blue d-flex align-center mb-4">
            <svg width="32" height="32" viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
              <ellipse cx="24" cy="28" rx="16" ry="10" fill="#1E293B" stroke="#2563EB" stroke-width="2.5"/>
              <ellipse cx="18" cy="22" rx="7" ry="6" fill="#1E293B" stroke="#3B82F6" stroke-width="2.5"/>
              <ellipse cx="30" cy="20" rx="6" ry="5" fill="#1E293B" stroke="#F97316" stroke-width="2.5"/>
              <ellipse cx="24" cy="28" rx="10" ry="7" fill="#1E293B" stroke="#3B82F6" stroke-width="2.5"/>
              <ellipse cx="24" cy="28" rx="5" ry="3.5" fill="#2563EB" fill-opacity="0.7"/>
            </svg>
            <span class="footer-title ms-2">KubeCloud</span>
          </div>
          <p class="footer-description">
            The most advanced Kubernetes management platform for modern cloud-native applications.
          </p>
          <div class="footer-social">
            <v-btn icon variant="text" color="primary" class="social-btn kubecloud-hover-blue">
              <v-icon>mdi-github</v-icon>
            </v-btn>
            <v-btn icon variant="text" color="primary" class="social-btn kubecloud-hover-blue">
              <v-icon>mdi-twitter</v-icon>
            </v-btn>
            <v-btn icon variant="text" color="primary" class="social-btn kubecloud-hover-blue">
              <v-icon>mdi-linkedin</v-icon>
            </v-btn>
            <v-btn icon variant="text" color="primary" class="social-btn kubecloud-hover-blue">
              <v-icon>mdi-discord</v-icon>
            </v-btn>
          </div>
        </div>
        
        <div class="footer-links">
          <div class="footer-section">
            <h3 class="footer-section-title">Product</h3>
            <ul class="footer-link-list">
              <li><router-link to="/features" class="footer-link kubecloud-hover-blue">Features</router-link></li>
              <li><router-link to="/pricing" class="footer-link kubecloud-hover-blue">Pricing</router-link></li>
              <li><router-link to="/usecases" class="footer-link kubecloud-hover-blue">Use Cases</router-link></li>
              <li><router-link to="/docs" class="footer-link kubecloud-hover-blue">Documentation</router-link></li>
            </ul>
          </div>
          
          <div class="footer-section">
            <h3 class="footer-section-title">Company</h3>
            <ul class="footer-link-list">
              <li><router-link to="/about" class="footer-link kubecloud-hover-blue">About</router-link></li>
              <li><router-link to="/careers" class="footer-link kubecloud-hover-blue">Careers</router-link></li>
              <li><router-link to="/blog" class="footer-link kubecloud-hover-blue">Blog</router-link></li>
              <li><router-link to="/contact" class="footer-link kubecloud-hover-blue">Contact</router-link></li>
            </ul>
          </div>
          
          <div class="footer-section">
            <h3 class="footer-section-title">Support</h3>
            <ul class="footer-link-list">
              <li><router-link to="/help" class="footer-link kubecloud-hover-blue">Help Center</router-link></li>
              <li><router-link to="/community" class="footer-link kubecloud-hover-blue">Community</router-link></li>
              <li><router-link to="/status" class="footer-link kubecloud-hover-blue">Status</router-link></li>
              <li><router-link to="/security" class="footer-link kubecloud-hover-blue">Security</router-link></li>
            </ul>
          </div>
          
          <div class="footer-section">
            <h3 class="footer-section-title">Legal</h3>
            <ul class="footer-link-list">
              <li><router-link to="/privacy" class="footer-link kubecloud-hover-blue">Privacy Policy</router-link></li>
              <li><router-link to="/terms" class="footer-link kubecloud-hover-blue">Terms of Service</router-link></li>
              <li><router-link to="/cookies" class="footer-link kubecloud-hover-blue">Cookie Policy</router-link></li>
              <li><router-link to="/compliance" class="footer-link kubecloud-hover-blue">Compliance</router-link></li>
            </ul>
          </div>
        </div>
      </div>
      
      <div class="footer-bottom">
        <div class="footer-copyright">
          <p>&copy; 2025 KubeCloud. All rights reserved.</p>
        </div>
        <div class="footer-newsletter">
          <v-btn 
            color="secondary" 
            variant="elevated" 
            to="/sign-up" 
            class="newsletter-btn btn-enhanced kubecloud-glow-orange"
          >
            Get Started Free
          </v-btn>
        </div>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
// Footer component logic
</script>

<style scoped>
.app-footer {
  background: var(--kubecloud-slate);
  border-top: 2px solid var(--kubecloud-blue);
  padding: 4rem 0 2rem 0;
  position: relative;
}

.footer-content {
  max-width: 1200px;
  margin: 0 auto;
}

.footer-main {
  display: grid;
  grid-template-columns: 1fr 2fr;
  gap: 4rem;
  margin-bottom: 3rem;
}

.footer-brand {
  max-width: 300px;
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--kubecloud-white);
  text-decoration: none;
  text-shadow: 0 0 2px var(--kubecloud-blue);
  transition: all var(--transition-normal);
}

.footer-logo {
  display: flex;
  align-items: center;
  font-family: 'Space Grotesk', 'Inter', sans-serif;
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--kubecloud-blue);
  text-shadow: 0 0 4px var(--kubecloud-blue);
}

.footer-title {
  color: var(--kubecloud-blue);
  font-weight: 700;
}

.footer-description {
  color: var(--kubecloud-light-gray);
  line-height: 1.6;
  margin-bottom: 2rem;
  font-size: 1rem;
}

.footer-social {
  display: flex;
  gap: 1rem;
}

.social-btn {
  background: rgba(37, 99, 235, 0.1);
  border: 1px solid rgba(37, 99, 235, 0.2);
  border-radius: 8px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.social-btn:hover {
  background: rgba(37, 99, 235, 0.2);
  border-color: var(--kubecloud-blue);
  transform: translateY(-1px);
  box-shadow: var(--shadow-kubecloud-blue);
}

.footer-links {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 2rem;
}

.footer-section-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--kubecloud-white);
  margin-bottom: 1.5rem;
  font-family: 'Space Grotesk', 'Inter', sans-serif;
}

.footer-link-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.footer-link-list li {
  margin-bottom: 0.75rem;
}

.footer-link {
  color: var(--kubecloud-light-gray);
  text-decoration: none;
  font-size: 0.95rem;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  display: inline-block;
}

.footer-link:hover {
  color: var(--kubecloud-blue);
  text-shadow: 0 0 1px rgba(37, 99, 235, 0.3);
}

.footer-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 2rem;
  border-top: 1px solid rgba(37, 99, 235, 0.2);
}

.footer-copyright {
  color: var(--kubecloud-gray);
  font-size: 0.9rem;
}

.newsletter-btn {
  box-shadow: var(--shadow-kubecloud-orange);
  color: var(--kubecloud-white) !important;
  font-weight: 600;
  padding: 0.75rem 1.5rem;
  border-radius: 12px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
}

.newsletter-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 0 15px rgba(249, 115, 22, 0.2), 0 0 30px rgba(249, 115, 22, 0.1);
}

/* Responsive Design */
@media (max-width: 960px) {
  .footer-main {
    grid-template-columns: 1fr;
    gap: 3rem;
  }
  
  .footer-links {
    grid-template-columns: repeat(2, 1fr);
    gap: 2rem;
  }
  
  .footer-bottom {
    flex-direction: column;
    gap: 1.5rem;
    text-align: center;
  }
}

@media (max-width: 600px) {
  .app-footer {
    padding: 3rem 0 1.5rem 0;
  }
  
  .footer-links {
    grid-template-columns: 1fr;
    gap: 2rem;
  }
  
  .footer-social {
    justify-content: center;
  }
  
  .footer-brand {
    text-align: center;
  }
}

.footer-logo:hover {
  text-shadow: 0 0 4px var(--kubecloud-blue);
}

.footer-link:hover {
  color: var(--kubecloud-blue);
  text-shadow: 0 0 1px rgba(37, 99, 235, 0.3);
}
</style>
