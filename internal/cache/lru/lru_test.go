// lru implements an LRU cache.
package lru

import (
	"fmt"
	"time"
)

func main() {
	cache := New(3, 5*time.Second)
	defer cache.StopCleanup()

	fmt.Println("Setting initial entries...")
	cache.Set("itemcode1", nil, 20*time.Second)
	cache.Set("itemcode2", nil, 10*time.Second)
	cache.Set("itemcode3", nil, 0)

	fmt.Printf("Cache length: %d\n", cache.Len())

	if val, found := cache.Get("itemcode1"); found {
		fmt.Printf("Get itemcode1: %s\n", val)
	}

	fmt.Println("\nSetting itemcode4, expecting itemcode2 to be evicted (LRU)...")
	cache.Set("itemcode4", nil, 10*time.Second)
	fmt.Printf("Cache length: %d\n", cache.Len())

	if _, found := cache.Get("itemcode2"); !found {
		fmt.Println("itemcode2 correctly evicted or expired.")
	}

	fmt.Println("\nWaiting for itemcode4 TTL 10s and itemcode1 TTL 20s to potentially expire...")
	fmt.Println("itemcode3 should not expire.")

	time.Sleep(12 * time.Second)

	if val, found := cache.Get("itemcode1"); found {
		fmt.Printf("Get itemcode1 after 12s: %s (Still alive)\n", val)
	} else {
		fmt.Println("itemcode1 expired or evicted.")
	}

	if _, found := cache.Get("itemcode4"); !found {
		fmt.Println("itemcode4 correctly expired.")
	} else {
		fmt.Printf("itemcode4 is still present (should have expired or been cleaned up)\n")
	}

	if val, found := cache.Get("itemcode3"); found {
		fmt.Printf("Get itemcode3 after 12s: %s (Should always be present)\n", val)
	} else {
		fmt.Println("itemcode3 (no-expire) was unexpectedly removed.")
	}

	fmt.Printf("Cache length after 12s: %d\n", cache.Len())

	fmt.Println("\nWaiting for remaining items to expire (itemcode1)...")
	time.Sleep(10 * time.Second) // Tổng cộng 22 giây, itemcode1 (20s TTL) sẽ hết hạn

	if _, found := cache.Get("itemcode1"); !found {
		fmt.Println("itemcode1 correctly expired after ~22s.")
	} else {
		fmt.Printf("itemcode1 is still present (should have expired)\n")
	}
	fmt.Printf("Cache length after ~22s: %d\n", cache.Len())
}
