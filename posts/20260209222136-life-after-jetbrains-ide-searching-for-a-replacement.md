title: Life After JetBrains IDE: Searching for a Replacement
tag: ide, editor, git
--


After almost 1 year using JetBrains products, it’s time to switch because I don’t have the budget to subscribe to the IDE. So now mostly I use 2 IDEs: first Zed, and second is VS Code.

I mostly use Zed because it feels more snappy and is already good with Python and Go. On the other hand, VS Code is already mature with a lot of plugins, so sometimes I still go back and forth with these two IDEs. But there is still a missing piece in both of these IDEs compared to JetBrains: Test Suite and Conflict Resolver.

For me, test suite is not a big deal. Most of the time when I need to run tests, I prefer using the CLI, go test or uv run pytest, so it’s still okay. But I’m still searching for a replacement for how JetBrains has the best DX to resolve conflicts. Currently, I use or try to use VS Code and Sublime Merge with three-way diff.

![resolve conflicts using VS Code](static/images/vscode-conflict.png "vscode conflict")


And this is on Sublime Merge

![resolve conflicts using Sublime merge](static/images/sublime-merge-conflict.png "sublime conflict")

I still want to try a lot of tools that have DX like JetBrains, but most of the apps are already pricey, and if I must pay, I prefer just subscribing to JetBrains products.

You may ask why I said I use Zed but end up using another tool just to resolve conflicts. Zed is fun to use because it is more snappy, but the UI for resolving conflicts is not my preference. I prefer using VS Code or Sublime Merge for resolving conflicts. So before Zed implements a feature like JetBrains did, I choose to resolve conflicts using VS Code or Sublime Merge.

And for now, the journey is still in progress….
