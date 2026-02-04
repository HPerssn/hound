# Hound

Self-hosted full-stack application running on a k3s cluster.  
Provides authenticated access and internal routing for services, with a lightweight vanilla JavaScript frontend.

---

## Overview

Hound is designed as a foundational service for self-hosted infrastructure.  
It combines a backend API with a minimal frontend, deployed Kubernetes-native and secured via cluster-aware authentication.

---

## Features

- Full-stack application (API + frontend)
- Vanilla JavaScript frontend (no framework)
- Authenticated access to internal services
- Kubernetes-native deployment (k3s)
- Environment-driven configuration

---

## Tech Stack

- **Backend:** Node.js / TypeScript
- **Frontend:** Vanilla JavaScript, HTML, CSS
- **Runtime:** Docker
- **Orchestration:** k3s (Kubernetes)
- **Auth:** Token-based
- **CI/CD:** GitHub Actions

---

## Deployment

Deployed to a **k3s cluster** using Kubernetes manifests.

- Containerized services
- Secrets and config via Kubernetes resources
- Ingress handles routing and auth enforcement

---

## Local Development

```bash
npm install
npm run dev
