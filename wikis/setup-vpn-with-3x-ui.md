---
title: "setup vpn with 3x-ui"
desc: ""
createdAt: "2025-12-21T14:33:44Z"
---

reference: [搭建節點](https://www.cnblogs.com/JourneyOfFlower/p/18314466), [guide](https://semenov.work/posts/3x-ui-vless-reality-vpn)

# install 3x-ui

```
bash <(curl -Ls https://raw.githubusercontent.com/MHSanaei/3x-ui/refs/tags/v2.5.8/install.sh)
```

# Setup firewall rules on VPS

ip address range: `0.0.0.0/0`
allow allow protocol and port

# cloudfare

- ensure proxy is not enable,

# Create v2ray

choose vless + reality

# setup SSL for 3x-ui

Add a new record in cloudflare
name: vps
ip: ip of the vps

get the ssl from 3x-ui
domain name: vps.`your domain name`
