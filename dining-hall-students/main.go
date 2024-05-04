package main

import (
	"fmt"
	"sync"
)

type Semaphore struct { // este semáforo implementa quaquer numero de creditos em "v"
	v    int           // valor do semaforo: negativo significa proc bloqueado
	fila chan struct{} // canal para bloquear os processos se v < 0
	sc   chan struct{} // canal para atomicidade das operacoes wait e signal
}

func NewSemaphore(init int) *Semaphore {
	s := &Semaphore{
		v:    init,                   // valor inicial de creditos
		fila: make(chan struct{}),    // canal sincrono para bloquear processos
		sc:   make(chan struct{}, 1), // usaremos este como semaforo para SC, somente 0 ou 1
	}
	return s
}

func (s *Semaphore) Wait() {
	s.sc <- struct{}{} // SC do semaforo feita com canal
	s.v--              // decrementa valor
	if s.v < 0 {       // se negativo era 0 ou menor, tem que bloquear
		<-s.sc               // antes de bloq, libera acesso
		s.fila <- struct{}{} // bloqueia proc
	} else {
		<-s.sc // libera acesso
	}
}

func (s *Semaphore) Signal() {
	s.sc <- struct{}{} // entra sc
	s.v++
	if s.v <= 0 { // tem processo bloqueado ?
		<-s.fila // desbloqueia
	}
	<-s.sc // libera SC para outra op
}

const N = 100000

var studentsLeft = 0
var eating, readyToLeave int
var mutex = sync.Mutex{}
var okToLeave = NewSemaphore(1)

func student(id int) {
	getFood(id)

	mutex.Lock()
	eating++
	if eating == 2 && readyToLeave == 1 {
		okToLeave.Signal()
		readyToLeave--
	}
	mutex.Unlock()

	dine(id)

	mutex.Lock()
	eating--
	readyToLeave++
	if eating == 1 && readyToLeave == 1 {
		mutex.Unlock()
		okToLeave.Wait()
	} else if eating == 0 && readyToLeave == 2 {
		okToLeave.Signal()
		readyToLeave -= 2
		mutex.Unlock()
	} else {
		readyToLeave--
		mutex.Unlock()
	}

	studentsLeft++
	leave(id)
}

func getFood(id int) {
	fmt.Printf("Student %d getting food...\n", id)
}

func dine(id int) {
	fmt.Printf("Student %d Dining!\n", id)
}

func leave(id int) {
	fmt.Printf("FINISH - Student %d Leaving...\n", id)
}

func main() {
	var wg sync.WaitGroup

	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			student(id)
		}(i)
	}

	wg.Wait()
	fmt.Printf("\n%d students left!", studentsLeft)
}
