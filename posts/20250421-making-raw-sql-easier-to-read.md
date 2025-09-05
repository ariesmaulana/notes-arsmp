title: Making Raw SQL Easier to Read with Row Factory
tag: python, sqlite, postgre
--

One thing that annoy me when using raw query in Python is how the returned value is not "easy to remember". Look at this query:
```python
import sqlite3
con = sqlite3.connect("library.sqlite3")
query = "SELECT id, isbn, title FROM books"
con = con.execute(query)
rows = con.fetchall()
```
To print the returned value I must remember the index of each field:
```python
for r in rows:
	print(f"{r[0]} - {r[1]} - {r[2]}")

# index 0 = id
# index 1 = isbn
# index 2 = title

1 - 1234567890 - Book 1 by Author 1
2 - 2345678901 - Book 1 by Author 1
3 - 3456789012 - Book 1 by Author 1
4 - 4567890123 - Book 1 by Author 2
5 - 5678901234 - Book 1 by Author 2
```

And if you think it's only coming from sqlite, you're wrong. I tried PostgreSQL with psycopg and it has same behavior:
```python
import psycopg2
conn = psycopg2.connect()
cursor = conn.cursor()
query = "SELECT id, isbn, title FROM books"

cursor.execute(query)
rows = cursor.fetchall()
for r in rows:
    print(f"{r[0]} - {r[1]} - {r[2]}")
```
So is using ORM the answer? Luckily, it's not the only answer. After reading the docs, I found sqlite and postgre lib have something called row factory.

### Sqlite3.Row
```python
import sqlite3
con = sqlite3.connect("library.sqlite3")
con.row_factory = sqlite3.Row  # add this

query = "SELECT id, isbn, title FROM books"
con = con.execute(query)
rows = con.fetchall()

for r in rows:
	print(f"{r['id']} - {r['isbn']} - {r['title']}")
```

### Psycopg2.DictCursor
```python
import psycopg2
from psycopg2 import extras

conn = psycopg2.connect()
cursor = conn.cursor(cursor_factory=extras.DictCursor)
query = "SELECT id, isbn, title FROM books"

cursor.execute(query)
rows = cursor.fetchall()
for r in rows:
    print(f"{r['id']} - {r['isbn']} - {r['title']}")
```
Or you can return result like ORM using dataclass (for sqlite or psycopg3) or NamedTuple (for psycopg2):

### sqlite with Dataclass
```python
import sqlite3
from dataclasses import dataclass

@dataclass
class Books:
	id: int
	isbn: str
	title: str

def converter_books(cursor, row):
	fields = [column[0] for column in cursor.description]
	return Books(**{k: v for k, v in zip(fields, row)})

con = sqlite3.connect("library.sqlite3")
con.row_factory = converter_books

query = "SELECT id, isbn, title FROM books"
con = con.execute(query)
rows = con.fetchall()

for r in rows:
	print(f"{r.id} - {r.isbn} - {r.title}")
```

### Psycopg2 With NamedTuple
```python
import psycopg2
from psycopg2 import extras

conn = psycopg2.connect()
cursor = conn.cursor(cursor_factory=extras.NamedTupleCursor)
query = "SELECT id, isbn, title FROM books"

cursor.execute(query)
rows = cursor.fetchall()
for r in rows:
    print(f"{r.id} - {r.isbn} - {r.title}")
```
Or another version of psycopg

### Psycopg3 With Dataclass
```python
import psycopg
from dataclasses import dataclass
from psycopg.rows import class_row

@dataclass
class Books:
    id: int
    isbn: str
    title: str

conn = psycopg.connect()
cursor = conn.cursor(row_factory=class_row(Books))
query = "SELECT id, isbn, title FROM books"

cursor.execute(query)
rows = cursor.fetchall()
for r in rows:
    print(f"{r.id} - {r.isbn} - {r.title}")

```
So which one is better?

I tried small benchmark and found out:

### sqlite
```bash
Default:
11946 function calls in 0.157 seconds

sqlite3.Row:
11946 function calls in 0.165 seconds

Dataclass:
52279 function calls in 0.231 seconds
```

### Psycopg
```bash
Default:
78624 function calls in 0.521 seconds

DictRow:
288642 function calls in 0.578 seconds

Dataclass/class_row:
288983 function calls in 0.569 seconds
```

So for sqlite I prefer sqlite3.Row, and for postgre I think class_row/dataclass is good enough.