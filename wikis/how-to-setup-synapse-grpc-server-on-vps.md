---
title: "How to setup grpc server on vps"
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
    # or
    # server_name <your-domain>;

    access_log /var/log/nginx/grpc_access.log;
    error_log /var/log/nginx/grpc_error.log;

    # The location matches the "package.Service" name from your .proto file
    location /synapse.Generate/ {
        grpc_pass grpc://localhost:50051;
        # or
        # grpc_pass grpc://<ip>:<port>;

        # increase timeouts if needed
        #grpc_read_timeout 600s;
        #grpc_send_timeout 600s;
        #grpc_connect_timeout 60s;

        grpc_set_header Host $host;
        grpc_set_header X-Real-IP $remote_addr;
        grpc_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    location / {
        return 404;
    }

    listen 80;
}
```

Enable the site and test Nginx configuration:

```bash
sudo ln -s /etc/nginx/sites-available/grpc.skpstack.uk /etc/nginx/sites-enabled/
```

Get the SSL Certificate:

```bash
sudo certbot --nginx -d grpc.skpstack.uk
```

```bash
sudo nginx -t
sudo systemctl daemon-reload
sudo systemctl restart nginx
```

Correct this line after certbot:

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
2. You can now access your gRPC server via `grpc.skpstack.uk:443`.
