title: The Day I Realized I Don’t Need Half My Tools
tag: linux, docker, systemd, python, go, svelte
--

Another day, another removed dependency. When I first moved my entire pet project from AWS Lightsail, I used Coolify to handle the deployment process. However, I found that Coolify was too bloated for me, so I just used pure Docker Compose. After that, I realized that I could just use systemd to manage my services.

I play around with various languages, mostly in the web ecosystem. Currently on my machine I have 1 Svelte app, 2 Python apps (Django & FastAPI), and 1 Go app — this blog is the Go app I mentioned. All of them were previously installed using Docker Compose. Then, for Go, since I can just build it into a binary and register it with systemd, I chose that way. So I just built it, installed it with systemd, set up a reverse proxy using Caddy, and done.

Second, for Django I did the same. I just cloned it from GitHub, installed packages using uv, registered the Granian web server to systemd, added another reverse proxy, and done.

Now, the remaining apps installed via Docker are Svelte and FastAPI. For FastAPI it's “easy” because I can just copy what I did with Django, but for Svelte I'm still thinking about it — can it just serve as static web with build, or do I need to use pm2 to manage it?

To be honest, I want to try overengineering everything. I want to install k8s or at least Minikube, but somehow I don’t have enough time to do it, or maybe it’s just not my priority right now. And it turns out I really love working with minimal dependencies. So I still want to try Minikube someday, but for now, my systemd, Caddy, and of course Cloudflared are enough for my pet projects.
