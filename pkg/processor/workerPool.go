package processor

import (
	"fmt"
	"sync"
)

func Pool(tasks <-chan Task, goCount int, worker *SendWorker) {
	wg := &sync.WaitGroup{}
	wg.Add(goCount)

	for i := 0; i < goCount; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("recover from panic in pool: %v\n", r)
				}
			}()
			defer wg.Done()
			for taskItem := range tasks {
				fmt.Printf("incoming task %+v\n", taskItem)
				worker.Send(taskItem)
			}
		}()
	}

	fmt.Println("workers started")
	wg.Wait()
}
