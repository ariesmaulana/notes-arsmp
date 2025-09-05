title: Goodbye Coolify
tag: blog, go, docker, coolify, web
--

Several months ago I moved all my services to my old laptop, I wrote about this [here](https://notes.arsmp.com/post/hosting-a-blog-on-an-old-laptop), but the honeymoon era was not that long. There were several issues and I think it was too bloated for me.

1. HomeLab

I installed Coolify in the old laptop, I can access Coolify only via Tailscale, so I can't get the full benefit of Coolify services. You know, something like auto deployment for simple Docker. Maybe I have a skill issue about this, but I skipped it.

2. Docker

Coolify supports Docker which is good, you can set multiple projects in Docker and Coolify will handle it, but sometimes Docker is not needed. Like I just run a Go service in systemd, pass it to config Cloudflared and it's done, simple.

3. Docker compose is enough

I host my [blog.arsmp.com](https://blog.arsmp.com/) and this site in my local laptop. The first one is a blog made by Ghost, and this one I just built a simple web app that reads markdown files. With simple Docker and systemd service it's enough. I can just ssh to my local laptop, run docker compose reload or systemd restart and done. And if you use docker compose and want simple "auto deployment" using Watchtower or a cron job is enough.

So yeah now my entire setup is just like this:

1. Systemd service  
2. Docker (and docker compose)  
3. Cloudflare tunnel
