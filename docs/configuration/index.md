# Configuration

This section covers all configuration options for the Pingora Gateway Controller.

## Overview

Configuration is managed at three levels:

1. **Controller flags** - Runtime behavior (logging, metrics, leader election)
2. **Helm values** - Deployment configuration (replicas, resources, images)
3. **PingoraConfig CRD** - Pingora proxy connection settings

## Sections

<div class="grid cards" markdown>

-   :material-console:{ .lg .middle } **Controller Options**

    ---

    CLI flags and environment variables for controller configuration.

    [:octicons-arrow-right-24: Controller Options](controller.md)

-   :material-file-cog:{ .lg .middle } **Helm Values**

    ---

    Complete reference for Helm chart values and customization.

    [:octicons-arrow-right-24: Helm Values](helm-values.md)

-   :material-connection:{ .lg .middle } **PingoraConfig CRD**

    ---

    Custom resource for configuring Pingora proxy connection.

    [:octicons-arrow-right-24: PingoraConfig](gatewayclassconfig.md)

</div>

## Configuration Hierarchy

```mermaid
graph TD
    subgraph Kubernetes Resources
        GC[GatewayClass]
        PC[PingoraConfig]
    end

    subgraph Helm Chart
        HV[values.yaml]
    end

    subgraph Controller
        CF[CLI Flags]
        EV[Environment Variables]
    end

    HV -->|generates| GC
    HV -->|generates| PC
    GC -->|parametersRef| PC
    CF -->|overrides| EV
```

## Quick Reference

| Configuration | Scope | Reload |
|--------------|-------|--------|
| CLI flags | Controller | Restart required |
| Environment variables | Controller | Restart required |
| Helm values | Deployment | `helm upgrade` |
| PingoraConfig | Proxy connection | Dynamic (watch) |
