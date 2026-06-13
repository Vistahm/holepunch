# HolePunch

**Share files directly between computers — no cloud, no accounts, no limits.**

HolePunch creates an instant, token/password protected file server that's accessible
from anywhere. It automatically configures your router via UPnP, so you don't
need to mess with port forwarding.

## Features

- 🔌 **Automatic NAT traversal** — uses UPnP to open ports on your router
- 🔑 **Built-in authentication** — token or Basic Auth
- 🔐 TLS encryption - with `--tls` flag, you can enable the encryption so your ISP won't be able to spy the data transit
- 👀 No surveillance - there's no middleman; only you and the people who have access to your URL will see the files
- 💻 **Single binary** — no dependencies, no installation
- 🛡️ **Survives restricted internet** — it will work in countries with restricted internet access
- 📁 **Simple directory listing** — browse and download files from any browser

## Installation

Download the latest binary from [Releases](https://github.com/Vistahm/HolePunch/releases)

Available for x86-64 architecture on Windows and Linux.

## Quick Start


 Download the binary for your platform, then:
 
 On Linux:

`./holepunch -dir ~/Music`

On Windows:

`.\holepunch.exe -dir D:\Music`


## Output:
```
Auth: Token-based
Token: a1b2c3d4e5f6a7b8
🔗 Local:   http://localhost:8080
🌐 Remote:  http://203.0.213.5:8080/?token=a1b2c3d4e5f6a7b8
```

## How It Works
- HolePunch starts an HTTP file server on your chosen port
- It discovers your router via UPnP/SSDP
- It asks the router to forward the external port to your computer
- It generates access credentials and displays the remote URL
- Remote users connect to your public IP and browse files


## Use Cases

- Share any file you want with whoever you want
- Give a known one access to files without uploading them anywhere
- Create a temporary file drop during a LAN party
- Access your files remotely, even when cloud services are blocked

## ⚠️ Disclaimer

HolePunch is provided "as is", without warranty of any kind. Use at your own risk.

- This tool opens a port on your router. Anyone with the access URL can reach your files.
- The author is not responsible for unauthorized access, data loss, or any damages.
- Respect copyright laws — only share files you have the right to distribute.
- UPnP behavior varies by router; not all networks are supported.
- Stop the program when you're done sharing — don't leave the port open unnecessarily.
- Never share the access URL publicly or with people you don't trust.
- If using `--tls`, keep the generated certificate files (`.crt`, `.key`) private.
