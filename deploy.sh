#!/bin/bash

echo "resetting local changes..."
git reset --hard

echo "Pulling latest changes..."
git pull origin main
git rebase

echo "Cleaning previous builds..."
go clean

echo "Building the Go application..."
go build -o website.o

echo "Restarting the application using systemd..."
sudo systemctl restart my-website@$USER

echo "Deployment successful!"
