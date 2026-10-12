# w42-eu-web

The [w42.eu](https://w42.eu) landing page: the projects under w42.eu
([s.i41.eu](https://s.i41.eu), [mcbot.w42.eu](https://mcbot.w42.eu), [mc.w42.eu](https://mc.w42.eu)),
with links to [mj41.cz](https://mj41.cz) and [GitHub](https://github.com/mj41) for more projects.

It also serves [s.i41.eu](https://s.i41.eu), the Stackchan project page (the robots' manager at
m.i41.eu, their apps under sa.w42.eu, with the robot (`s/robot3d-*.png`) and screenshots of the apps
(`s/*.webp`)) and a page per app: [s.i41.eu/pet](https://s.i41.eu/pet),
[/focus](https://s.i41.eu/focus), [/raw](https://s.i41.eu/raw), [/dev](https://s.i41.eu/dev) (`app.html` around
`apps/<app>.html`; the parts each app uses link to that part on [s.i41.eu/s3d](https://s.i41.eu/s3d/)),
and [mcbot.w42.eu](https://mcbot.w42.eu), the Minecraft robots' page, with a screenshot of their
dashboard (`mcbot/*.png|webp`: small in the page, the full size on click, by CSS alone).

The robot pictures `s/robot3d-*.png` are 3D renders of the robot (M5Stack's StackChan structure
files, MIT, and photos of it) with real screens from the robot, written by `robot3d-distribute` in
s-w42-eu-assets: change them there, not here.

One Go binary with the pages embedded (`index.html`, `s.html`, `mcbot.html`, the app pages and their images), no
dependencies. The request's host picks the page: `s.i41.eu` gets `s.html`, `mcbot.w42.eu`
`mcbot.html`, any other host `index.html`; `s.i41.eu/<app>` an app's page (a new app: add
`apps/<app>.html` and its entry in `apps` in `main.go`); `/mcbot/<name>.png|webp` and `/s/<name>.png|webp` serve the
images on any host.

```bash
go run .                 # http://localhost:8080
curl -H 'Host: s.i41.eu' http://localhost:8080/
```

A `v*` tag builds `ghcr.io/mj41/w42-eu-web:<tag>`.

## License

Apache License 2.0, see [LICENSE](LICENSE).

Stack-chan (スタックチャン) is a registered trademark of Shinya Ishikawa; this project is
independent and only made to work with [Stack-chan](https://github.com/stack-chan/stack-chan) robots.
