title: Django Update Fields
tag: python, django
--
Usually when working with Django ORM and want to update data, there are two patterns I always use: .save() and .update().

When using .save(), the code becomes like this:
```python
model = Model.objects.get(id=1)
model.field = val
model.save()
```
This code will act like this:
- Run clean() on Django model
- Run the actual query
- Run related signal(s)

Using .update()

When using .update(), usually it's to change multiple data. So instead of model instance, we use queryset instance.
```python
Model.objects.filter(param=param).update(field=value)
```
This code only runs the generated query. No signal or clean method affected.

**Generated Query**

Based on information above, the difference between .save() and .update() is how the code runs in the system. One calls internal functions like clean() and signals, and the other only runs the query. But it's not only that, the generated query is also different.

Let's take a look at both snippets. In the snippet I only change 1 value for 1 field, right? But if we look at the generated query...

With .save()
```sql
update model_table set field='xxx', other='xxx' ... where id=pk
```

With .update()
```sql
update model_tabel set field=value where param=param
```

It bugs me how the .save() generated query works. I just want to change 1 field, but other fields suddenly appear. The proper query is from .update(), but .update() doesn't call clean() and signal (which is fine, I’m not big fan of signal).

But what if I need the signal, and still want the query to only update the correct fields? update_fields is the answer. You just change the code like this:

```python
model = Model.objects.get(id=1)
model.field = val
model.save(update_fields=["field"])
```
It will generate the query:
```sql
update model_tabel set field=value where id=pk
```