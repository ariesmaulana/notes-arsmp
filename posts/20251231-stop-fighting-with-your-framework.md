title: Stop Fighting With Your Framework or Language
tag: blog, go, python, programming
--

I have a bad habit of enforcing my ideals on everything, and this includes programming. In the early days, I tried to use everything that sounded so good in my work, but it turned out it was not doing well. So this is my journey about how stupid I am.

When I first started using Laravel, I really liked it, but I read too many “how to make it better” articles. One of the articles I read was about implementing some repository pattern in Laravel. The article was so good at explaining the benefits of this pattern, and me, as a young programmer who wanted to level up, blindly followed the article. And guess what? I implemented a poor repository pattern and just built a mediocre wrapper.

Second, I learned Go almost 3 years ago, worked professionally (I mean using Go in actual work that generates money, not just a hobby) around 2 years ago, and although a lot of people think error handling in Go is boilerplate, for me I really like it. And besides Go, I like Python too. So guess what? I tried to implement err as a value in every function. In some cases it worked well, but in some cases it became a headache, usually when working with context managers that rely on exceptions.

So what’s now? Now I just try to be conservative. If I use an opinionated framework like Laravel or Django, I just follow what the docs say, or if I build something from “scratch”, I just follow the guidelines from the language itself.

Yes, don’t battle with your tools, just use them and follow the guidelines. And if you really want something slightly different, just consider the pros and cons.