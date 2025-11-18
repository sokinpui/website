#!/bin/bash

echo "resetting local changes..."
git reset --hard origin/main

echo "Pulling latest changes..."
git pull origin main

echo "Building the Go application..."
go clean
go build -o website.o

echo "Restarting the application..."
sudo systemctl restart my-website

echo "Deployment successful!"
