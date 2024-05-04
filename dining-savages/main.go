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

const N = 10000

var savagesLeft = 0
var servings = 0
var mutex = sync.Mutex{}
var M = 10
var emptyPot = NewSemaphore(0)
var fullPot = NewSemaphore(0)

func cook() {
	for {
		emptyPot.Wait()
		putServingsInPot()
		fullPot.Signal()
	}
}

func putServingsInPot() {
	fmt.Printf("\n\nCooking %d servings...", M)
	servings = M
	fmt.Printf("\nputting servings in pot\n\n")
}

func savage(id int) {
	mutex.Lock()
	if servings == 0 {
		emptyPot.Signal()
		fullPot.Wait()
		servings = M
	}
	servings -= 1
	getServingFromPot(id)
	savagesLeft++
	mutex.Unlock()

	eat(id)
}

func getServingFromPot(id int) {
	fmt.Printf("Savage %d getting serving from pot\n", id)
}

func eat(id int) {
	fmt.Printf("%d eating\n", id)
}

func main() {
	wg := sync.WaitGroup{}

	go cook()
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			savage(i)
		}(i)
	}

	wg.Wait()
	fmt.Printf("\n\n%d savages have eaten", savagesLeft)
}
