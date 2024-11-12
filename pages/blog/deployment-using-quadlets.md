---
title: How I self-host using Ansible and Quadlets
description: |
    Article relating my experience using Ansible and Quadlets for self-hosting services
date: "2025 June 02"
format: blog_entry
require_prism: true
---

# Introduction

I have a NAS and a VPS running a small set of services that I use often; think services like [Jellyfin](https://jellyfin.org/), [Home Assistant](https://home-assistant.io), [Victoria Metrics](https://victoriametrics.com), [Docker Registry](https://github.com/distribution/distribution), etc.

Recently I wanted to change how I deploy these services to make it easier on myself; this post will go through the changes I made.

But first, a small summary of my setup.

I'm using [AlmaLinux](https://almalinux.org) or [Fedora Server](https://fedoraproject.org/en/server/). To deploy my services currently I either install a RPM package if it exists or upload a binary on the host.

While this setup works fine, there are some drawbacks:
1. keeping a service up to date is annoying if you're not installing from a RPM repository
2. if the service is not a simple binary, for example if it needs Node or Python, it can be a pain in the ass to get working without messing the system which is why I usually avoid complex services

I can already hear you saying "What about Docker ?"; hold that thought.

All of this is managed using Ansible with roles and playbooks I crafted for the last 10 years.

# Improving the setup

Yes, containers are the answer to my problems here:
1. an update is one `docker pull` or `podman pull` away
2. everything needed is self-contained and there's no messing with the host system. I can use services written in Python or Node without worries.

Yes, I could use Docker and it would work fine, but I don't want to.

Partly for technical reasons: I much prefer how Podman containers are integrated with systemd.
But also, this is my personal setup where I get to do what I want, so I'd like to try something new.

Enter [Podman](https://podman.io) and their [Quadlet](https://docs.podman.io/en/stable/markdown/podman-systemd.unit.5.html) concept.

# What is Quadlet ?

Quadlet is a tool included in recent Podman versions that turns _unit_ files into systemd services. It is the successor to `podman-systemd-generate`.

There are different types of unit: container, volume, network, build, pod and kube. We'll only look at the first three since that's the only ones I've used so far.

The unit files use the same format as a systemd unit files; in fact you can put any standard systemd stuff and it well be passed untouched.
There is however a custom section for each unit type which we'll look into below.

The files must be stored in a _search path_ like `/etc/containers/systemd` (look at the [documentation](https://docs.podman.io/en/stable/markdown/podman-systemd.unit.5.html#podman-rootful-unit-search-path) for all the search paths).

The files are then read by the `quadlet` binary which is a [systemd generator](https://www.freedesktop.org/software/systemd/man/latest/systemd.generator.html).
When the systemd daemon is booted or reloaded `quadlet` will generate systemd services for the files it found in the _search paths_.

![quadlet generator](./deployment-using-quadlets/quadlet_generator.avif)

Once the service has been generated it is a completely standard systemd unit so you have to all the common tools (systemctl to control it, journalctl to read its logs, and more).

Now let's see what those unit files look like.

# What are `.container` files ?

A `.container` file is a unit describing the container to be run as a service.

This unit requires a custom `[Container]` section to define the container to run.

The only required option in this section is the `Image` which is simply the container image you want to run.
Of course there a ton of options to configure how to run the container which you can see in the [documentation](https://docs.podman.io/en/stable/markdown/podman-systemd.unit.5.html#container-units-container).

If you're familiar with `podman run` or `docker run` or even a Docker Compose manifest, it shouldn't be too hard to map the different flags to options.

For example, let's say I have this `docker run` command to run [Miniflux](https://miniflux.app/):
```bash
docker run -d \
  -p 80:8080 \
  --name miniflux \
  -e "DATABASE_URL=postgres://miniflux:*password*@*dbhost*/miniflux?sslmode=disable" \
  -e "RUN_MIGRATIONS=1" \
  -e "CREATE_ADMIN=1" \
  -e "ADMIN_USERNAME=*username*" \
  -e "ADMIN_PASSWORD=*password*" \
  docker.io/miniflux/miniflux:latest
```

This is a equivalent `miniflux.container` unit file:
```systemd
[Container]
Image=docker.io/miniflux/miniflux:latest
ContainerName=miniflux
PublishPort=80:8080
Environment="DATABASE_URL=user=miniflux password=foobar host=postgresql dbname=miniflux sslmode=disable"
Environment="RUN_MIGRATIONS=1"
Environment="CREATE_ADMIN=1"
Environment="ADMIN_USERNAME=foobar"
Environment="ADMIN_PASSWORD=foobar"
```

Put this in `/etc/containers/systemd`, run `systemctl daemon-reload` then `systemctl start miniflux` and you're good to go.

There are quite a lot of options available but on the off chance you don't find what you want, you can easily provide Podman arguments, for example:
```
PodmanArgs=--memory-swap 1G
```
