title: ACL at Microservices
tag: go,k3s,microservices
---

> Previous post: [Building a Microservice Application](/post/building-a-microservice-application)

In the previous post I already built two services that successfully talk to each other all the way up to managing access rights. As a reminder, the technical choices I made look like this:

1. Access management lives in *user-service*
2. *user-service* stores the information in the database and *syncs* it to *Redis*
3. Leveraging the *forward auth* feature from k3s to make sure several pages are protected
4. Other services read the value from *Redis*

## Optimizing *Forward Auth*

With the current approach I'm using *forward auth* and hitting the *verify* function, which has a pretty simple job:

1. Clear the *custom header X-User-Id*
2. Validate the JWT *header*
3. Send the *id* from the JWT into *X-User-ID*

With this approach other services can confidently read the *header* information because it comes from the function, not from user input. But for me this feels wasteful, why are we only sending the *ID* when *user-service* can hand us the access rights along with it? So, on top of sending the *ID*, the function now also sends the access rights, and its job shifts slightly to:

1. Clear the *custom headers X-User-Id* and *X-User-Permissions*
2. Validate the JWT *header*
3. Send the *id* from the JWT into *X-User-ID*
4. Send the access rights from the JWT into *X-User-Permissions*

## Changes in the Other Services

Now that we can grab the access rights straight from the *header*, the need for *Redis* goes away, so this app only needs a database connection and no *Redis* at all.

There's one worry, actually, when focusing on sending things in the *header*: how many bytes are we pushing around? But after getting *claude* to help me figure out the limit, it turns out you'd need hundreds of thousands of permissions to even hit the *header* limit, and that's never going to happen, so let's just ignore it.
