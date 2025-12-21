<template>
  <v-container fluid class="pa-0" style="background: linear-gradient(120deg, #0a192f 60%, #1e293b 100%), radial-gradient(ellipse at 70% 30%, #60a5fa33 0%, #0a192f 80%); min-height: 100vh;">
    <!-- Hero Section -->
    <v-row no-gutters class="hero-section mx-auto" style="max-width: 1100px; margin-top:4rem">
      <v-col cols="12">
        <v-container class="pt-16 pb-4">
          <v-row justify="center">
            <v-col cols="12" md="10" lg="8" class="text-center pt-4 pt-md-16">
              <h1 class="hero-title">
                Privacy Policy
              </h1>
              <p class="text-h6 text-blue-lighten-2 line-height-1-7 opacity-92">
                Last Updated: October 27th, 2025
              </p>
            </v-col>
          </v-row>
        </v-container>
      </v-col>
    </v-row>

    <!-- Content Section -->
    <v-container fluid class="pb-16">
      <v-row justify="center">
        <v-col cols="12" md="10" lg="8">
          <v-card class="privacy-card pa-8 pa-md-12 rounded-xl" variant="tonal">
            <div class="privacy-content">
              <p class="intro-text">
                This Privacy Policy describes how TF HUB Limited ("we", "us", or "our") collects, uses, and protects your personal information when you use the MyceliumCloud platform (the "Service").
              </p>
              <p class="intro-text">
                By using the Service, you agree to the terms of this Privacy Policy.
              </p>

              <section class="privacy-section">
                <h2>1. Information We Collect</h2>
                <p>We may collect the following information when you use the Service:</p>
                <ul>
                  <li><strong>Account Information:</strong> email address, username</li>
                  <li><strong>Usage & Log Data:</strong> IP address, timestamps, cluster identifiers, and operational logs related to workloads (for security and platform functionality)</li>
                  <li><strong>Payment Information:</strong> billing details and transaction records processed by Stripe (we do not store your full payment information)</li>
                </ul>
              </section>

              <section class="privacy-section">
                <h2>2. How We Use Your Information</h2>
                <p>We use collected data to:</p>
                <ul>
                  <li>operate and maintain the Service</li>
                  <li>send essential account or billing emails (via SendGrid)</li>
                  <li>monitor platform performance and security</li>
                  <li>comply with legal obligations</li>
                </ul>
              </section>

              <section class="privacy-section">
                <h2>3. Third-Party Providers</h2>
                <p>We use third parties to support the Service, including:</p>
                <ul>
                  <li>Stripe (payments)</li>
                  <li>SendGrid (email delivery)</li>
                </ul>
                <p>These providers may process your data according to their own privacy policies.</p>
              </section>

              <section class="privacy-section">
                <h2>4. Data Retention</h2>
                <p>We retain your data for as long as your account is active or as required to operate the Service. We may retain logs for security and compliance purposes. You may request account deletion at any time.</p>
              </section>

              <section class="privacy-section">
                <h2>5. Data Requests (Access & Deletion)</h2>
                <p>You may request access to or deletion of your personal data by contacting us at:</p>
                <p><a href="mailto:info@threefold.io" class="contact-link">info@threefold.io</a></p>
              </section>

              <section class="privacy-section">
                <h2>6. Cookies & Analytics</h2>
                <p>We may use cookies or analytics tools in the future to improve the Service. If we do, we will update this policy accordingly.</p>
              </section>

              <section class="privacy-section">
                <h2>7. Security</h2>
                <p>We take reasonable technical measures to protect your data, but no system is completely secure. You are responsible for maintaining control of your login credentials.</p>
              </section>

              <section class="privacy-section">
                <h2>8. International Data Transfers</h2>
                <p>Your data may be processed in other jurisdictions where our providers operate. By using the Service, you consent to such transfers.</p>
              </section>

              <section class="privacy-section">
                <h2>9. Changes to This Policy</h2>
                <p>We may update this Privacy Policy from time to time. Continued use of the Service after updates constitutes acceptance of the revised Policy.</p>
              </section>

              <section class="privacy-section">
                <h2>10. Contact</h2>
                <p>For questions or concerns, you may contact us at:</p>
                <p><strong>Email:</strong> <a href="mailto:info@threefold.io" class="contact-link">info@threefold.io</a></p>
                <p><strong>Address:</strong> Intershore Chambers, Road Town, Tortola, British Virgin Islands</p>
              </section>
            </div>
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </v-container>
</template>

<script setup lang="ts">
// No script logic needed for static content
</script>

<style scoped>
.hero-title {
  font-size: clamp(2.5rem, 6vw, 3.5rem);
  font-weight: 500;
  color: #fff;
  letter-spacing: -1px;
  line-height: 1.1;
  margin-bottom: 1rem;
}

.hero-section {
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  z-index: 2;
}

.privacy-card {
  background: rgba(10, 25, 47, 0.8);
  backdrop-filter: blur(8px);
  border: 1px solid rgba(96, 165, 250, 0.1);
}

.privacy-content {
  color: #e0e7ef;
  line-height: 1.8;
  font-size: 1rem;
}

.intro-text {
  font-size: 1.1rem;
  margin-bottom: 1.5rem;
  color: #e0e7ef;
  opacity: 0.95;
}

.privacy-section {
  margin-top: 3rem;
  margin-bottom: 2rem;
}

.privacy-section:first-of-type {
  margin-top: 2rem;
}

.privacy-section h2 {
  font-size: 1.75rem;
  font-weight: 600;
  color: #60a5fa;
  margin-bottom: 1.5rem;
  margin-top: 2.5rem;
  scroll-margin-top: 100px;
}

.privacy-section h2:first-child {
  margin-top: 0;
}

.privacy-section p {
  margin-bottom: 1.25rem;
  color: #e0e7ef;
  opacity: 0.9;
  line-height: 1.8;
}

.privacy-section ul {
  margin-left: 1.5rem;
  margin-bottom: 1.5rem;
  padding-left: 0;
}

.privacy-section li {
  margin-bottom: 0.75rem;
  color: #e0e7ef;
  opacity: 0.9;
  line-height: 1.7;
  list-style-type: disc;
}

.privacy-section strong {
  color: #60a5fa;
  font-weight: 600;
}

.contact-link {
  color: #60a5fa;
  text-decoration: none;
  transition: color 0.2s;
}

.contact-link:hover {
  color: #38bdf8;
  text-decoration: underline;
}

@media (max-width: 900px) {
  .hero-section {
    min-height: 30vh;
  }

  .privacy-card {
    padding: 2rem !important;
  }

  .privacy-section h2 {
    font-size: 1.5rem;
  }
}

@media (max-width: 600px) {
  .hero-section {
    min-height: 25vh;
  }

  .privacy-card {
    padding: 1.5rem !important;
  }

  .privacy-content {
    font-size: 0.95rem;
  }

  .privacy-section h2 {
    font-size: 1.35rem;
  }
}
</style>

