---
title: Terms & Conditions
updatedAt: November 20, 2025
---

These Terms and Conditions ("Terms") govern your access to and use of the Mycelium Cloud platform and related services (the "Service"), operated by TF HUB Limited, a company organized under the laws of the British Virgin Islands ("we", "us", "our").

By creating an account, deploying workloads, or using the Service in any way, you agree to be bound by these Terms.

If you do not agree, you must discontinue use of the Service.

## 1. Definitions
- **"Cluster"** means a Kubernetes cluster provisioned or managed through the Service.
- **"Node"** means any virtual machine or server – centralized or decentralized – participating in a Cluster.
- **"Workload"** means any application, container, or process deployed on a Cluster.
- **"TFT"** (ThreeFold Token) means a digital utility token used internally for measuring and accounting resource consumption on the Service. TFT has no redemption, ownership, or financial rights.
- **"Credits"** means compute credits purchased in fiat and internally converted into TFT for usage accounting.
- **"Node Provider"** means an independent third party supplying decentralized compute resources.
- **"Control Plane"** means the Kubernetes orchestration layer managed by the Service.
- **"Customer Data"** means any data, configuration, image, or workload submitted to the Service.

## 2. About the Service
Mycelium Cloud provides tools for deploying, managing, and scaling Kubernetes clusters on centralized and decentralized infrastructure.

The Service includes:
- Kubernetes cluster provisioning
- Node and workload orchestration
- Monitoring and metrics
- Access control and API usage
- Billing and usage tracking

We may enhance, modify, or discontinue features at any time.

## 3. Eligibility

You must be legally authorized to enter binding contracts in your jurisdiction. If you use the Service on behalf of an entity, you represent that you have authority to bind that entity to these Terms.

## 4. Accounts

You must create an account to access the Service. You agree to:

- provide accurate information,
- maintain confidentiality of your credentials,
- be responsible for all actions taken through your account.

We may suspend or terminate accounts that violate these Terms or create security, legal, or operational risk.

## 5. Acceptable Use Policy (AUP)

You must not use the Service for any unlawful, harmful, or abusive activity.

The following activities are strictly prohibited:

<div class="grid">

- ### 5.1 Security & Network Abuse
    - DDoS attacks, port scanning, or malicious probing
    - Hosting malware, ransomware, spyware, or botnets
    - Using the platform as a VPN, proxy, anonymizer, or TOR exit node
    - Disrupting or degrading the infrastructure

- ### 5.2 Resource Abuse
    - Cryptocurrency mining of any kind
    - CPU/GPU-intensive workloads intended to exploit pricing
    - Unbounded fork bombs or uncontrolled resource loops
</div>

<div class="grid">

- ### 5.3 Legal & Content Restrictions
    - Hosting illegal content
    - Storage or processing of highly regulated data (PCI, PHI, HIPAA) unless you have ensured compliance
    - Copyright infringement

- ### 5.4 High-Risk Use
    The Service may not be used for life-critical systems, medical devices, autonomous vehicles, aviation, or nuclear operations.

    Violation of this AUP may result in immediate suspension without refund.
</div>

## 6. Shared Responsibility Model

Mycelium Cloud follows a cloud-industry-standard shared responsibility model.

<div class="grid">

- ### 6.1 Our Responsibilities
    We are responsible for:

    - the Kubernetes Control Plane
    - orchestration logic
    - node provisioning interface
    - platform security of core systems
    - monitoring and metering
    - network infrastructure between platform components

- ### 6.2 Your Responsibilities
    You are solely responsible for:

    - your application code, workloads, and containers
    - RBAC, network policies, and ingress rules
    - secrets and credentials
    - workload security and patching
    - backups and recovery of Customer Data
    - regulatory compliance
    - preventing misuse of your workloads
</div>

We do not monitor workloads for security vulnerabilities, compliance, or misconfigurations.

## 7. Payments, Billing & Refunds

<div class="grid">

- ### 7.1 Payment Processing

    Payments are processed through third-party providers such as Stripe. By purchasing credits, you authorize us and our partners to charge your selected payment method.

- ### 7.2 Credits & TFT

    Fiat payments are internally converted into TFT for usage metering. You acknowledge the following:

    - pricing may be based on TFT value
    - TFT value may fluctuate
    - you accept all associated exchange-rate risks
</div>

<div class="grid">

- ### 7.3 No Refunds
    All payments, credits, and balances are non-refundable. This applies regardless of:

    - unused credits
    - account termination
    - downtime
    - user error
    - workload misconfiguration

- ### 7.4 Overages & Insufficient Balance
    If your credit balance reaches zero:

    - workloads may pause or terminate,
    - nodes may be deprovisioned,
    - data or cluster state may be lost.

    We are not liable for any resulting loss.
</div>

## 8. Service Availability & SLA

<div class="grid">

- ### 8.1 SLA Commitment
    We target 99.9% uptime for the Kubernetes orchestration layer. This SLA does not apply to:

    - user workloads
    - third-party Node Providers
    - user misconfigurations
    - maintenance windows
    - network failures outside our control.

- ### 8.2 Remedies
    The sole remedy for SLA breaches is service credits, calculated at our discretion. Service credits have no cash value and are not refundable.
</div>

- ### 8.3 No Guarantee for Decentralized Nodes
     You acknowledge that decentralized Node Providers are:

    - independent third parties,
    - not controlled by TF HUB Limited,
    - variable in performance and uptime.

    We provide no warranty for node availability or reliability.

## 9. Third-Party Services

The Service may rely on third-party providers, including but not limited to:

- Stripe (payments)
- SendGrid (emails)
- Container registries (Docker Hub, GHCR, etc.)
- Node Providers (decentralized infrastructure operators)

You agree to the separate terms of those services. We are not liable for downtime or errors caused by third parties.

## 10. Customer Data & Privacy

<div class="grid">

- ### 10.1 Data Ownership
    You retain full ownership of Customer Data. You grant us a limited license to store, process, and transmit data solely to operate the Service.

- ### 10.2 Data Locations
    Customer Data may be processed:

    - globally
    - on decentralized nodes outside your jurisdiction
    - by third-party Node Providers
</div>

<div class="grid">

- ### 10.3 Backups
    We do not guarantee backups or data durability. You are solely responsible for maintaining backups.

- ### 10.4 Data Protection
    We follow reasonable administrative and technical measures, but cannot guarantee:

    - absolute security,
    - encryption on third-party nodes,
    - data integrity on decentralized networks.
</div>

## 11. Security
You must:

- secure your workloads and images
- implement encryption where required
- use strong access controls
- rotate credentials regularly
We do not review your workloads for security risks.

## 12. Early Access Features
From time to time, we may provide access to Early Access Features, experimental or pre-release functionality.

Early Access Features are provided "as-is", without warranty, and may be modified or removed at any time. They are excluded from all SLAs and support guarantees. Use them at your own risk.

## 13. Termination

- ### 13.1 By You
    You may stop using the Service at any time.

<div class="grid">

- ### 13.2 By Us
    We may suspend or terminate your account if you:

    - violate these Terms or AUP
    - engage in harmful or abusive behavior
    - pose a security risk
    - fail to pay usage fees

- ### 13.3 Impact of Termination
    Upon termination:

    - workloads may be immediately destroyed
    - Customer Data may be deleted within 30 days

    We are not liable for data loss.
</div>

## 14. Export Control & Sanctions
You may not use the Service if you are:

- located in an OFAC-sanctioned country
- on a sanctions or restricted parties list
- using the Service for export-restricted technologies

By using the Service, you confirm compliance with all applicable export control laws.

## 15. Intellectual Property
The platform, software, brand, and all related IP are owned by TF HUB Limited or its licensors. You may not reverse engineer, copy, or modify any part of the Service.

## 16. Indemnification
You agree to indemnify and hold harmless TF HUB Limited and its affiliates against claims arising from:

- your misuse of the Service
- violations of these Terms
- your workloads or data
- your breach of laws or regulations

## 17. Limitation of Liability
To the maximum extent permitted by law:

- TF HUB Limited is not liable for indirect, incidental, or consequential damages
- liability is limited to the amount you paid in the last 12 months, capped at $100 USD
- we are not responsible for data loss, downtime, or workload failures

This limitation survives termination.

## 18. Changes to Terms
We may update these Terms at any time. Your continued use of the Service constitutes acceptance of the revised Terms.

## 19. Governing Law
These Terms are governed by the laws of the British Virgin Islands (BVI). Any disputes shall be resolved exclusively in the courts of BVI.

## 20. Contact

`TF HUB Limited`

Intershore Chambers, Road Town<br/>
Tortola, British Virgin Islands

Email: <a href="mailto:info@threefold.io" class="text-link">info@threefold.io</a>
