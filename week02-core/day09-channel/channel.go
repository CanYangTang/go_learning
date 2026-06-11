package channel

import "time"

func SendMessage(message string) string {
	messages := make(chan string)
	go func() {
		messages <- message
	}()
	return <-messages
}

func BufferedMessages(messages []string) []string {
	messagesChan := make(chan string, len(messages))
	go func() {
		for _, message := range messages {
			messagesChan <- message
		}
	}()
	result := make([]string, len(messages))
	for i := range messages {
		result[i] = <-messagesChan
	}
	return result
}

func GenerateNumbers(n int) []int {
	numbers := make(chan int, n)
	go func() {
		for i := 1; i <= n; i++ {
			numbers <- i
		}
		close(numbers)
	}()
	var result []int
	for num := range numbers {
		result = append(result, num)
	}
	return result
}

func ReceiveWithTimeout(message string, delay time.Duration, timeout time.Duration) (string, bool) {
	messages := make(chan string, 1)
	go func() {
		time.Sleep(delay)
		messages <- message
	}()
	select {
	case msg := <-messages:
		return msg, true
	case <-time.After(timeout):
		return "", false
	}
}
