title: Building a Microservice Application
tag: go,k3s,microservices
--

As the title suggests, this post is about my attempt to build a web application using a microservice architecture. I'm not sure whether my approach follows all the best practices or not, but I'd like to walk through what I built and why I made those decisions.

# Motivation

If you've ever looked for a job as an experienced or senior software engineer, you've probably come across questions like, "Have you worked with microservices?" or, more specifically, "Have you built a microservice application on Kubernetes?"

On another occasion, I joined a project that claimed to use a microservice architecture. Whether my impression was right or wrong, it didn't really feel like microservices to me. From what I've read over the years, a microservice architecture is often described as one where each service can operate independently. As long as the communication contract stays the same, you should be able to deploy or even refactor one service without affecting the others.

By the way, the application is already up and running. Feel free to check it out here: [https://court-api.arsmp.com](https://court-api.arsmp.com)
```
username: demouser
password: demo123
```

Btw, here's the approach I took.

# Responsibilities of Each Service

The application itself is a simple sports court management system. It's nothing fancy—just a CRUD application. Although calling something "simple" and "microservice" in the same sentence feels a little strange, I decided to keep things small and only built two services.

- user-service: Responsible for everything related to users, including authentication, login, and user management.
- court-service: Responsible for managing sports courts.

Each service has its own repository and its own database.

# Technology Stack

Both services are written in Go and use PostgreSQL as their database. There's also a single Redis instance shared between both services for storing permission-related information.

To simulate a Kubernetes environment, I used k3s. I'm running everything on a $10 VPS that also hosts several of my personal websites, including this blog.

# Infrastructure and Communication
Since I'm basically a CRUD engineer, I built both services the same way I'd normally build a Web API. Each service exposes its own API contract.

The first question that came to mind was, "How does the frontend know which service to talk to?"

Fortunately, the answer was straightforward. Traefik, acting as a reverse proxy inside k3s (and Kubernetes in general), handles this routing without much effort.

But that only solves part of the problem.

The next question is: what about authentication and authorization?

# Authentication and Authorization
In a monolithic application, this is usually solved with a middleware.

But what about a microservice architecture?

After spending some time researching, I came to the conclusion that authentication should remain the responsibility of user-service. It should be the only service responsible for validating login sessions, JWTs, and everything related to user identity.

Authorization, however, felt a little different.

In my approach, user-service owns all permission data and assigns permissions to users. Other services, like court-service, simply need a way to verify whether the current user has the required permission.

The remaining question was: how should that permission check be implemented?


## First Approach: Direct Service-to-Service Communication

My first idea was fairly straightforward.

For authentication, I duplicated the JWT validation middleware inside court-service. It's "just" JWT validation after all.

For authorization, court-service would simply ask user-service whether the current user had the required permission.

The code looked something like this:
```go
func (s *service) Update(ctx context.Context, input *UpdateInput) *UpdateOutput {
	resp := &UpdateOutput{}
	if s.user.CheckPermission(ctx, input.RequesterId, "permission") {
		// do something
	}
	return resp
}
```
The more I looked at this approach, the less I liked it.

Both services now shared authentication logic, and court-service became heavily dependent on user-service just to process a request.

That didn't feel right, so I scrapped this idea and moved on to a different approach.

## Second Approach: ForwardAuth + Redis

I'm still using JWT, but this time I configured Traefik so that every endpoint under the /dashboard prefix is considered protected.

Whenever a request hits one of those endpoints, Traefik sends it to a custom ForwardAuth middleware, which calls user-service/api/verify.
```
Internet 
↓ 
Authorization: Bearer <JWT> 
↓ 
Protected Endpoint (/dashboard/*)
↓
ForwardAuth Middleware 
↓ user-service/api/verify 
↓ Success → Continue request Failed → Return HTTP 401
```
This worked pretty well.

However, court-service still needed to parse the JWT just to know which user was making the request.

Since the token had already been verified, parsing it again felt unnecessary.

So I changed the flow slightly.

Instead of only validating the token, the middleware removes any incoming X-Auth-* headers, verifies the JWT, and injects a new trusted header containing the authenticated user's ID.

The flow becomes:
```
Internet
    ↓
JWT + Custom Header
    ↓
ForwardAuth Middleware
    ↓
Remove incoming X-Auth-* headers
    ↓
Verify JWT
    ↓
Inject verified user ID into X-Auth-User
    ↓
Forward request
```
At this point, authentication is completely handled by Traefik and user-service. Since clients cannot inject those headers themselves, court-service can safely trust that X-Auth-User always comes from Traefik.

That solves authentication.

What about authorization?

Instead of calling user-service every time, court-service reads permissions directly from Redis.

The synchronization flow is simple:

user-service

Create or update permissions
Save them to PostgreSQL
Synchronize them to Redis

court-service

Read the authenticated user's permissions directly from Redis using the user ID from X-Auth-User

Overall, the architecture looks something like this:

```
                                              Internet
                                                  |
                                             HTTPS Request
                                                  |
                                                  v
                                       +-----------------------+
                                       |    Traefik Ingress    |
                                       |    Reverse Proxy      |
                                       +-----------------------+
                                                  |
                  +-------------------------------+-------------------------------+
                  |                                                               |
                  |                                                               |
          Public Endpoint                                                Protected Endpoint
      (tidak perlu autentikasi)                                        (/dashboard/*)
                  |                                                               |
                  |                                                       ForwardAuth
                  |                                                               |
                  |                                                               v
                  |                                                +--------------------------+
                  |                                                |       user-service       |
                  |                                                |      /api/verify         |
                  |                                                +--------------------------+
                  |                                                               |
                  |                                         JWT Valid? -----------+----------+
                  |                                              |                           |
                  |                                           Invalid                     Valid
                  |                                              |                           |
                  |                                              |                           |
                  |                                              |                     Inject Header
                  |                                              |                  X-Auth-User: <id>
                  |                                              |                           |
                  |                                              |                           |
                  |                                              +-----------> 401 <---------+
                  |                                                               |
                  +------------------------------+--------------------------------+
                                                 |
                          +----------------------+----------------------+
                          |                                             |
                          |                                             |
                          v                                             v
                +-------------------------+                  +-------------------------+
                |      user-service       |                  |     court-service       |
                |                         |                  |                         |
                | Public API              |                  | Public API              |
                | Dashboard API           |                  | Dashboard API           |
                +-------------------------+                  +-------------------------+
                     |             |                               |             |
                     |             |                               |             |
                     |             |                               |             |
                     v             |                               |             v
             +---------------+     |                               |     +---------------+
             | PostgreSQL    |     |                               |     | PostgreSQL    |
             | User DB       |     |                               |     | Court DB      |
             +---------------+     |                               |     +---------------+
                     ^             |                               |             ^
                     |             |                               |             |
                     | Read/Write  |                               | Read Only   |
                     | Permission  |                               | Permission  |
                     | Cache       |                               | Cache       |
                     |             |                               |             |
                     +-------------+---------------+---------------+-------------+
                                                   |
                                                   v
                                         +----------------------+
                                         |        Redis         |
                                         |   Permission Cache   |
                                         +----------------------+
```

So, is this microservice architecture or just distributed monolith?
