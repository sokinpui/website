---
title: How to deploy the website
desc: Documentation for deploying this website
createdAt: "2025-11-15T10:00:00Z"
---

You can [setup zsh](./how-to-setup-terminal.md) first.

Guide for setup this website:

- build website binary
- Nginx for reverse proxy
- get SSL certificate with Certbot
- setup auto git pull with webhook

# Requirements

1. vps
2. domain name
3. SSL certificate

# config cloudflare DNS

1. Go to Cloudflare dashboard and select your domain
2. Go to DNS settings
3. Add two record with your domain name pointing to your VPS IP address

- Type: A

  - Name: @
  - IPv4 address: your VPS IP address
  - TTL: Auto
  - Proxy status: Proxied

- Type: A
  - Name: www
  - IPv4 address: your VPS IP address
  - TTL: Auto
  - Proxy status: Proxied

4. SSL/TLS settings
   - SSL/TLS encryption mode: Full (strict)
   - Always use HTTPS: ON
   - Automatic HTTPS Rewrites: ON

# Build website binary

1. Clone the repository to your VPS

```bash
git clone https://github.com/sokinpui/website.git
```

2. build the website

```bash
cd website
go clean
go build -o website.o
```

# Setup Nginx for reverse proxy

1. Install Nginx

```bash
sudo apt update
sudo apt install nginx
sudo systemctl start nginx
sudo systemctl enable nginx
```

2. Configure Nginx

```bash
sudo -e /etc/nginx/sites-available/www.skpstack.uk
```

Add the following configuration:

```nginx
server {
  listen 80;
    server_name www.skpstack.uk;

  location / {
    proxy_pass http://localhost:12352;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
  }
}
```

3. Enable the server block

```bash
sudo ln -s /etc/nginx/sites-available/www.skpstack.uk /etc/nginx/sites-enabled/
```

4. Test Nginx configuration and restart

```bash
sudo nginx -t
sudo systemctl restart nginx
```

# Get SSL certificate with Certbot

Detail guide can be found on [certbot official](https://certbot.eff.org/instructions)

1. Install Certbot

```bash
sudo apt update
sudo apt install python3 python3-dev python3-venv libaugeas-dev gcc
sudo apt remove certbot
```

2. Setup Certbot

```bash
sudo python3 -m venv /opt/certbot/
sudo /opt/certbot/bin/pip install --upgrade pip
sudo /opt/certbot/bin/pip install certbot certbot-nginx
sudo ln -s /opt/certbot/bin/certbot /usr/bin/certbot
```

3. Obtain SSL certificate

```bash
sudo certbot --nginx
```

4. Set up automatic renewal

```bash
echo "0 0,12 * * * root /opt/certbot/bin/python -c 'import random; import time; time.sleep(random.random() * 3600)' && sudo certbot renew -q" | sudo tee -a /etc/crontab > /dev/null
```

# Setup auto git pull with webhook

## Setup webhook listener

```bash
sudo apt install git webhook
```

modify `hooks.json` file in `website` directory:

```
sed -i "s|<username>|$USER|g" hooks.json
```

example:

```json
[
  {
    "id": "redeploy-website",
    "execute-command": "/home/<username>/website/deploy.sh",
    "command-working-directory": "/home/<username>/website",
    "response-message": "Deployment initiated.",
    "trigger-rule": {
      "and": [
        {
          "match": {
            "type": "payload-hmac-sha256",
            "secret": "abc",
            "parameter": {
              "source": "header",
              "name": "X-Hub-Signature-256"
            }
          }
        },
        {
          "match": {
            "type": "value",
            "value": "refs/heads/main",
            "parameter": {
              "source": "payload",
              "name": "ref"
            }
          }
        }
      ]
    }
  }
]
```

## create webhook systemd service

1. create a systemd service to run webhook

```bash
sudo -e /etc/systemd/system/webhook.service
```

```ini
[Unit]
Description=Webhook
After=network.target

[Service]
User=<username>
Group=<username>
ExecStart=webhook -hooks /home/<username>/website/hooks.json -verbose
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

2. Start and enable the service

```bash
sudo systemctl daemon-reload
sudo systemctl start webhook.service
sudo systemctl enable webhook.service
```

## Create website systemd service

1. create a systemd service to serve the website binary

```bash
sudo -e /etc/systemd/system/my-website.service
```

```ini
[Unit]
Description=My Go Website
After=network.target

[Service]
User=<username>
Group=<username>
WorkingDirectory=/home/<username>/website
ExecStart=/home/<username>/website/website.o
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

2. Start and enable the service

```bash
sudo systemctl daemon-reload
sudo systemctl start my-website.service
sudo systemctl enable my-website.service
```

add Webhook to nginx config:

```bash
sudo -e /etc/nginx/sites-available/www.skpstack.uk
```

```bash
location /hooks/ {
    proxy_pass http://127.0.0.1:9000/hooks/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
}
```

```bash
sudo nginx -t
sudo systemctl reload nginx
```

## allow run restart my-website.service without sudo

```bash
sudo visudo
```

add this line at the bottom

```
<username> ALL=NOPASSWD: /usr/bin/systemctl restart my-website
```

You should be able to access the website now

---

# Development

Refresh the website:

- Chrome, Firefox, or Edge (Windows/Linux): `Ctrl + Shift + R` or `Ctrl + F5`
- Chrome, Firefox, or Edge (Mac): `Cmd + Shift + R`
- Safari (Mac): `Cmd + Option + R`

Restart the website:

```bash
cd ~/website
./deploy.sh
```

## clear cluodflare cache

sometimes after deploy the website, the old cache is still there. You can clear the Cloudflare cache by following these steps:

1. Log in to Cloudflare: Go to the Cloudflare dashboard https://dash.cloudflare.com and log in to your account.
2. Select Your Website: On the dashboard home, click on the domain name you want to manage (e.g., xrayxrtas.space ).
3. Navigate to Caching: In the left-hand sidebar, find and click on the Caching icon (it looks like a cylinder or database).
4. Go to Configuration: Within the Caching section, make sure you are on the Configuration tab.
5. Purge Everything: You will see a section called "Purge Cache". For your situation, the simplest and most effective option is to
   purge everything.
   • Click the "Purge Everything" button.
   • A confirmation dialog will appear. Confirm that you want to purge the entire cache.
