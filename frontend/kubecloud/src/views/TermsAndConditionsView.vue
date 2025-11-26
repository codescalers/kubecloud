<template>
  <v-container fluid class="pa-0" style="background: linear-gradient(120deg, #0a192f 60%, #1e293b 100%), radial-gradient(ellipse at 70% 30%, #60a5fa33 0%, #0a192f 80%); min-height: 100vh;">
    <!-- Hero Section -->
    <v-row no-gutters class="hero-section mx-auto" style="max-width: 1100px; margin-top:4rem">
      <v-col cols="12">
        <v-container class="pt-16 pb-4">
          <v-row justify="center">
            <v-col cols="12" md="10" lg="8" class="text-center pt-4 pt-md-16">
              <h1 class="hero-title">
                Terms & Conditions
              </h1>
              <p class="text-h6 text-blue-lighten-2 line-height-1-7 opacity-92">
                Last Updated: November 20, 2025
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
          <v-card class="terms-card pa-8 pa-md-12 rounded-xl" variant="tonal">
            <div class="terms-content">
              <p class="intro-text">
                These Terms and Conditions ("Terms") govern your access to and use of the Mycelium Cloud platform and related services (the "Service"), operated by TF HUB Limited, a company organized under the laws of the British Virgin Islands ("we", "us", "our").
              </p>
              <p class="intro-text">
                By creating an account, deploying workloads, or using the Service in any way, you agree to be bound by these Terms.
              </p>
              <p class="intro-text">
                If you do not agree, you must discontinue use of the Service.
              </p>

              <section class="terms-section">
                <h2>1. Definitions</h2>
                <ul>
                  <li><strong>"Cluster"</strong> means a Kubernetes cluster provisioned or managed through the Service.</li>
                  <li><strong>"Node"</strong> means any virtual machine or server – centralized or decentralized – participating in a Cluster.</li>
                  <li><strong>"Workload"</strong> means any application, container, or process deployed on a Cluster.</li>
                  <li><strong>"TFT"</strong> (ThreeFold Token) means a digital utility token used internally for measuring and accounting resource consumption on the Service. TFT has no redemption, ownership, or financial rights.</li>
                  <li><strong>"Credits"</strong> means compute credits purchased in fiat and internally converted into TFT for usage accounting.</li>
                  <li><strong>"Node Provider"</strong> means an independent third party supplying decentralized compute resources.</li>
                  <li><strong>"Control Plane"</strong> means the Kubernetes orchestration layer managed by the Service.</li>
                  <li><strong>"Customer Data"</strong> means any data, configuration, image, or workload submitted to the Service.</li>
                </ul>
              </section>

              <section class="terms-section">
                <h2>2. About the Service</h2>
                <p>Mycelium Cloud provides tools for deploying, managing, and scaling Kubernetes clusters on centralized and decentralized infrastructure.</p>
                <p>The Service includes:</p>
                <ul>
                  <li>Kubernetes cluster provisioning</li>
                  <li>Node and workload orchestration</li>
                  <li>Monitoring and metrics</li>
                  <li>Access control and API usage</li>
                  <li>Billing and usage tracking</li>
                </ul>
                <p>We may enhance, modify, or discontinue features at any time.</p>
              </section>

              <section class="terms-section">
                <h2>3. Eligibility</h2>
                <p>You must be legally authorized to enter binding contracts in your jurisdiction. If you use the Service on behalf of an entity, you represent that you have authority to bind that entity to these Terms.</p>
              </section>

              <section class="terms-section">
                <h2>4. Accounts</h2>
                <p>You must create an account to access the Service. You agree to:</p>
                <ul>
                  <li>provide accurate information,</li>
                  <li>maintain confidentiality of your credentials,</li>
                  <li>be responsible for all actions taken through your account.</li>
                </ul>
                <p>We may suspend or terminate accounts that violate these Terms or create security, legal, or operational risk.</p>
              </section>

              <section class="terms-section">
                <h2>5. Acceptable Use Policy (AUP)</h2>
                <p>You must not use the Service for any unlawful, harmful, or abusive activity.</p>
                <p>The following activities are strictly prohibited:</p>

                <h3>5.1 Security & Network Abuse</h3>
                <ul>
                  <li>DDoS attacks, port scanning, or malicious probing</li>
                  <li>Hosting malware, ransomware, spyware, or botnets</li>
                  <li>Using the platform as a VPN, proxy, anonymizer, or TOR exit node</li>
                  <li>Disrupting or degrading the infrastructure</li>
                </ul>

                <h3>5.2 Resource Abuse</h3>
                <ul>
                  <li>Cryptocurrency mining of any kind</li>
                  <li>CPU/GPU-intensive workloads intended to exploit pricing</li>
                  <li>Unbounded fork bombs or uncontrolled resource loops</li>
                </ul>

                <h3>5.3 Legal & Content Restrictions</h3>
                <ul>
                  <li>Hosting illegal content</li>
                  <li>Storage or processing of highly regulated data (PCI, PHI, HIPAA) unless you have ensured compliance</li>
                  <li>Copyright infringement</li>
                </ul>

                <h3>5.4 High-Risk Use</h3>
                <p>The Service may not be used for life-critical systems, medical devices, autonomous vehicles, aviation, or nuclear operations.</p>
                <p>Violation of this AUP may result in immediate suspension without refund.</p>
              </section>

              <section class="terms-section">
                <h2>6. Shared Responsibility Model</h2>
                <p>Mycelium Cloud follows a cloud-industry-standard shared responsibility model.</p>

                <h3>6.1 Our Responsibilities</h3>
                <p>We are responsible for:</p>
                <ul>
                  <li>the Kubernetes Control Plane</li>
                  <li>orchestration logic</li>
                  <li>node provisioning interface</li>
                  <li>platform security of core systems</li>
                  <li>monitoring and metering</li>
                  <li>network infrastructure between platform components</li>
                </ul>

                <h3>6.2 Your Responsibilities</h3>
                <p>You are solely responsible for:</p>
                <ul>
                  <li>your application code, workloads, and containers</li>
                  <li>RBAC, network policies, and ingress rules</li>
                  <li>secrets and credentials</li>
                  <li>workload security and patching</li>
                  <li>backups and recovery of Customer Data</li>
                  <li>regulatory compliance</li>
                  <li>preventing misuse of your workloads</li>
                </ul>
                <p>We do not monitor workloads for security vulnerabilities, compliance, or misconfigurations.</p>
              </section>

              <section class="terms-section">
                <h2>7. Payments, Billing & Refunds</h2>

                <h3>7.1 Payment Processing</h3>
                <p>Payments are processed through third-party providers such as Stripe. By purchasing credits, you authorize us and our partners to charge your selected payment method.</p>

                <h3>7.2 Credits & TFT</h3>
                <p>Fiat payments are internally converted into TFT for usage metering. You acknowledge the following:</p>
                <ul>
                  <li>pricing may be based on TFT value</li>
                  <li>TFT value may fluctuate</li>
                  <li>you accept all associated exchange-rate risks</li>
                </ul>

                <h3>7.3 No Refunds</h3>
                <p>All payments, credits, and balances are non-refundable. This applies regardless of:</p>
                <ul>
                  <li>unused credits</li>
                  <li>account termination</li>
                  <li>downtime</li>
                  <li>user error</li>
                  <li>workload misconfiguration</li>
                </ul>

                <h3>7.4 Overages & Insufficient Balance</h3>
                <p>If your credit balance reaches zero:</p>
                <ul>
                  <li>workloads may pause or terminate,</li>
                  <li>nodes may be deprovisioned,</li>
                  <li>data or cluster state may be lost.</li>
                </ul>
                <p>We are not liable for any resulting loss.</p>
              </section>

              <section class="terms-section">
                <h2>8. Service Availability & SLA</h2>

                <h3>8.1 SLA Commitment</h3>
                <p>We target 99.9% uptime for the Kubernetes orchestration layer. This SLA does not apply to:</p>
                <ul>
                  <li>user workloads</li>
                  <li>third-party Node Providers</li>
                  <li>user misconfigurations</li>
                  <li>maintenance windows</li>
                  <li>network failures outside our control.</li>
                </ul>

                <h3>8.2 Remedies</h3>
                <p>The sole remedy for SLA breaches is service credits, calculated at our discretion. Service credits have no cash value and are not refundable.</p>

                <h3>8.3 No Guarantee for Decentralized Nodes</h3>
                <p>You acknowledge that decentralized Node Providers are:</p>
                <ul>
                  <li>independent third parties,</li>
                  <li>not controlled by TF HUB Limited,</li>
                  <li>variable in performance and uptime.</li>
                </ul>
                <p>We provide no warranty for node availability or reliability.</p>
              </section>

              <section class="terms-section">
                <h2>9. Third-Party Services</h2>
                <p>The Service may rely on third-party providers, including but not limited to:</p>
                <ul>
                  <li>Stripe (payments)</li>
                  <li>SendGrid (emails)</li>
                  <li>Container registries (Docker Hub, GHCR, etc.)</li>
                  <li>Node Providers (decentralized infrastructure operators)</li>
                </ul>
                <p>You agree to the separate terms of those services. We are not liable for downtime or errors caused by third parties.</p>
              </section>

              <section class="terms-section">
                <h2>10. Customer Data & Privacy</h2>

                <h3>10.1 Data Ownership</h3>
                <p>You retain full ownership of Customer Data. You grant us a limited license to store, process, and transmit data solely to operate the Service.</p>

                <h3>10.2 Data Locations</h3>
                <p>Customer Data may be processed:</p>
                <ul>
                  <li>globally</li>
                  <li>on decentralized nodes outside your jurisdiction</li>
                  <li>by third-party Node Providers</li>
                </ul>

                <h3>10.3 Backups</h3>
                <p>We do not guarantee backups or data durability. You are solely responsible for maintaining backups.</p>

                <h3>10.4 Data Protection</h3>
                <p>We follow reasonable administrative and technical measures, but cannot guarantee:</p>
                <ul>
                  <li>absolute security,</li>
                  <li>encryption on third-party nodes,</li>
                  <li>data integrity on decentralized networks.</li>
                </ul>
              </section>

              <section class="terms-section">
                <h2>11. Security</h2>
                <p>You must:</p>
                <ul>
                  <li>secure your workloads and images</li>
                  <li>implement encryption where required</li>
                  <li>use strong access controls</li>
                  <li>rotate credentials regularly</li>
                </ul>
                <p>We do not review your workloads for security risks.</p>
              </section>

              <section class="terms-section">
                <h2>12. Early Access Features</h2>
                <p>From time to time, we may provide access to Early Access Features, experimental or pre-release functionality.</p>
                <p>Early Access Features are provided "as-is", without warranty, and may be modified or removed at any time. They are excluded from all SLAs and support guarantees. Use them at your own risk.</p>
              </section>

              <section class="terms-section">
                <h2>13. Termination</h2>

                <h3>13.1 By You</h3>
                <p>You may stop using the Service at any time.</p>

                <h3>13.2 By Us</h3>
                <p>We may suspend or terminate your account if you:</p>
                <ul>
                  <li>violate these Terms or AUP</li>
                  <li>engage in harmful or abusive behavior</li>
                  <li>pose a security risk</li>
                  <li>fail to pay usage fees</li>
                </ul>

                <h3>13.3 Impact of Termination</h3>
                <p>Upon termination:</p>
                <ul>
                  <li>workloads may be immediately destroyed</li>
                  <li>Customer Data may be deleted within 30 days</li>
                </ul>
                <p>We are not liable for data loss.</p>
              </section>

              <section class="terms-section">
                <h2>14. Export Control & Sanctions</h2>
                <p>You may not use the Service if you are:</p>
                <ul>
                  <li>located in an OFAC-sanctioned country</li>
                  <li>on a sanctions or restricted parties list</li>
                  <li>using the Service for export-restricted technologies</li>
                </ul>
                <p>By using the Service, you confirm compliance with all applicable export control laws.</p>
              </section>

              <section class="terms-section">
                <h2>15. Intellectual Property</h2>
                <p>The platform, software, brand, and all related IP are owned by TF HUB Limited or its licensors. You may not reverse engineer, copy, or modify any part of the Service.</p>
              </section>

              <section class="terms-section">
                <h2>16. Indemnification</h2>
                <p>You agree to indemnify and hold harmless TF HUB Limited and its affiliates against claims arising from:</p>
                <ul>
                  <li>your misuse of the Service</li>
                  <li>violations of these Terms</li>
                  <li>your workloads or data</li>
                  <li>your breach of laws or regulations</li>
                </ul>
              </section>

              <section class="terms-section">
                <h2>17. Limitation of Liability</h2>
                <p>To the maximum extent permitted by law:</p>
                <ul>
                  <li>TF HUB Limited is not liable for indirect, incidental, or consequential damages</li>
                  <li>liability is limited to the amount you paid in the last 12 months, capped at $100 USD</li>
                  <li>we are not responsible for data loss, downtime, or workload failures</li>
                </ul>
                <p>This limitation survives termination.</p>
              </section>

              <section class="terms-section">
                <h2>18. Changes to Terms</h2>
                <p>We may update these Terms at any time. Your continued use of the Service constitutes acceptance of the revised Terms.</p>
              </section>

              <section class="terms-section">
                <h2>19. Governing Law</h2>
                <p>These Terms are governed by the laws of the British Virgin Islands (BVI). Any disputes shall be resolved exclusively in the courts of BVI.</p>
              </section>

              <section class="terms-section">
                <h2>20. Contact</h2>
                <p><strong>TF HUB Limited</strong></p>
                <p>Intershore Chambers, Road Town<br>Tortola, British Virgin Islands</p>
                <p>Email: <a href="mailto:info@threefold.io" class="contact-link">info@threefold.io</a></p>
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

.terms-card {
  background: rgba(10, 25, 47, 0.8);
  backdrop-filter: blur(8px);
  border: 1px solid rgba(96, 165, 250, 0.1);
}

.terms-content {
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

.terms-section {
  margin-top: 3rem;
  margin-bottom: 2rem;
}

.terms-section:first-of-type {
  margin-top: 2rem;
}

.terms-section h2 {
  font-size: 1.75rem;
  font-weight: 600;
  color: #60a5fa;
  margin-bottom: 1.5rem;
  margin-top: 2.5rem;
  scroll-margin-top: 100px;
}

.terms-section h2:first-child {
  margin-top: 0;
}

.terms-section h3 {
  font-size: 1.3rem;
  font-weight: 600;
  color: #38bdf8;
  margin-top: 2rem;
  margin-bottom: 1rem;
}

.terms-section p {
  margin-bottom: 1.25rem;
  color: #e0e7ef;
  opacity: 0.9;
  line-height: 1.8;
}

.terms-section ul {
  margin-left: 1.5rem;
  margin-bottom: 1.5rem;
  padding-left: 0;
}

.terms-section li {
  margin-bottom: 0.75rem;
  color: #e0e7ef;
  opacity: 0.9;
  line-height: 1.7;
  list-style-type: disc;
}

.terms-section strong {
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

  .terms-card {
    padding: 2rem !important;
  }

  .terms-section h2 {
    font-size: 1.5rem;
  }

  .terms-section h3 {
    font-size: 1.2rem;
  }
}

@media (max-width: 600px) {
  .hero-section {
    min-height: 25vh;
  }

  .terms-card {
    padding: 1.5rem !important;
  }

  .terms-content {
    font-size: 0.95rem;
  }

  .terms-section h2 {
    font-size: 1.35rem;
  }

  .terms-section h3 {
    font-size: 1.1rem;
  }
}
</style>

