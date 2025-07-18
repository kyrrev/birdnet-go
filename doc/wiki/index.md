# BirdNET-Go Documentation

Welcome to the BirdNET-Go documentation. This index will help you navigate through all available resources for installing, configuring, and using BirdNET-Go.

## Getting Started

* [BirdNET‐Go Guide](https://github.com/tphakala/birdnet-go/wiki/BirdNET%E2%80%90Go-Guide) - Overview and basic concepts
* [Installation Guide](installation.md) - Comprehensive installation instructions for all methods
* [Recommended Hardware](hardware.md) - Hardware recommendations for optimal performance

## Installation Methods

* [Docker Installation (Linux)](installation.md#recommended-method-installsh-linux) - Using the automated `install.sh` script
* [Docker Compose Installation](docker_compose_guide.md) - Setting up BirdNET-Go with Docker Compose
* [Manual Docker Installation](installation.md#manual-docker-installation-advanced-linux-only) - Advanced Docker setup
* [Manual Binary Installation](installation.md#manual-binary-installation-all-platforms) - Windows, macOS, and Linux binary installation

**Note**: Docker images are available from both [GitHub Container Registry](https://github.com/tphakala/birdnet-go/pkgs/container/birdnet-go) and [Docker Hub](https://hub.docker.com/r/tphakala/birdnet-go).

## Advanced Features

* [BirdNET Detection Pipeline](BirdNET‐Go-Guide#birdnet-detection-pipeline) - Understanding how settings affect detections
* [BirdNET Range Filter](BirdNET‐Go-Guide#birdnet-range-filter) - Location and time-based species filtering
* [Web Dashboard](BirdNET‐Go-Guide#web-dashboard) - Using the visualization dashboard
* [Remote Internet Access](cloudflare_tunnel_guide.md) - Exposing BirdNET-Go to the internet securely
* [Weather Integration](BirdNET‐Go-Guide#weather-integration) - Connecting to weather data providers
* [Audio Processing](BirdNET‐Go-Guide#audio-processing) - Advanced audio processing capabilities
* [Deep Detection](BirdNET‐Go-Guide#deep-detection) - Improving detection reliability
* [Live Audio Streaming](BirdNET‐Go-Guide#live-audio-streaming) - Streaming audio from the web interface
* [Species-Specific Settings](BirdNET‐Go-Guide#species-specific-settings) - Customized detection rules for specific species

## Integration & Security

* [MQTT Integration](BirdNET‐Go-Guide#integration-options) - Connecting to IoT systems
* [BirdWeather API](BirdNET‐Go-Guide#integration-options) - Sharing data with BirdWeather.com
* [Authentication](cloudflare_tunnel_guide.md#enabling-authentication) - Securing your BirdNET-Go instance
* [Cloudflare Tunnel Setup](cloudflare_tunnel_guide.md) - Detailed guide for secure internet access

## Privacy & Telemetry

* [Error Tracking & Telemetry](telemetry.md) - Optional privacy-first error tracking system
* [Privacy & Data Collection](telemetry-privacy.md) - Detailed privacy information and data protection
* [Telemetry Setup Guide](telemetry-setup.md) - Step-by-step configuration instructions
* [Telemetry Troubleshooting](telemetry-troubleshooting.md) - Resolving telemetry-related issues

## Troubleshooting & Support

* [RTSP Troubleshooting](rtsp-troubleshooting.md) - Comprehensive guide for RTSP camera issues and configuration
* [Docker Troubleshooting](BirdNET‐Go-Guide#docker-installation-troubleshooting) - Resolving common Docker issues
* [Support Script](BirdNET‐Go-Guide#support-script) - Generating diagnostic information
* [Reporting Issues](BirdNET‐Go-Guide#reporting-issues) - How to effectively report bugs

## Reference

* [Command Line Interface](BirdNET‐Go-Guide#command-line-interface) - Available commands and options
* [Detection Pipeline Flow](BirdNET‐Go-Guide#birdnet-detection-pipeline) - How settings interact and affect detection results
* [Range Filter Commands](BirdNET‐Go-Guide#inspection-and-debugging) - CLI commands for inspecting range filter results
* [Supported Languages](BirdNET‐Go-Guide#supported-languages-for-species-labels) - Language options for species labels
* [Log Rotation](BirdNET‐Go-Guide#log-rotation) - Managing log files 