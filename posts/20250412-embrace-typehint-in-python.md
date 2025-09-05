title: Embrace Typehint In Python
tag: python, django
--
I tried typehint in python and i didn’t like it, especially when working in legacy code, and then i move to other project, wrote golang code, feels good with the type, want to imitate the type in python and yes i visited again the typehint.

**Expectation**

Before we move forward, I must set expectation to myself. The default Python is dynamic typing, so it will not smooth like Go. Why? Let's take a look in this code:

```python
python
def must_int(param: int) -> int:
    return param

param = "this is string"
res = must_int(param)
```

You will see is strange code. The function give a hint the parameter must be int, and return int. But we use string and no break. The code still working as usual. So why we bother with typehint?

**TypeCheck and IDE**

Yes, you can't enforce type on runtime, but it not mean you can't check at all. You can use third party like mypy, pyright, or basedpyright.

With typecheck library, you can audit your code to found possible issue like code above. I use basedpyright, and when I run it to audit my code, I got information like this:

```bash
Users/ariesm/code/python/wip/playg/views.py
  /Users/ariesm/code/python/wip/playg/views.py:12:22 - error: Argument of type "Literal['string']" cannot be assigned to parameter "param" of type "int" in function "must_int"
    "Literal['string']" is not assignable to "int" (reportArgumentType)
1 error, 0 warnings, 0 notes
```

And because we use typehint in the code, with proper IDE we can generate autocomplete easily like this.

![image: Show auto complete django orm on vscode](/posts/images/typehint.png)

> In PyCharm, somehow even if you don't use typehint, you still can get autocompletion. But in VSCode, variable with typehint generate autocomplete, and variable without typehint doesn't have autocomple.

BasedPyRight

There are a lot typecheck library right, but I choose the basedpyright with django-types because it can add rules gradually. It works for me that working with legacy code.

Using django-types to generate types/stub for Django is game changer. Before, when I only use mypy, it driving me crazy how to setup proper type for Django project. With django-types, it become more easy. And with configuration in basedpyright, I can gradually put type on each Django app so it not add too much burden.

This is my config on my legacy Django project.

```toml
[tool.basedpyright]
# gradual, only add app "playg"
include = ["playg"]
exclude = ["**/migrations", "**/__pycache__", "**/settings.py"]
# stub coming from django stub/types
extraPaths = ["typings"] # prepare for stub
pythonVersion = "3.13"
pythonPlatform = "Linux"
typeCheckingMode = "basic"
venvPath = "."

# Enable missing imports detection for better error tracking
reportMissingImports = true
reportMissingTypeStubs = false

# Reduce noise from dynamically typed Django parts
reportUnknownMemberType = false
reportUnknownVariableType = false
reportUnknownArgumentType = false
reportUntypedClassDecorator = false
reportUntypedFunctionDecorator = false
reportUntypedBaseClass = false
reportIncompatibleVariableOverride = true
reportAttributeAccessIssue = false  # Only if the stub doesn't fully fix it

[tool.django-types]
django_settings_module = "core.settings"
```