title: AI Benefit: Reduce Dependencies
tag: ai, python, node, php, docker, go, test
--

When i started my career in this industry i got suggestions of always believe the giants or we must stand on shoulder of giant, with that mindset in my early days it translated to me always use popular packages/libraries whenever possible. So whenever i got an issue first thing first is what kind of package/library i need for this. Turns out i become a composer require addict, pip install as much as possible and npm install enthusiast.

## A Package Is Not Silver Bullet

Several years ago i got called from my colleagues, he asks me "why some of our pipelines is taking too long to just build a docker image?" After investigation we found the culprit, we install a pandas, a popular python library for data manipulation. I ask him "why we need this?" He said "i have a task to import excel file."

Using pandas literally fix the task, import excel file working, but again do we literally need this package? Yes, pandas is great, it can be used for excel import and more than that and pandas is popular and maybe always become top of mind of a lot of people to use it. But using pandas just because the package can be used for import excel is like using your sports car just to visit your neighbor. After we talk about the specs of task, we conclude we remove pandas and use another package that may not be as rich as pandas but good enough for our task, the pipeline now working normally.

## A Hack

I talk with my other senior colleague what happen if we found this issue again, there are several suggestions if i remember correctly. First fork the package, reduce unnecessary features and dependencies. This approach we can use the feature we want from the package but to set up this we must have good knowledge about the package, remove the feature somehow have the same difficulty as creating a feature. Other approach is to make a new package with "copy paste" from the popular one, and third, just use other package.

## AI

Building full-fledged app without third party packages somehow become harder, in some ecosystem everything is a package. Glad now i use go, basic functionality already exist inside the standard library. But still some packages are still needed for the app to work, for example, test suites. I building my own starter kit in go, i still use some third party packages for the app to work but i need some packages for the test suites. The default from the standard library is enough for some little functions but for my needs i need other packages and it is hard to find a specific package that fits my needs. The problem always becomes like this, i found a good one but is too bloated and somehow it looks like forking from different language ecosystem, or, i found the package that is too simple and if too simple i prefer to use the standard library instead of third party.

With the current AI, and i know what i want, i know what design for my test suites, and i know it is too specific for my needs i just write a plan, write pros and cons, write the needs for my test suites, and let the AI do the rest. To be fair the result is acceptable, is good enough for my starter kit, and it just extends from the standard library.

## Conclusion

The point of my thinking is that i need a libs/package that is usually for a specific function or specific needs, and i understand some libraries and packages will be as general as possible, so with the AI capabilities, time for "build my own package" have same amount --and maybe more or less -- time to search a library or package that already exist.
