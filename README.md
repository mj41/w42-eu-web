# w42-eu-web

The [w42.eu](https://w42.eu) landing page: the projects under w42.eu
([s.w42.eu](https://s.w42.eu), [mcbot.w42.eu](https://mcbot.w42.eu), [mc.w42.eu](https://mc.w42.eu)),
with links to [mj41.cz](https://mj41.cz) and [GitHub](https://github.com/mj41) for more projects.

It also serves [s.w42.eu](https://s.w42.eu), the Stackchan project page (the robots' manager at
sm.w42.eu, their apps under sa.w42.eu, with the robot and screenshots of the apps: `s/*.webp`),
and [mcbot.w42.eu](https://mcbot.w42.eu), the Minecraft robots' page, with a screenshot of their
dashboard (`mcbot/*.png`: small in the page, the full size on click, by CSS alone).

The robot pictures in `s/` are 3D renders of the robot (M5Stack's StackChan structure files, MIT,
and photos of it) with real screens from the robot.

One Go binary with the pages embedded (`index.html`, `s.html`, `mcbot.html` and their images), no
dependencies. The request's host picks the page: `s.w42.eu` gets `s.html`, `mcbot.w42.eu`
`mcbot.html`, any other host `index.html`; `/mcbot/<name>.png` and `/s/<name>.webp` serve the
images on any host.

```bash
go run .                 # http://localhost:8080
curl -H 'Host: s.w42.eu' http://localhost:8080/
```

A `v*` tag builds `ghcr.io/mj41/w42-eu-web:<tag>`.

## License

Apache License 2.0, see [LICENSE](LICENSE).
