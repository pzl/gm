Manager
=======

For lack of a better name, and until I switch to better monitoring (ELK stack & prometheus+grafana): this is my status dashboard page for my home server.

It primarily monitors systemd service units, some of which are `rkt` containers. It monitors the status of the services, ports in use, server disk information (blocks size, usage, inodes), and where applicable, rkt container info: mounts, ports, status, image info.

**Manager** is accessed as a website. The backend is written in `Go`. This fetches the systemd info via `dbus`, and rkt info via the [rkt api-service](https://coreos.com/rkt/docs/latest/subcommands/api-service.html) if you have it running. The front-end is a static site built with [Nuxt](https://nuxtjs.org/) and [Vue](https://vuejs.org/). The site is fully responsive, and fetches new info from the backend on every tab change. No refresh necessary.

![dashboard screenshot](https://raw.githubusercontent.com/pzl/manager/assets/main.png)

![status screenshot](https://raw.githubusercontent.com/pzl/manager/assets/svc.png)

![data screenshot](https://raw.githubusercontent.com/pzl/manager/assets/data.png)

Building & Installation
-------------

Go dependencies are managed using [dep](https://golang.github.io/dep/). And the frontend with npm. You must have the following installed for a build environment:

- nodejs
- npm
- go
- dep

The makefile command `install-deps` has been provided to perform the dep install and npm install for you. Just run `make install-deps` to run installation for the various third-party dependencies.

The server can be build with `go build` to create a `manager` binary for the server. The frontend can be build with running `npm run generate` to create the frontend as a directory of static assets. You may run `make` to perform both of these steps automatically.

### Distribute or Install

With the server and assets built, no extra runtime is required on the machine that will host the `manager` program, if different than the build machine (Go, Node: _not_ required).

Simply copy the `manager` binary wherever you want, and move the `frontend/dist` folder and contents to be next to that binary (and rename `dist` to `frontend`).


License
--------

This project is licensed under the MIT License. See `Copying` for the full license. Copyright 2018 Dan Panzarella.