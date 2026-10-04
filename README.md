# w42-eu-web

The [w42.eu](https://w42.eu) landing page: the projects under w42.eu
([mc.w42.eu](https://mc.w42.eu), [s.w42.eu](https://s.w42.eu),
[home-w42-eu](https://github.com/mj41/home-w42-eu)), with links to
[mj41.cz](https://mj41.cz) and [GitHub](https://github.com/mj41).

It also serves [home.w42.eu](https://home.w42.eu), a page pointing to the
[home-w42-eu](https://github.com/mj41/home-w42-eu) repositories, where the up to date
information lives, and [s.w42.eu](https://s.w42.eu), the Stackchan project page (the robots'
manager at sm.w42.eu and their apps under sa.w42.eu).

One Go binary with the pages embedded (`index.html`, `home.html`, `s.html`), no dependencies. The
request's host picks the page: `home.w42.eu` gets `home.html`, `s.w42.eu` `s.html`, any other host `index.html`.

```bash
go run .                 # http://localhost:8080
curl -H 'Host: home.w42.eu' http://localhost:8080/
```

A `v*` tag builds `ghcr.io/mj41/w42-eu-web:<tag>`.

## License

Apache License 2.0, see [LICENSE](LICENSE).
