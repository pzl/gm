Manager
=======

For lack of a better name, and until I switch to better monitoring (ELK stack & prometheus+grafana): this is my status dashboard page for my home server.

It primarily monitors systemd service units, some of which are `podman` containers. It monitors the status of the services, ports in use, server disk information (blocks size, usage, inodes), and where applicable, podman container info: mounts, ports, status, image info.

**Manager** is accessed as a website. The backend is written in `go`. This fetches the systemd info via `dbus`, and podman info via the podman REST API if you have it running. The front-end is a static site built with [Nuxt](https://nuxtjs.org/) and [Vue](https://vuejs.org/). The site is fully responsive, and fetches new info from the backend on every tab change. No refresh necessary.

![dashboard screenshot](https://raw.githubusercontent.com/pzl/manager/assets/main.png)

![status screenshot](https://raw.githubusercontent.com/pzl/manager/assets/svc.png)

![data screenshot](https://raw.githubusercontent.com/pzl/manager/assets/data.png)

Building & Installation
-------------

Just run `make` to install any needed dependencies with `npm`, and with `go` automatically.

### Distribute or Install

With the server and assets built, no extra runtime is required on the machine that will host the `manager` program, if different than the build machine (Go, Node: _not_ required).

Simply copy the `manager` binary wherever you want.


License
--------

This project is licensed under the MIT License. See `LICENSE` for the full license. Copyright 2018 Dan Panzarella.