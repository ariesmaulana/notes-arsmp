title: Til Postgre Json Jsonb
tag: sql, postgre, json, jsonb
--

Several days ago, I heard a joke:
> You don't need Mongo, PostgreSQL is enough.

So... let's try it!

## Scopes

Before we move forward, let’s set the scope of this test.
In my experience, the use cases where I used Mongo are:

- Store raw information coming from request/response, e.g. webhook, API call, etc.

- Store personalized config with JSON structure.

So we will try to use these cases using json and jsonb column in PostgreSQL.

## JSON and JSONB

There are two types for storing JSON format in PostgreSQL: json and jsonb.

The difference based on my understanding:

- JSON: Store the JSON as is, no normalization, no restructuring. Just store it like plain text.

- JSONB: Store the JSON after normalize process — like remove whitespaces, restructure keys, and store it as binary format.

So i think:

- JSON is a good case for audit log or storing raw data.

- JSONB is a perfect match for personal config.

### Playing Around with JSON and JSONB

First we will playing aroung at json column, i have several dummy records like this

```bash
user_id |                              logs
---------+----------------------------------------------------------------
     102 | {                                                             +
         |   "after": {"email": "alice.b@email.com", "name": "Alice B."},+
         |   "before": {"name": "Alice", "email": "alice@email.com"}     +
         | }
      15 | {                                                             +
         |   "before": {"address": "No. 123"},                           +
         |   "after": {"address": "No. 123B"}                            +
         | }
```
We will just playing around to handle searching with value of the json column..

```bash
example=# select user_id, logs -> 'before' ->> 'email' as email_before, logs -> 'after' ->> 'email' as after_email from user_logs;
 user_id |   email_before   |    after_email
---------+------------------+-------------------
       1 | john@old.com     | john@new.com
       2 | alice@mail.com   | alice@mail.com
       3 |                  |
       4 |                  |
       5 | charlie@mail.com | charlie@mail.com
       6 |                  |
       7 | eve@mail.com     | eve.new@mail.com
       8 |                  |
       9 | grace@mail.com   | grace@newmail.com
      10 | hana@mail.com    | hana@mail.com
      11 | ivan@mail.com    | ivan@updated.com
      12 | judy@mail.com    | judy123@mail.com
      13 |                  |
      14 | laura@mail.com   | laura@updated.com
      15 |                  |
     102 | alice@email.com  | alice.b@email.com
(16 rows)
```

For this kind of example i used a lot operator '->' and '->>'.
'->' This operator read value as json object
'->>' This operator convert the value as string so we can filter as normal field.
Let try with filter
```bash
select user_id, logs -> 'before' ->> 'email' as email_before,  logs -> 'after' ->> 'email' as after_email  from user_logs 
where logs -> 'after' ->> 'email' = 'john@new.com';

 user_id | email_before | after_email
---------+--------------+--------------
       1 | john@old.com | john@new.com
(1 row)
```


JSONB

Like I said before, in JSONB, before the value is stored in the column, it goes through some kind of normalization.
```bash
-- insert
INSERT INTO user_config  
            (user_id,  
             config)  
VALUES      (5,  
             '{ "theme": "dark", "language": "id", "status": "free_user" }');
-- result

user_id |                             config
---------+-----------------------------------------------------------------
       1 | {"theme": "dark", "status": "premium_user", "language": "jp"}
       2 | {"theme": "dark", "status": "premium_user", "language": "en"}
       3 | {"theme": "dark", "status": "admin", "language": "id"}
       4 | {"theme": "dark", "status": "premium_user", "language": "id"}
       5 | {"theme": "solarized", "status": "free_user", "language": "id"}
```

See? Data on table little bit different with the query right?

Let’s play around with filters.
Because in JSONB the data is stored in binary format, we can use the contains operator @>:
```bash
select * from user_config where config @> '{"theme":"dark"}' limit 5;

 user_id |                            config
---------+---------------------------------------------------------------
       1 | {"theme": "dark", "status": "premium_user", "language": "jp"}
       2 | {"theme": "dark", "status": "premium_user", "language": "en"}
       3 | {"theme": "dark", "status": "admin", "language": "id"}
       4 | {"theme": "dark", "status": "premium_user", "language": "id"}
       7 | {"theme": "dark", "status": "free_user", "language": "id"}
```

Or the exist operator ?:
```bash
example=# select * from user_config where config ? 'notification' limit 5;
 user_id | config
---------+--------
(0 rows)

example=# select * from user_config where config ? 'language' limit 5;
 user_id |                             config
---------+-----------------------------------------------------------------
       1 | {"theme": "dark", "status": "premium_user", "language": "jp"}
       2 | {"theme": "dark", "status": "premium_user", "language": "en"}
       3 | {"theme": "dark", "status": "admin", "language": "id"}
       4 | {"theme": "dark", "status": "premium_user", "language": "id"}
       5 | {"theme": "solarized", "status": "free_user", "language": "id"}
(5 rows)
```

But that doesn’t mean we can’t use the previous operators (->, ->>).

And the good part about JSONB: the column can be indexed. I have 100k records. Before I put the index:
```bash
explain analyze select user_id, config  from user_config where config @> '{"theme":"dark"}';

Seq Scan on user_config  
(cost=0.00..2574.50 rows=35199 width=68)  
(actual time=0.027..46.364 rows=34805 loops=1)
Filter: (config @> '{"theme": "dark"}'::jsonb)
Rows Removed by Filter: 69395
Execution Time: 48.419 ms

```

Then I added the index:
```bash
CREATE INDEX idx_user_config_config_jsonb ON user_config USING gin (config);
```

The result is good:
```bash
explain analyze select user_id, config  from user_config where config @> '{"theme":"dark"}';
Bitmap Heap Scan on user_config  
(cost=263.71..1975.70 rows=35199 width=68)  
(actual time=12.631..32.547 rows=34805 loops=1)
   Recheck Cond: (config @> '{"theme": "dark"}'::jsonb)
   Heap Blocks: exact=1272
   ->  Bitmap Index Scan on idx_user_config_config_jsonb  
      (cost=0.00..254.91 rows=35199 width=0)  
      (actual time=12.393..12.394 rows=34805 loops=1)
Execution Time: 34.529 ms
```

There are many things I still need to learn about this column, but for now, it's enough for me to say:
I think for my use case, I don't need MongoDB.