// Generates ten thousand chunks in parallel.

package main

import "github.com/Nightgunner5/stuzzd/networking"
import "sync"

func main() {
	var wg sync.WaitGroup
	for x := int32(-50); x < 50; x++ {
		for z := int32(-50); z < 50; z++ {
			wg.Add(1)
			go func() {
				networking.ChunkGen(x, z)
				wg.Done()
			}()
		}
	}
	wg.Wait()
}
