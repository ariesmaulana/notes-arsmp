title: The Reason I Stay Away From Generic Cbv
tag: python, django
---

When I write code in Django for the first time, I feel mesmerized by class-based views (CBV), especially the generic views. Look at this CBV example using Django REST Framework:

```python
# serializer.py
class PostSerializer(serializers.ModelSerializer):
    class Meta:
        model = Post
        fields = '__all__'
        
  # views
  class PostListCreate(ListCreateAPIView):
     serializer_class = PostSerializer
     queryset = Post.objects.all()
     
 class PostEdit(RetrieveUpdateDestroyAPIView):
    serializer_class = PostSerializer
    queryset = Post.objects.all()
```

With this simple code, I already built full CRUD functionality. This code already supports:

  - List all "Post"

  - Insert new "Post"

  - Detail of the "Post"

  - Update the "Post"

  - Delete the "Post"

  - Validation based on the Post model

Feels like magic, right? Sometimes when I use Generic CBV, it feels like just adding configuration rather than writing real code.
But for now, I prefer Function-Based Views (FBV).

There’s nothing wrong with CBV, but the generic ones feel too magic for me. In my opinion, it can be hard to collaborate with other team member especially if someone is new to Django.

With generic CBV, it's easy to create a new feature, but it gradually becomes hard to update that feature. Look again at that example, now imagine we need to add more functionality:

  - Add pagination

  - Add custom validation

  - Add authorization

  - Add a custom response

So... where do you put those things?

In the example above, i didn’t write much code, i just passed some params to the existing class. But now where should I put pagination? In model? View? Serializer?
Where should I put validation? Model? View? Serializer?

If you’re experienced with Django CBV, that might sound easy.
But for a new to django it’s really confusing.

That’s why I now prefer FBV.
```python
@api_view(['GET', 'POST'])
def list_and_create(request):
    if request.method == 'POST':
       # Validasi sebelum diproses 
       # Persiapkan data 
       # Insert ke database 
       # return 201 jika sukses
    else:
        # get data dari database
        # return 200 jika ada, return 200 tetap dilakukan jika data 0

@api_view(['GET', 'PUT', 'DELETE'])
def detail_update_delete(request, pk):
    if request.method == 'GET':
        # validation
        # get data
        # return 200 if success
        # return 404 if not found
    else if request.method == 'PUT':
        # validation
        # actual process
        # return
    else:
        # validation
        # get data
        # return 200 if success
        # return 404 if not found
```

Yes, it’s more verbose.
But when we want to add new functionality, we can do it easily, because all the code is there.

Maybe it takes more time when creating a new feature, but it’s much easier to add or update functionality later.

And if you still want to use CBV, I recommend using APIView instead of the generic views.

```python
class PostView(APIView):

    def get(self, request):
        pass
    def post(self, request):
        pass
    def put(self, request, pk):
        pass
    def delete(self, request, pk):
        pass
```