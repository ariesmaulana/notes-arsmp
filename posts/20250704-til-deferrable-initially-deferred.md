title: Til Deferrable Initially Deferred
tag: sql, database, postgre
--

Today at work I learn something new, I face some unexpected behaviour when test the CRUD functionality, it's not bug but it's clearly how stupid I am.

Let say we have table structure like this:

```bash
Parent  
Id  
Name  

Child  
Id  
parent_id  
Name  
```

So I create two test, number 1 test happy path
```bash
-- parent already exist with id = 1
INSERT INTO child (parent_id, name)  
VALUES (1, 'childs');  
```
Expected no error and pass.

Second I tried to insert not existent parent into child
```bash
INSERT INTO child (parent_id, name)  
VALUES (999, 'childs');  
```
Expected error but somehow is no error at all.

Why?

But it raises an error when i commit the transaction. Why?

Turn out the foreign key from `parent_id` is used constraint `DEFERRABLE INITIALLY DEFERRED`, that as far as I understand, the constraint not immediate but processed until commit.

Yeah today I learn.

> note for my self, in postgre have different constraint not only that, I think it worth to read and make as blog post