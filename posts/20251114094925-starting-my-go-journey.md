title: Starting My Go Journey
tag: blog, web, python, go
--

Even though my web dev journey started from PHP — which was pretty common back then with HTML + PHP + JS — I ended up loving Python the most.
Yes, Python is not the fastest language, probably one of the slowest, but I think it’s kind of a trade-off. For me, Python gives one of the best developer experiences — except for managing packages. Thanks to uv, maybe that part’s finally fixed.

When my friends ask me and/or I need to set up a quick project, I always choose Django or FastAPI. I even built my own “starter kit” based on FastAPI. Again, that’s how much I love Python — it works for simple scripting up to full-fledged web apps.

For almost one year, I’ve been trying to move my starter kit to Go. Although I’ve already used Go for more than a year — in my current company we use Go as the main language — I still used Python for my pet projects. Now it’s time to move on, and this is the reason I’m rewriting my starter kit in Go.

## Types

Python is duck-typed, and I’m fine with that. Actually, when I tried Go for the first time, I didn’t really like types. But over time, I started to love having types.
I know in the current Python era we can use type checkers like _mypy_ or _basedpyright_, or for runtime, use _pydantic_. But I think it would be much nicer if that feature was already built into the language itself.

## Dev & Deployment

From what I remember, I always built Python projects using Docker or virtualenv — mostly Docker — and I was fine with it.
I still use Docker for databases if I need to install multiple DBs for debugging, but when I tried Go, most of the time I just use my machine directly.
For deployment, I just build the binary, register it to systemd, and done.
I know a lot of Go projects I found on GitHub still use Docker, but again, if before it was necessary to add Docker, with Go it’s just “nice to have.”

## Actually, I love Go code

People often say Go code is too verbose, and don’t forget the memes about `err != nil`. But actually, I really like the verbosity — it’s easy to read and probably easier to understand.
People often say `err != nil` is unnecessary because it pollutes the main business process, but when you think about it, business processes aren’t about the happy path — it’s about handling when something goes wrong too.
So the verbosity of `err != nil` for me is actually necessary.

## Conclusion

I still love Python — especially Django and FastAPI. My starter kit “steals” a lot from Django and FastAPI. But for now, Go is my go-to language for web dev. I’m excited to see what I can build with Go in the future.
