# LRU Cache

This is an implementation of an in memory [LRU Cache!](https://en.wikipedia.org/wiki/Cache_replacement_policies#Least_recently_used_(LRU))

For some example uses see [main.go](./cmd/cache/main.go) for some trivial examples
or the tests [TestCache_GetItem](https://github.com/jbonzo/LRUCache/blob/master/pkg/cache/cache_test.go#L285)
and [TestCache_AddItem](https://github.com/jbonzo/LRUCache/blob/master/pkg/cache/cache_test.go#L215)

This cache uses the [Write Through policy](https://en.wikipedia.org/wiki/Cache_(computing)#Writing_policies). 
 
## Areas of improvements

* There definitely can be more test cases added since caching can get quite complicated
  * The test cases shown show a minimal viable cache and provide coverage for the basic cases
* This cache does `O(n)` retrieval for the LRU. To implement a `O(log(n)` solution refer to the comment
in [this file](https://github.com/jbonzo/LRUCache/blob/master/pkg/cache/cache.go#L149-L152)
* The code is more strict towards the current Write Through policy, so changing to a new policy would
require some refactoring, though not much and I did comment where some of those changes would be like [here](https://github.com/jbonzo/LRUCache/blob/master/pkg/cache/cache.go#L136-L138)
