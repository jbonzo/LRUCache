package main

import (
	"log"

	"cache/pkg/cache"
)

func main() {
	myCache := cache.NewLRUCache(5)

	expectedItem := &cache.CacheItem{
		Tag: "a",
		Data: 0,
	}

	myCache.AddItem("a", 0)

	item, err := myCache.GetItem("a")
	if err != nil {
		log.Fatalf("Got an error getting item; err: %s", err)
	}

	log.Printf("We expected the item: %v\nWe got the item: %v\n", expectedItem, item)
}
