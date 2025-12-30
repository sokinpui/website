---
title: "How to setup synapse server on vps"
desc: ""
createdAt: "2025-11-20T06:52:01Z"
---

Assume you use Cloudflare for proxying your domain.

For this guide, the domain is `synapse.skpstack.uk`

# Enable gRPC support in Cloudflare

1. Log in to Cloudflare.
2. Go to Network.
3. Toggle gRPC to On.
4. Go to SSL/TLS. Ensure it is set to Full or Full (Strict).

# Config Nginx

```bash
sudo -e /etc/nginx/sites-available/synapse.skpstack.uk

```

or

```bash
sudo -e /etc/nginx/sites-available/<your-domain>
```

Add the following configuration:

```nginx
server {
    listen 80 http2; # 'http2' is required for the gRPC location
    server_name synapse.skpstack.uk;

    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log;

    # --- 1. gRPC Service (Port 50051) ---
    location /synapse.Generate/ {
        grpc_pass grpc://localhost:50051;

        grpc_set_header Host $host;
        grpc_set_header X-Real-IP $remote_addr;
        grpc_set_header X-Forwarded-For $proxy_add_x_forwarded_for;

        # Optional: Increase timeouts if gRPC calls are long
        # grpc_read_timeout 600s;
    }

    # --- 2. HTTP & SSE Service (Port 8080) ---
    location / {
        proxy_pass http://localhost:8080;

        # --- SSE & Streaming Requirements ---

        # 1. Use HTTP 1.1 (Required for keep-alive connections)
        proxy_http_version 1.1;

        # 2. Clear the Connection header to ensure keep-alive works
        proxy_set_header Connection "";

        # 3. Disable Buffering (CRITICAL for SSE)
        # If this is 'on', Nginx waits to fill a buffer before sending data to client,
        # breaking the real-time stream.
        proxy_buffering off;

        # 4. Disable Caching
        proxy_cache off;

        # 5. Increase Read Timeout
        # SSE connections stay open for a long time. Default is 60s.
        # Set this higher than your application's keep-alive/heartbeat interval.
        proxy_read_timeout 24h;

        # --- Standard Proxy Headers ---
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Enable the site and test Nginx configuration:

```bash
sudo ln -s /etc/nginx/sites-available/synapse.skpstack.uk /etc/nginx/sites-enabled/
```

Get the SSL Certificate:

```bash
sudo certbot --nginx -d synapse.skpstack.uk
```

**Correct this line after certbot:**

```nginx
listen 443 ssl http2;
```

```bash
sudo nginx -t
sudo systemctl daemon-reload
sudo systemctl restart nginx
```

# download synapse:

```bash
git clone https://github.com/sokinpui/synapse.go.git
sudo apt update
sudo apt install protobuf-compiler libprotobuf-dev
```

---

1. Start your gRPC server on port 50051 or the port you specified in the Nginx config.
2. You can now access your gRPC server via `synapse.skpstack.uk:443`.
