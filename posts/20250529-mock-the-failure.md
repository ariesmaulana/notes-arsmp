title: Mock the Failure
tags: python, django, test, pytest


I'm not a big fan of mocking, especially when working with databases. For me, when we create tests involving a database, we can't just mock the result because I want to make sure my code actually stores the correct data. I think it is important to verify that the data is stored correctly during testing. Even if we have a nice algorithm in our code, if the stored data is incorrect, it becomes useless.

But that doesn't mean I'm not using mock at all. Sometimes I do use mock, for example, in this case, to mock a failure.

Lets look at this snippet.
```python
# app/service.py
def insert_retry(data) -> int:
    count = 1
    while count <= 5:
        try:
            Play.objects.create(**data)
            break
        except Exception as e:
            print(e)
            count += 1
    return count
```

It already uses the ORM, so I think we don't need to test simple CRUD operations. But for this snippet, we still need to check the failure case, does it still retry up to 5 times if something goes wrong? To test that behavior, I used mock.

```python
class InsertRetryTest(TestCase):

    @patch('app.service.Play.objects.create')
    def test_retry_always_fail(self, mock_create):

        before = Play.objects.count()
        mock_create.side_effect = Exception("Always Error")

        data = {
            "title": "test",
            "description": "desc"
        }
        result = insert_retry(data)

        after = Play.objects.count()

        # count start from 1 so will stop at 6
        self.assertEqual(result, 6)
        # make sure we call the ORM 5 times
        self.assertEqual(mock_create.call_count, 5)
        # before and after equal, no insert
        self.assertEqual(after, before)
```
Then run the test. By the way, this example I uses Django.

```bash
./manage.py test

Creating test database for alias 'default'...
System check identified no issues (0 silenced).
.
----------------------------------------------------------------------
Ran 1 tests in 0.001s

OK
```


And because we used mock for the failure, we can try another scenario, let's update our snippet.

```python
def insert_retry(data) -> int:
    count = 1
    while count <= 5:
        try:
            Play.objects.create(**data)
            break
        except (ValueError, DatabaseError):
            # no need to retry
            return count
        except Exception as e:
            count += 1
    return count
```

In this updated snippet, we separate the types of failures. Sometimes, we return early, while some errors still need to be retried. So we can update the test like this.

```python
class InsertRetryTest(TestCase):

    @patch('app.service.Play.objects.create')
    def test_retry_always_fail(self, mock_create):

        before = Play.objects.count()
        mock_create.side_effect = Exception("Always Error")

        data = {
            "title": "test",
            "description": "desc"
        }
        result = insert_retry(data)

        after = Play.objects.count()

        # count start from 1 so will stop at 6
        self.assertEqual(result, 6)
        # make sure we call the ORM 5 times
        self.assertEqual(mock_create.call_count, 5)
        # before and after equal, no insert
        self.assertEqual(after, before)

    @patch('app.service.Play.objects.create')
    def test_no_retry_fail(self, mock_create):
        before = Play.objects.count()
        mock_create.side_effect = [ValueError, DatabaseError]

        data = {
            "title": "test",
            "description": "desc"
        }
        result = insert_retry(data)

        after = Play.objects.count()

        # count start from 1 so will stop at 1
        self.assertEqual(result, 1)
        # make sure we call the ORM 1 time
        self.assertEqual(mock_create.call_count, 1)
        # before and after equal, no insert
        self.assertEqual(after, before)
```

And Run again the test
```bash
Creating test database for alias 'default'...
System check identified no issues (0 silenced).
..
----------------------------------------------------------------------
Ran 2 tests in 0.001s

OK
```