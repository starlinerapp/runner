# Starliner Runner

The runner is the application that runs the builds in Starliner.

## Get Started
Download the latest release 
```
curl -L -o runner-v0.0.1-linux-amd64.tar https://github.com/starlinerapp/runner/releases/download/v0.0.1/runner-v0.0.1-linux-amd64.tar
```
Extract the archive:
```
tar xf runner-v0.0.1-linux-amd64.tar
```
Create a VM:
```
cd runner-v0.0.1-linux-amd64
sudo ./runner vm create
```