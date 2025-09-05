title: Understanding Django Queries
tag: python, django, orm, sql, database, sqlite, postgre, json
--

Django ORM is great, but remember, every Django query we write gets translated into raw SQL.

In this post, I’ll show how Django generates those queries, the cost behind them, and how different ORM methods affect performance, in terms of memory, execution time, and how the query is generated.

This post won't cover indexing or general SQL optimization techniques, it focuses purely on how Django ORM translates queries into raw SQL.

To get the information in this test, I was helped by Claude and GPT to create a simple profiler for this kind of task. You can access it here:

> https://gist.github.com/ariesmaulana/654ec4a553ddc15189cc50f4ac85f25b

## Data
For this post, I created one big model in a single app. The model looks like this:

```python
class BenchDataPostgres(models.Model):
    name = models.CharField(max_length=255)
    email = models.EmailField()
    age = models.IntegerField()
    is_active = models.BooleanField(default=True)
    joined_date = models.DateTimeField()
    score = models.DecimalField(max_digits=10, decimal_places=2)
    description = models.TextField()
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)
    balance = models.FloatField()
    tags = models.JSONField(null=True, blank=True)
    preferences = models.JSONField(null=True, blank=True)
    history = models.JSONField(null=True, blank=True)
    notes = models.TextField()
    status = models.CharField(max_length=50)
    metadata = models.JSONField(null=True, blank=True)

    class Meta:
        verbose_name = "Benchmark Data PostgreSQL"
```

And i generated 1500 data to this model.

Okay, let’s begin!

The simplest Django query to get all data from that model is like this:

```python
BenchDataPostgres.objects.all()
```
So, what kind of raw SQL is generated from this? Django translates it into a full SELECT query with all fields listed.
```bash
Query Generated: SELECT "bench_benchdatapostgres"."id", "bench_benchdatapostgres"."name", "bench_benchdatapostgres"."email", "bench_benchdatapostgres"."age", "bench_benchdatapostgres"."is_active", "bench_benchdatapostgres"."joined_date", "bench_benchdatapostgres"."score", "bench_benchdatapostgres"."description", "bench_benchdatapostgres"."created_at", "bench_benchdatapostgres"."updated_at", "bench_benchdatapostgres"."balance", "bench_benchdatapostgres"."tags", "bench_benchdatapostgres"."preferences", "bench_benchdatapostgres"."history", "bench_benchdatapostgres"."notes", "bench_benchdatapostgres"."status", "bench_benchdatapostgres"."metadata" FROM "bench_benchdatapostgres"
Mem Usage: 39.71 MB
Exec Time: 605.66 ms
```
That’s pretty much the same as SELECT * FROM bench_benchdatapostgres, right?

With this approach, we end up retrieving every field defined in the model, but sometimes, we don’t need all of them. For example, if you have two pages, one for a list and one for a detail view, the list view probably only needs a few fields. Django provides several ways to optimize for this.

### `ONLY()`
As the name suggests, it lets us select only the fields we want. For example, if we only want name, email, and age, we can use `.only()` to limit the data retrieved.

The resulting query will only include those fields, and you’ll see a noticeable improvement in both memory usage and execution time.

Give this example, we change the query above to become like this
```python
BenchDataPostgres.objects.all().only("name", "age", "email")
```
This will translated to:
```bash
Query Generated: SELECT "bench_benchdatapostgres"."id", "bench_benchdatapostgres"."name", "bench_benchdatapostgres"."email", "bench_benchdatapostgres"."age" FROM "bench_benchdatapostgres"
Mem Usage: 1.29 MB
Exec Time: 196.82 ms
```
See the memory usages and execution times, is major improvement right?

### `DEFER()`

Unlike `.only()`, which includes specific fields, `.defer()` lets us exclude certain fields.

Let’s say we want to exclude metadata and history because they contain large JSON data. We can use `.defer()` to skip loading those fields initially.

This also results in improved performance.

```python
BenchDataPostgres.objects.all().defer("metadata","history")
```

And will translated to:
```bash
Query Generated: SELECT "bench_benchdatapostgres"."id", "bench_benchdatapostgres"."name", "bench_benchdatapostgres"."email", "bench_benchdatapostgres"."age", "bench_benchdatapostgres"."is_active", "bench_benchdatapostgres"."joined_date", "bench_benchdatapostgres"."score", "bench_benchdatapostgres"."description", "bench_benchdatapostgres"."created_at", "bench_benchdatapostgres"."updated_at", "bench_benchdatapostgres"."balance", "bench_benchdatapostgres"."tags", "bench_benchdatapostgres"."preferences", "bench_benchdatapostgres"."notes", "bench_benchdatapostgres"."status" FROM "bench_benchdatapostgres"
Mem Usage: 8.66 MB
Exec Time: 431.63 ms
```

However, there’s a hidden trap when using `.only()` and `.defer()`. Even though you’re selecting or excluding certain fields, Django won’t stop you from accessing any field later on.

Lets take example in this code
```python
data = BenchDataPostgres.objects.all().only("name", "age", "email")
metadata = []
for d in data:
    metadata.append(d.metadata)
return metadata
```
The code will run perfectly, no error, no raise exception, but let see the generated query
```bash
Queries Generated (1501 total):
  1. SELECT "bench_benchdatapostgres"."id", "bench_benchdatapostgres"."name", "bench_benchdatapostgres"."email", "bench_benchdatapostgres"."age" FROM "bench_benchdatapostgres"
  2. SELECT "bench_benchdatapostgres"."id", "bench_benchdatapostgres"."metadata" FROM "bench_benchdatapostgres" WHERE "bench_benchdatapostgres"."id" = 1 LIMIT 21
  3. SELECT "bench_benchdatapostgres"."id", "bench_benchdatapostgres"."metadata" FROM "bench_benchdatapostgres" WHERE "bench_benchdatapostgres"."id" = 2 LIMIT 21
  4. SELECT "bench_benchdatapostgres"."id", "bench_benchdatapostgres"."metadata" FROM "bench_benchdatapostgres" WHERE "bench_benchdatapostgres"."id" = 3 LIMIT 21
  .
  .
  .
  1501. SELECT ....
Mem Usage: 18.23 MB
Exec Time: 1659.42 ms
```

This leads to an N+1 query. look at the total: 1501 queries, and I only have 1500 rows. This happens not just with `.only()`, but with `.defer()` too.

N+1 queries usually happen on parent–child tables. You can solve it using `select_related` or `prefetch_related` (I’ll post about that later). But how do we make sure we only select specific fields and prevent access to fields that weren’t returned?

First, using values_list()

We can change the django query become like this:
```python
lst = BenchDataPostgres.objects.values_list("name", "age", "email")
    collect = []
    for l in lst:
        collect.append( l[0])
        collect.append( l[1])
        collect.append( l[2])
```

When using values_list(), it will return a list of tuples, not a queryset with model instances. That’s the downside, you can't use functions like `.update()` in this case. But the query itself still gets translated into this SQL:
```bash
Query Generated: SELECT "bench_benchdatapostgres"."name", "bench_benchdatapostgres"."age", "bench_benchdatapostgres"."email" FROM "bench_benchdatapostgres"
Mem Usage: 973.77 KB
Exec Time: 158.80 ms
```

So you need to remember how you order the fields, because you have to access the values by index. If you try to access outside the index, it will raise an error.

The second way to restrict fields is by using `.values()`. It's similar to `.values_list()`, but instead of a list of tuples, it returns a list of dicts, more human-friendly, since you don’t need to remember the field order. Look at this example:
```python
lst = BenchDataPostgres.objects.values("name", "age", "email")
collect = []
for l in lst:
    collect.append( l["name"])
    collect.append( l["age"])
    collect.append( l["email"])
```
The generated query become like this:
```bash
Query Generated: SELECT "bench_benchdatapostgres"."name", "bench_benchdatapostgres"."age", "bench_benchdatapostgres"."email" FROM "bench_benchdatapostgres"
Mem Usage: 1.20 MB
Exec Time: 159.93 ms
```
And if you try to access outside returned fields, it will raise error like this

```bash
    KeyError                                  Traceback (most recent call last)
Cell In[2], line 1
----> 1 test_values()

File ~/code/python/wip/bench/service.py:32, in test_values()
     30 collect.append( l["name"])
     31 collect.append( l["age"])
---> 32 collect.append( l["metadata"])
     33 collect.append( l["email"])
```
Yes django orm is powerfull but remember, orm is always put some "magic" for the end user, understand the "magic" will help us to write more efficient query.