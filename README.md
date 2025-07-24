<img src="2quadlet-logo.png" alt="2quadlet" width="300"/>

# Convert Podman Commands and Compose Files to Podman Quadlets

**2quadlet** is a simple and powerful command-line tool that helps you generate [Podman Quadlet](https://docs.podman.io/en/latest/markdown/podman-systemd.unit.html) files from:

- Single `podman run` or `podman create` commands
- `docker-compose.yaml` files

It bridges the gap between development workflows and production-ready systemd service files for containers using **Podman**.

---

## Features

- Convert basic `podman run` or `create` commands into valid quadlet `.container` files
- Translate `docker-compose.yaml` into multiple quadlet units
- Output clean, systemd-compliant unit files
- Ideal for creating system services or simplifying container deployment
- Lightweight and fast – perfect for automation or CI/CD

---

## Installation and Usage

Clone the repo and run the tool directly:

```bash
git clone https://github.com/youruser/2quadlet.git
cd 2quadlet

go run compose-to-qadlet.go docker-compose.yml .
go run podman-to-quadlet.go -cmd 'podman run -d --name webapp -p 8080:80 -v /data:/app/data:Z docker.io/nginx:latest'  -output .
```

Author: Thomas Mundt - tmundt@softxpert.de
