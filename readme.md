# Rewrite Header

Rewrite header is a middleware plugin for [Traefik](https://traefik.io) which replace a header in the response

## Configuration

### Static

```yaml
pilot:
  token: "xxxx"

experimental:
  plugins:
    rewriteHeaders:
      modulename: "github.com/XciD/traefik-plugin-rewrite-headers"
      version: "v0.0.3"
```

### Dynamic

To configure the Rewrite Head plugin you should create a [middleware](https://docs.traefik.io/middlewares/overview/) in your dynamic configuration as explained [here](https://docs.traefik.io/middlewares/overview/). 
The following example creates and uses the rewriteHeaders middleware plugin to modify the Location header

```yaml
http:
  routes:
    my-router:
      rule: "Host(`localhost`)"
      service: "my-service"
      middlewares : 
        - "rewriteHeaders"
  services:
    my-service:
      loadBalancer:
        servers:
          - url: "http://127.0.0.1"
  middlewares:
    rewriteHeaders:
      plugin:
        rewriteHeaders:
          rewrites:
            - header: "Location"
              regex: "^http://(.+)$"
              replacement: "https://$1"
```
