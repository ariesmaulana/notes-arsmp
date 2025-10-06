title: Avoid Django Admin
tag: blog, web, python, django
--

One of the best features Django provides is the built-in "admin dashboard". In the early days of your app, sometimes it's always the priority of end users, I mean public users, not internal users. But although we focus on public users, we still need to maintain the CRUD, so we must provide a dashboard for internal users. With the Django admin, we can skip that part for some weeks.

Django admin itself is really helpful in the early stage of an app, but I recommend never relying on Django admin. The reason is too much magic, and sometimes it's "hard" to customize.

## Design

I'm not a designer, I suck at design, I'm bad at CSS, but if you just build some HTML + CSS like most common web apps, it's still "easy" to modify. On the other hand, Django admin's approach is programmatic. In your "admin.py", you can define fields, filters, and actions — all of them programmatically — and it will show your "list" page. But be careful, when you need some field related to another field or the filter depends on other models, beware of performance issues. I have experienced one Django admin page where whenever people visited, the resource usage spiked.

## Form

Again, you can declare fields and field behavior using your "admin.py", or if you want a more programmatic way, you can add "forms.py". And when a lot of CUD happens, it will raise a common question: "Where do I put the validation?" Yes, you can put common validators in "models", or you can put them in "forms", or maybe in "admin.py". But still, to answer that, you must have experience or, in rare cases, read the documentation carefully. Compare it to your HTML + CSS and JS, maybe you can easily put validation on the frontend using HTML or JS, and the validator itself is just in "views.py".

## Trade-Off

Like I said before, in the early stage Django admin is good and helpful, but please don't wait until your app becomes too big and people are already comfortable with Django admin. Otherwise, you’ll end up thinking not about "building the dashboard" but "migrating the dashboard".