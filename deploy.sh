#!/bin/bash

echo "Pulling latest changes..."
git reset --hard origin/main
git pull origin main

echo "Building the Go application..."
go clean
go build -o website.o

echo "Restarting the application..."
sudo systemctl restart my-website

echo "Deployment successful!"
