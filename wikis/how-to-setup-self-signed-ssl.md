---
title: How to Set Up Self-Signed SSL
desc: how to create and configure self-signed SSL certificates for local development
---

# Use openSSL to create "self-signed" SSL certificates

To create self-signed SSL certificates for local development, we use `openssl` here.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout localhost.key -out localhost.crt
```

example nginx configuration:

```nginx
server {
  listen 80; # Listen for incoming HTTP requests
  listen 443 ssl; # Listen for incoming HTTPS requests
  server_name localhost;

  # SSL certificate configuration
  ssl_certificate /path/to/your/localhost+2.pem; # Or localhost.crt
  ssl_certificate_key /path/to/your/localhost+2-key.pem; # Or localhost.key

  # Recommended SSL protocols
  ssl_protocols TLSv1.2 TLSv1.3;

  location / {
    proxy_pass http://localhost:12352;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
  }
}
```

Restart Nginx to apply the changes:

```bash
sudo systemctl restart nginx
```

or

```bash
brew services restart nginx
```
