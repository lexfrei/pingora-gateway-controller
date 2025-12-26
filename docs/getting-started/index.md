# Getting Started

This section covers everything you need to get the Pingora Gateway Controller
running in your Kubernetes cluster.

## Overview

The controller implements the Gateway API for a Pingora-based reverse proxy.
Before installing, you need:

1. A Kubernetes cluster with Gateway API CRDs installed
2. Helm 3.x for installation (recommended)
3. Network connectivity between controller and backend services

## Sections

<div class="grid cards" markdown>

-   :material-check-circle:{ .lg .middle } **Prerequisites**

    ---

    Required components, Kubernetes version, and Gateway API CRDs setup.

    [:octicons-arrow-right-24: Prerequisites](prerequisites.md)

-   :material-download:{ .lg .middle } **Installation**

    ---

    Install the controller and proxy using Helm chart with all configuration options.

    [:octicons-arrow-right-24: Installation](installation.md)

-   :material-rocket-launch:{ .lg .middle } **Quick Start**

    ---

    Create your first HTTPRoute and expose a service through Pingora proxy.

    [:octicons-arrow-right-24: Quick Start](quickstart.md)

</div>

## Next Steps

After completing the getting started guide:

- Learn about [Configuration](../configuration/index.md) options
- Explore [Gateway API](../gateway-api/index.md) features and examples
- Set up [Monitoring](../guides/monitoring.md) for production deployments
