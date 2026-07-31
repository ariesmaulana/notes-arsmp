title: Crud Engineer: How Django ORM Works
tag: tags django,python,orm,database
--

In this post I am trying to describe how the workflow of Django ORM is until the data is saved in the *database*. This writing aims to try to understand the workflow of Django ORM and avoid surprises from the ORM itself. If you think "then just don't use ORM, use SQL directly" I completely agree, however, the context of this writing is again so that we can write ORM a little better.

## Auto Commit

Before discussing further I will show a Django ORM code snippet and whether you can find a problem in that code:
```python
def create_user(obj: User) -> User:
    user = User.objects.create(**obj)
    profile = Profile.objects.create(user=user)
    return user
```
Does anyone know what the possible problem is?

1

2

3

If you can see it, you already know about *auto commit*, for those who don't, I will try to explain what the problem is in the code above. The answer depends on the feature specification first, with the assumption of the code above that *User* will have a *Profile* when registering, so the code above still leaves one problem. The code above if we translate it to SQL will look like this:
```sql
INSERT INTO users (username, email, password) VALUES (?, ?, ?) RETURNING id; ---> commit -> saved to DB
INSERT INTO profiles (user_id) VALUES (?); ---> commit -> saved to DB
```
In the translation example above, it is visible that every SQL row directly performs a *commit* to the *database*, without exception. This condition becomes problematic if, for example, an error occurs on the second row, the first row is already saved, the second row fails causing the data to become invalid because it allows a *User* to be created without a *Profile*.

```sql
INSERT INTO users (username, email, password) VALUES (?, ?, ?) RETURNING id; ---> commit -> saved to DB
INSERT INTO profiles (user_id) VALUES (?); ---> error -> rollback
```

So what is the solution? If it is still in the context of SQL, the easiest way is to wrap it so that the code runs in the same transaction
```sql
BEGIN TRANSACTION;
INSERT INTO users (username, email, password) VALUES (?, ?, ?) RETURNING id;
INSERT INTO profiles (user_id) VALUES (?);
COMMIT/ROLLBACK;
```
With the code like the latest version, this condition produces either everything succeeds or everything fails, making that feature valid that a *User* created will have a *Profile*

But that is in SQL, what about in Django? Because in Django it is initially *auto commit* there are several ways to ensure a function is wrapped in the same transaction.

## ATOMIC_REQUEST

In our *database* config we add `ATOMIC_REQUESTS = True` on `settings.py` to ensure every *request* runs in the same transaction. So if the code above is made into a *view* and has been set to *True* then the code will run in the same transaction.

```python
# settings.py
DATABASES = {
    'default': {
        'ENGINE': 'django.db.backends.postgresql',
        'NAME': 'mydb',
        'ATOMIC_REQUESTS': True,   # <- all views wrapped in 1 transaction
    }
}
def create_user(request) -> User:
    user = User.objects.create(username=request.POST['username'], email=...)
    profile = Profile.objects.create(user=user)
    return user
```

Roughly it will be translated to like this

```sql
BEGIN TRANSACTION;
INSERT INTO users (username, email, password) VALUES (?, ?, ?) RETURNING id;
INSERT INTO profiles (user_id) VALUES (?);
COMMIT/ROLLBACK;
```

## Manual With Context Manager & Decorator

If we want to be more explicit we can use a *context manager* like this:
```python
from django.db import transaction

def create_user(request) -> User:
    with transaction.atomic():
        user = User.objects.create(username=request.POST['username'], email=...)
        profile = Profile.objects.create(user=user)
        return user
```

Or in the form of a decorator:
```python
from django.db import transaction

@transaction.atomic
def create_user(request) -> User:
    user = User.objects.create(username=request.POST['username'], email=...)
    profile = Profile.objects.create(user=user)
    return user
```

Both of these codes will be translated in the same transaction.

## Save Point

As the name suggests, *save point* functions to save a transaction point that can be used later. Imagine you are playing Football Manager using your favorite team, the next match is a crucial derby, you save the current *progress* first because if you lose the *derby* you can restart it. Something like that.

In the context of a feature, you can imagine you are building a multi-step feature like, for example, registering a new student which is split into several stages:
First stage: Fill in personal data
Second stage: Choose Major
Third stage: Choose Payment Option

Assuming all these stages are processed together and not sent separately in one *request*, how do we make sure if a problem occurs we can manage which part we save and which part gets *rollback*? This is the function of *savepoint*.

Let's assume we have a function for the feature above like the following:
```python
def step_one_profile():
    with transaction.atomic():
        user = User.objects.create(**obj)
        profile = Profile.objects.create(user=user)       

def step_two_academic():
    with transaction.atomic():
        academic = Academic.objects.create(**obj)


def step_three_payment():
    with transaction.atomic():
        payment = Payment.objects.create(**obj)

def register():
    step_one_profile()
    step_two_academic()
    step_three_payment()
```

In the example above, the code will not automatically create a *savepoint* because even though each stage is wrapped in a *transaction* they still stand alone instead of being wrapped in one *transaction*. The way is easy enough, we wrap it back into one *transaction*.
```python
def register():
    with transaction.atomic():
        step_one_profile()
        step_two_academic()
        step_three_payment()
```
The code above roughly will be translated to SQL as follows
```sql
BEGIN;

SAVEPOINT sp1;
  INSERT INTO users (...) VALUES (...);   -- step_one
RELEASE SAVEPOINT sp1;  
SAVEPOINT sp2;
  INSERT INTO academic (...) VALUES (...);  -- step_two, 
RELEASE SAVEPOINT sp2;   
SAVEPOINT sp3;
  INSERT INTO payment (...) VALUES (...);  -- step_three, 
RELEASE SAVEPOINT sp3;   

COMMIT;  
```

> To be honest, I am not a fan of *save point* and its complexity, but it is still worth knowing if you want to use Django ORM.

In order to be able to manage *savepoint*, we must be able to manage *error*.

## Managing *Error*

If you notice, the code above does not contain explicit *rollback* information, that is because inside a *transaction* block if an *exception* occurs it will signal that the *rollback* function will be called automatically. This is where the complexity lies because if the code is like above and all stages are only wrapped in a *transaction* and not managed properly, all stages will automatically end in *rollback*.

From the code above, here is what will happen if an *error* occurs
```bash
outer transaction (register)
  ├─ step_one() → success → savepoint 1 (release, no exception)
  ├─ step_two()
  │     └─ academic [savepoint 2]
  │           └─ exception raised
  │           └─ atomic exit detects exception
  │           └─ exception re-raised outwards
  ├─ step_three() → never reaches
  ├─ ROLLBACK, nothing is saved
```
But if the *exception* is managed before the outer *transaction*, there is a possibility that other data will be saved. For example, let's change `step_two` like this
```python
def step_two_academic():
    with transaction.atomic():
        try:
            academic = Academic.objects.create(**obj)
        except:
            // handle exception, silently pass
```
If this happens the flow becomes like this
```bash
outer transaction (register)
  ├─ step_one() → success → savepoint 1 (release, no exception)
  ├─ step_two()
  │     └─ academic [savepoint 2]
  │           └─ exception raised
  │           └─ exception dampened inside block (try/except catches before atomic exit)
  │           └─ atomic() exit NORMAL → RELEASE SAVEPOINT 2
  ├─ step_three() → success → savepoint 3 (release, no exception)
  ├─ COMMIT, data from step_one() and step_three() saved
```
Because the outer *transaction* only detects whether an *exception* occurs or not, if the *exception* that occurs inside each function is not "brought" out of the outer *transaction*, there will never be a "command" to run *rollback*. The easiest way besides letting the *exception* rise to the outermost level is to signal it directly with the command `transaction.set_rollback(True)`. With that code, it gives a command to run *rollback* as soon as possible.


Additional notes:
1. All SQL above is not a literal translation, Django does not provide an explicit BEGIN..... translation. The SQL above is only an illustration of the execution flow.
2. Not all databases support savepoint, please re-read the documentation for each database.

References:
[https://docs.djangoproject.com/en/6.0/topics/db/transactions/](https://docs.djangoproject.com/en/6.0/topics/db/transactions/)

[https://docs.djangoproject.com/en/6.0/topics/db/transactions/#autocommit](https://docs.djangoproject.com/en/6.0/topics/db/transactions/#autocommit)

[https://docs.djangoproject.com/en/6.0/ref/databases/](https://docs.djangoproject.com/en/6.0/ref/databases/)
