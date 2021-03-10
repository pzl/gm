<div align="center">

![icon](frontend/static/icon.png)

</div>

<h3 align="center">gm</h3>

<div align="center">

  [![Status](https://img.shields.io/badge/status-active-success.svg)]() 
  [![GitHub Issues](https://img.shields.io/github/issues/pzl/gm.svg)](https://github.com/pzl/gm/issues)
  [![GitHub Pull Requests](https://img.shields.io/github/issues-pr/pzl/gm.svg)](https://github.com/pzl/gm/pulls)
  [![GoDoc](https://godoc.org/github.com/pzl/gm?status.svg)](https://godoc.org/github.com/pzl/gm)
  [![License](https://img.shields.io/badge/license-MIT-blue.svg)](/LICENSE)

</div>

---

<p align="center"> Local server management, with podman actions
    <br> 
</p>


## About

`gm` is a local server monitoring minisite. It has no authentication, authorization, or access control. It is extremely simplistic. More polished alternatives like [cockpit](https://cockpit-project.org/) may be a better option for you. This is simply tailored to my needs over LAN.

It primarily monitors systemd service units, some of which are `podman` containers. It monitors the status of the services, ports in use, server disk information (blocks size, usage, inodes), and where applicable, podman container info: mounts, ports, status, image info.

**gm** is accessed as a website. The backend is written in `go`. This fetches the systemd info via `dbus`, and podman info via the podman REST API if you have it running. The front-end is a static site built with [Nuxt](https://nuxtjs.org/) and [Vue](https://vuejs.org/).


Building & Installation
-------------

Just run `make` to install any needed dependencies with `npm`, and with `go` automatically.


Podman API
-----------

To get podman information, you need the podman system service running. You can run the process manually via:

```sh
podman system service -t 0
```

or set it up via systemd service (`podman.service`) or auto-activation with socket:

```sh
sudo systemctl start podman.socket
sudo systemctl enable podman.socket
```


## License

MIT License (c) 2018 Dan Panzarella
