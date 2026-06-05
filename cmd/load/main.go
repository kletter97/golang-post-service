// load.go
package main

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	targetURL   = "http://localhost:8090/createpost"
	concurrency = 5 // Количество параллельных воркеров (потоков)
)

func main() {
	fmt.Printf("Запуск генератора нагрузки на %s в %d потоков...\n", targetURL, concurrency)
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			client := &http.Client{Timeout: 2 * time.Second}

			// Тестовый JSON (имитируем создание поста)
			jsonData := []byte(`{"author_id": "1", "content": "Спам-тест под нагрузкой!"}`)

			for {
				resp, err := client.Post(targetURL, "application/json", bytes.NewBuffer(jsonData))
				if err != nil {
					fmt.Printf("[Воркер %d] Ошибка запроса: %v\n", workerID, err)
				} else {
					resp.Body.Close()
				}

				// Небольшая пауза между запросами каждого воркера (10-50мс)
				time.Sleep(20 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
}
