---
title: "How to setup synapse grpc server on vps"
desc: ""
createdAt: "2025-11-20T06:52:01Z"
---

Assume you use Cloudflare for proxying your domain.

For this guide, the domain is `grpc.skpstack.uk`

# Enable gRPC support in Cloudflare

1. Log in to Cloudflare.
2. Go to Network.
3. Toggle gRPC to On.
4. Go to SSL/TLS. Ensure it is set to Full or Full (Strict).

# Config Nginx

```bash
sudo -e /etc/nginx/sites-available/grpc.skpstack.uk

```

or

```bash
sudo -e /etc/nginx/sites-available/<your-domain>
```

Add the following configuration:

```nginx
server {
    server_name grpc.skpstack.uk;
    # server_name <your-domain>;

    # Log files for debugging
    access_log /var/log/nginx/grpc_access.log;
    error_log /var/log/nginx/grpc_error.log;

    # 1. Route Synapse gRPC traffic
    # The location matches the "package.Service" name from your .proto file
    location /protos.Generate/ {
        grpc_pass grpc://localhost:50051;

        # Standard gRPC headers
        grpc_set_header Host $host;
        grpc_set_header X-Real-IP $remote_addr;
        grpc_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    # 2. (Optional) Return 404 for anything that isn't your gRPC service
    location / {
        return 404;
    }

    # SSL configuration will be added by Certbot automatically below
    listen 80 http2;
}
```

Enable the site and test Nginx configuration:

```bash
sudo ln -s /etc/nginx/sites-available/grpc.skpstack.uk /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl daemon-reload
sudo systemctl restart nginx
```

Get the SSL Certificate:

```bash
sudo certbot --nginx -d grpc.skpstack.uk
```
