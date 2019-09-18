package cache

import (
	"os"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
)

var (
	fakeTime  = time.Now().UTC()
	fakeClock = clockwork.NewFakeClockAt(fakeTime)

	time0 = fakeTime
	time1 = time0.Add(time.Second * 1)
	time2 = time0.Add(time.Second * 2)
	time3 = time0.Add(time.Second * 3)
	time4 = time0.Add(time.Second * 4)

	itemA *CacheItem
	itemB *CacheItem
	itemC *CacheItem
	itemD *CacheItem
	itemE *CacheItem
)

func resetItems() {
	itemA = &CacheItem{
		Tag:      "a",
		lastUsed: time0,
	}
	itemB = &CacheItem{
		Tag:      "b",
		lastUsed: time1,
	}
	itemC = &CacheItem{
		Tag:      "c",
		lastUsed: time2,
	}
	itemD = &CacheItem{
		Tag:      "d",
		lastUsed: time3,
	}
	itemE = &CacheItem{
		Tag:      "e",
		lastUsed: time4,
	}
}

func TestMain(m *testing.M) {
	resetItems()
	os.Exit(m.Run())
}

func TestFindLRU(t *testing.T) {
	myCache := NewLRUCacheWithClock(5, fakeClock)

	t.Run("normal_case", func(t *testing.T) {
		resetItems()

		cacheItems := map[string]*CacheItem{
			"a": itemA,
			"b": itemB,
			"c": itemC,
			"d": itemD,
			"e": itemE,
		}

		myCache.items = cacheItems

		lru := myCache.findLRU()
		assert.Equal(t, itemA, lru)
	})

	t.Run("two_share_time", func(t *testing.T) {
		resetItems()

		itemB.lastUsed = time0

		cacheItems := map[string]*CacheItem{
			"a": itemA,
			"b": itemB,
			"c": itemC,
			"d": itemD,
			"e": itemE,
		}
		myCache.items = cacheItems

		lru := myCache.findLRU()

		// it could either equal a or b. Our collision policy is to take either since
		// the case is so unlikely
		correctTime := lru == itemA || lru == itemB
		if !assert.True(t, correctTime) {
			t.Logf("Incorrect item: %v", lru)
		}
	})

	t.Run("different_order", func(t *testing.T) {
		resetItems()
		itemA.lastUsed = time4

		cacheItems := map[string]*CacheItem{
			"d": itemD,
			"a": itemA,
			"e": itemE,
			"b": itemB,
			"c": itemC,
		}
		myCache.items = cacheItems

		lru := myCache.findLRU()
		assert.Equal(t, itemB, lru)
	})
}

// test update cache and evict, for more units make a evict unit test but low priority
func TestUpdateCache(t *testing.T) {
	resetItems()

	myCache := NewLRUCacheWithClock(2, fakeClock)

	t.Run("evicted_case", func(t *testing.T) {
		items := map[string]*CacheItem{
			"a": itemA,
			"b": itemB,
		}

		myCache.items = items

		myCache.updateCache("c", 4)

		expectedItems := map[string]*CacheItem{
			"b": itemB,
			"c": itemC,
		}

		assert.Len(t, expectedItems, 2)

		assert.Contains(t, expectedItems, "b")
		assert.Contains(t, expectedItems, "c")
		assert.NotContains(t, expectedItems, "a")
	})

	t.Run("no_eviction_case", func(t *testing.T) {
		items := map[string]*CacheItem{
			"a": itemA,
		}

		myCache.items = items

		myCache.updateCache("c", 4)

		expectedItems := map[string]*CacheItem{
			"a": itemA,
			"c": itemC,
		}

		assert.Len(t, expectedItems, 2)

		assert.Contains(t, expectedItems, "a")
		assert.Contains(t, expectedItems, "c")
	})

	t.Run("multiple_evictions", func(t *testing.T) {
		items := map[string]*CacheItem{
			"a": itemA,
		}

		myCache.items = items

		myCache.updateCache("c", 4)

		expectedItems := map[string]*CacheItem{
			"a": itemA,
			"c": itemC,
		}

		assert.Len(t, expectedItems, 2)

		assert.Contains(t, expectedItems, "a")
		assert.Contains(t, expectedItems, "c")

		myCache.updateCache("b", 4)

		expectedItems = map[string]*CacheItem{
			"b": itemB,
			"c": itemC,
		}

		assert.Len(t, expectedItems, 2)

		assert.Contains(t, expectedItems, "b")
		assert.Contains(t, expectedItems, "c")
		assert.NotContains(t, expectedItems, "a")

		myCache.updateCache("d", 4)

		expectedItems = map[string]*CacheItem{
			"d": itemD,
			"c": itemC,
		}

		assert.Len(t, expectedItems, 2)

		assert.Contains(t, expectedItems, "d")
		assert.Contains(t, expectedItems, "c")
		assert.NotContains(t, expectedItems, "a")
		assert.NotContains(t, expectedItems, "b")
	})
}

func TestCache_AddItem(t *testing.T) {

}

func TestCache_GetItem(t *testing.T) {
	resetItems()

	cases := []struct {
		name              string
		getItemTag        string
		addToBackingStore bool
		storedData        interface{}
		items             map[string]*CacheItem
		expectedItem      *CacheItem
		expecterErr       error
		expectedUses      int
		expectedHits      int
		expectedMisses    int
	}{
		{
			name:              "hit one item:",
			addToBackingStore: true,
			getItemTag:        "a",
			storedData:        itemA.Data,
			items: map[string]*CacheItem{
				"a": itemA,
			},
			expectedItem: itemA,
			expectedUses: 1,
			expectedHits: 1,
		},
		{
			name:              "miss item but not backing store",
			addToBackingStore: true,
			getItemTag:        "a",
			storedData:        itemA.Data,
			items:             map[string]*CacheItem{},
			expectedItem:      itemA,
			expectedUses:      1,
			expectedMisses:    1,
		},
		{
			name:              "miss item and not backing store",
			addToBackingStore: false,
			getItemTag:        "a",
			storedData:        itemA.Data,
			items:             map[string]*CacheItem{},
			expecterErr:       ErrDataNotInBackingStore,
			expectedItem:      itemA,
			expectedUses:      1,
			expectedMisses:    1,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			myCache := NewLRUCacheWithClock(5, fakeClock)
			myCache.items = c.items

			if c.addToBackingStore {
				myCache.backingStore.addItem(c.getItemTag, c.storedData)
			}

			item, err := myCache.GetItem(c.getItemTag)
			if c.expecterErr != nil {
				assert.Error(t, err)
				assert.Equal(t, c.expecterErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.expectedItem, item)
			}

			assert.Equal(t, c.expectedUses, myCache.uses)
			assert.Equal(t, c.expectedHits, myCache.hits)
			assert.Equal(t, c.expectedMisses, myCache.misses)
		})
	}
}
