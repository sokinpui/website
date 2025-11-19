---
title: How to setup terminal environment
desc: Documentation for setup terminal config for new *nix device
createdAt: "2025-11-17T10:00:00Z"
---

# Terminal setup

run this script:

```bash
wget -O- https://raw.githubusercontent.com/sokinpui/terminal_dotfiles/refs/heads/main/zsh/setup | bash
```

It should install:

- zsh
- tmux
- neovim
- fd
- ripgrep
- fzf
- git
- curl
- pipx
- zoxide
- pyenv
- direnv
- lf
- wtgo
- itf
- pcat
- trash-cli

# Docker setup

```bash
sudo apt update
sudo apt install ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

sudo tee /etc/apt/sources.list.d/docker.sources <<EOF
Types: deb
URIs: https://download.docker.com/linux/ubuntu
Suites: $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}")
Components: stable
Signed-By: /etc/apt/keyrings/docker.asc
EOF

sudo apt update
```

```bash
sudo apt install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```

```bash
sudo systemctl start docker || sudo systemctl restart docker
```

```
sudo groupadd docker
sudo usermod -aG docker $USER
newgrp docker
```

# Nvidia

## Ubuntu

get the nvidia drivers

```bash
sudo apt update
sudo ubuntu-drivers autoinstall
```

then reboot

```bash
sudo reboot
```

# Build neovim from source

```bash

sudo apt update
sudo apt install cmake

git clone https://github.com/neovim/neovim.git
cd neovim
make CMAKE_BUILD_TYPE=RelWithDebInfo
sudo make install
sudo apt remove neovim
```

## Setup neovim plugin

install `nodejs` and `npm`
refer to https://nodejs.org/en/download/package-manager/#debian-and-ubuntu-based-linux-distributions

```bash
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash
\. "$HOME/.nvm/nvm.sh"
nvm install 24
node -v
npm -v
```
