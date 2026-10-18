# w42-eu-web

The [w42.eu](https://w42.eu) landing page: a list of mj41's projects
([i41.eu](https://i41.eu), [mcbot.w42.eu](https://mcbot.w42.eu), [mc.w42.eu](https://mc.w42.eu)),
with links to [mj41.cz](https://mj41.cz) and [GitHub](https://github.com/mj41) for more.

It also serves [mcbot.w42.eu](https://mcbot.w42.eu), the Minecraft robots' page, with a
views from the robots' action camera (`mcbot/*.png|webp`: small in the page, the full size on click, by
CSS alone). The Stackchan pages (s.i41.eu) are
[i41eu-web](https://github.com/i41-eu/i41eu-web)'s.

One Go binary with the pages embedded (`index.html`, `mcbot.html` and its images), no
dependencies. The request's host picks the page: `mcbot.w42.eu` gets `mcbot.html`, any other
host `index.html`; `/mcbot/<name>.png|webp` serves the images on any host.

```bash
go run .                 # http://localhost:8080
curl -H 'Host: mcbot.w42.eu' http://localhost:8080/
```

A `v*` tag builds `ghcr.io/mj41/w42-eu-web:<tag>`.

## License

Apache License 2.0, see [LICENSE](LICENSE).

Stack-chan (スタックチャン) was created by Shinya Ishikawa and the Stack-chan community ([github.com/stack-chan/stack-chan](https://github.com/stack-chan/stack-chan)); this project is independent of it and of M5Stack.
