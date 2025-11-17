---
title: How to deploy the website
desc: Documentation for deploying this website
---

Guide for setup this website:

- build website binary
- Nginx for reverse proxy
- get SSL certificate with Certbot
- setup auto git pull with webhook

# Requirements

1. vps
2. domain name
3. SSL certificate

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
vim /etc/nginx/sites-available/your_domain
```

Add the following configuration:

```nginx
server {
  listen 80;
    server_name your_domain;

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
sudo ln -s /etc/nginx/sites-available/your_domain /etc/nginx/sites-enabled/
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

The `deploy.sh` script is included in the repository

1. setup a simple program to listen webhook request

```bash
mkdir ~/github-webhook
cd ~/github-webhook
go mod init webhook
```

2. Create `main.go` file

```go
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// This should be securely stored, for example, as an environment variable.
const secret = "p@ssw0rd"
const repoPath = "/path/to/your/website/repo" // Change this to the actual path of your git repository on the server

func main() {
	http.HandleFunc("/webhook", handleWebhook)
	log.Println("Listening for webhooks on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Get the GitHub signature from the header
	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		http.Error(w, "Missing signature", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	if !isValidSignature(body, signature) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Run the deployment script in a separate goroutine to avoid blocking the response
	go func() {
		cmd := exec.Command("/bin/sh", "./deploy.sh")
		cmd.Dir = repoPath
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Error running deployment script: %s\n%s", err, output)
			return
		}
		log.Printf("Deployment script output:\n%s", output)
	}()

	fmt.Fprint(w, "Webhook received and processed")
}

func isValidSignature(body []byte, signature string) bool {
	// The signature from GitHub comes in the format "sha256=..."
	// We need to remove the "sha256=" prefix
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	actualSignature := strings.TrimPrefix(signature, "sha256=")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(actualSignature), []byte(expectedMAC))
}
```

3. create systemd service file

```bash
sudo -e /etc/systemd/system/my-website.service
```

Add the following content:

```ini
[Unit]
Description=My Go Website
After=network.target

[Service]
WorkingDirectory=/home/<username>/website
ExecStart=/home/<username>/website/website.o
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

4. Start and enable the service

```bash
sudo systemctl daemon-reload
sudo systemctl start my-website.service
sudo systemctl enable my-website.service
```

Make sure build the website binary before start the service.

5. Run the webhook listener

```bash
go run .
```

6. Configure Nginx for webhook listener

```bash
vim /etc/nginx/sites-available/your_domain
```

Add the following location block inside the server block:

```nginx
location /webhook {
    proxy_pass http://localhost:8080; # Port your webhook listener runs on
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

7. Test Nginx configuration and restart

```bash
sudo nginx -t
sudo systemctl reload nginx
```

---

You should be able to access the website now
