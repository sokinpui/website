#!/bin/bash

echo "resetting local changes..."
git reset --hard origin/main

echo "Pulling latest changes..."
git pull origin main

echo "Cleaning previous builds..."
go clean

echo "Building the Go application..."
go build -o website.o

echo "Restarting the application using systemd..."
sudo systemctl restart my-website

echo "Deployment successful!"
