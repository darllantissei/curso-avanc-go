package main

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

func init() {
	// Definir a quantidade de CPU que pode usar

	// Definir apenas 1 CPU
	// runtime.GOMAXPROCS(1)

	// Irá forçar a execução na quantidade máxima de CPU's, mas a partir da versão 1.5 do go não é mais necessário essa definição
	// runtime.GOMAXPROCS(runtime.NumCPU())
}

var result int

var mtx sync.Mutex

func main() {
	// Processo simultâneos
	// simultaneousProcess()

	// -----------------------------------------

	// Processos concorrentes - wait-groups
	// concurrentProcess()

	// -----------------------------------------

	// Prevenção de race condition com Mutex
	// useMutexToPreventRaceCondition()

	// -----------------------------------------

	// Channel
	// useChannels()

	// -----------------------------------------

	// Channel with waitgroup
	useChannelsWithWaitGroup()

}

func useChannelsWithWaitGroup() {

	msgs := make(chan string)

	wtg := sync.WaitGroup{}

	wtg.Add(9)

	go runProcess("P1", 20, &wtg, &msgs)

	go runProcess("P2", 30, &wtg, &msgs)

	go runProcess("P3", 40, &wtg, &msgs)

	go runProcess("P4", 50, &wtg, &msgs)

	go runProcess("P5", 60, &wtg, &msgs)

	go runProcess("P6", 70, &wtg, &msgs)

	go runProcess("P7", 80, &wtg, &msgs)

	go runProcess("P8", 90, &wtg, &msgs)

	go runProcess("P9", 100, &wtg, &msgs)

	go func(wait sync.WaitGroup, chn chan string) {

		wait.Wait()
		close(chn)

	}(wtg, msgs)

	for text := range msgs {

		if strings.EqualFold(text, "") {
			continue
		}

		fmt.Println(text)
	}

	fmt.Println("Final result: ", result)

}

func useChannels() {

	// make consegue, trabalhar apenas com 3 tipos: maps, slices e chan(channel)

	msg := make(chan string)

	go runProcess("P1", 20, nil, &msg)

	for {

		text := <-msg

		if strings.EqualFold(text, "") {
			break
		}

		fmt.Println(text)
	}

	close(msg)

}

func useMutexToPreventRaceCondition() {

	// Verificar race condition: go run -race main.go

	go runProcess("P1", 20, nil, nil)
	go runProcess("P2", 20, nil, nil)

	var s string

	// Obter os retorno da Goroutines
	fmt.Scanln(&s)

	fmt.Println("Final result: ", result)
}

func concurrentProcess() {
	var waitGroup sync.WaitGroup

	waitGroup.Add(2)

	go runProcess("P1", 20, &waitGroup, nil)
	go runProcess("P2", 20, &waitGroup, nil)

	waitGroup.Wait()
}

func simultaneousProcess() {
	go runProcess("P1", 20, nil, nil)
	go runProcess("P2", 20, nil, nil)

	var s string

	// Obter os retorno da Goroutines
	fmt.Scanln(&s)
}

func runProcess(name string, total int, waitGroup *sync.WaitGroup, chn *chan string) {

	for i := 1; i <= total; i++ {

		t := time.Duration(rand.Intn(255))

		time.Sleep(time.Millisecond * t)

		mtx.Lock() // Prevenir race condition - bloquear alteração fora de fluxo
		result++

		text := fmt.Sprintf("Processo: %10s -> %04d -- %30s -- Partial result: %04d",
			name,
			i,
			time.Now().Format("2006-01-02 15:04:05.999999999"),
			result)

		if chn == nil {
			fmt.Println(text)
		} else {
			*chn <- text
		}

		mtx.Unlock() // Prevenir race condition - desbloquear alteração para que o fluxo flua normalmente
	}

	if chn != nil {
		*chn <- ""
	}

	if waitGroup != nil {
		waitGroup.Done()
	}
}
