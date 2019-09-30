# Traverse

As per [habr/469441](https://habr.com/ru/post/469441/).

```bash
$ go get github.com/ernado/traverse
$ traverse --help
./traverse --help
Usage of ./traverse:
  -j int
        concurrent requests (default 8)
  -timeout duration
        timeout, zero means no timeout (default 1m0s)

# Full result with big concurrency:
$ traverse -j 12 --timeout 30ms
1=delectus aut autem
 2=quis ut nam facilis et officia qui
 3=fugiat veniam minus
  4=et porro tempora
  5=laboriosam mollitia et enim quasi adipisci quia provident illum

# Partial result with small concurrency:
$ ./traverse -j 1 --timeout 30ms
Failed: Get http://jsonplaceholder.typicode.com/todos/3: context deadline exceeded
1=delectus aut autem
 2=quis ut nam facilis et officia qui
 3=
  4=
  5=
```