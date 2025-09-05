title: Hosting a blog on an old laptop
tag: blog, tech, webserver
--

Living in a third-world country with a weak currency compared to the US dollar makes hobby projects difficult sometimes. I host several of my applications on AWS Lightsail for 7 dollars, it’s cheap, right? But again, in a third-world country, 7 dollars is kinda "big." So how do we reduce costs? I was inspired by several friends who self-host their own blogs on a homelab server, so why not try it myself?

## Tools
I bought a used ThinkPad about 5 years ago. For almost a year, I haven’t used it because I already have another laptop. So I decided to use the ThinkPad again as a server.

## Software
For the OS, I chose Ubuntu 22.04 because I’m more familiar with this distro.

I’m using Coolify as a platform manager. Yes, I could have used a simple web server like Caddy or something similar, but hey, why not try a bit of over-engineering for a fun project, right?

To make my app publicly accessible, I used Cloudflare Tunnel.

I prefer not to expose my Coolify dashboard, but I still need to access it when I’m out, so I chose Tailscale.

And yes, for now, this blog, notes.arsmp.com is hosted in my own home, on my own laptop. It’s zero dollars, at least for now, and I hope it stays that way in the future haha. I just pay yearly for the domain, and for electricity, I think there’s not much difference in cost, since whether I use the laptop as a server or not, it's common in my home to put laptops and computers in sleep mode rather than shutting them down.
