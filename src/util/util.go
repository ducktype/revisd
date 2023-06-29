package util

import "regexp"


import "fmt"
import "runtime"
import "runtime/debug"
//import "sync"
import "sync/atomic"

//----------------------------------------------------------------------------------------

func ParseMap(cfg_data string, default_key string) map[string][]string {
	var cfg = make(map[string][]string)
	//parse configuration data
	var rx = regexp.MustCompile(`(?ims)^\s*(?P<rem>#)?\s*(?P<k>[^=$]*?)(?:\s*=\s*(?P<v>[^$]*?)\s*)?$`)
	var matches = rx.FindAllStringSubmatch(cfg_data,-1)
	for _, match := range matches {
		var comment = match[1]
		var k = match[2]
		var v = match[3]
		if len(comment) > 0 { continue }
		if len(v) == 0 && len(k) == 0 { continue }
		if len(v) == 0 && len(k) > 0 {
			v = k
			k = default_key
		}
		//fmt.Printf("k: %q v: %q\n", k,v)
		cfg[k] = append(cfg[k],v)
	}
	return cfg
}

//----------------------------------------------------------------------------------------
//https://github.com/sourcegraph/conc
//https://stackoverflow.com/questions/18207772/how-to-wait-for-all-goroutines-to-finish-without-using-time-sleep
//https://gobyexample.com/atomic-counters
//https://medium.com/@deckarep/the-go-1-19-atomic-wrappers-and-why-to-use-them-ae14c1177ad8
//https://vorpus.org/blog/notes-on-structured-concurrency-or-go-statement-considered-harmful/
//https://stackoverflow.com/questions/29300607/golang-bug-or-intended-feature-on-map-literals

type Panic struct {
	Value 	any
	Callers []uintptr
	Stack 	[]byte
}

func (p *Panic) String() string {
	return fmt.Sprintf("panic: %v\nstacktrace:\n%s\n", p.Value, p.Stack)
}

type Gog struct {
	start_num	atomic.Uint64
  end_ch    chan *Panic
}

func NewGog() *Gog {
  return &Gog{
    end_ch:	make(chan *Panic),
  }
}

func (p *Gog) Go(f func()) {
	p.start_num.Add(1)
	go func() {
		defer func(){
			//p.wg.Done()
			var panic_val = recover()
			var panic *Panic
			if panic_val != nil {
				var callers [64]uintptr
				n := runtime.Callers(1, callers[:])
				panic = &Panic{
					Value:   panic_val,
					Callers: callers[:n],
					Stack:   debug.Stack(),
				}
			}
			p.end_ch <- panic
		}()
		f()
	}()
}

func (p *Gog) Wait(bpanic bool) []*Panic {
	var nend uint64
	panics := make([]*Panic,1)
	loop := true
	for loop {
		//fmt.Printf("for")
    select {
	    case pa := <-p.end_ch:
	    	//fmt.Printf("recv endch")
	      if pa != nil {
	      	if bpanic { panic(*pa) }
	      	panics = append(panics,pa)
	      }
	      nend += 1
	      nstart := p.start_num.Load()
	      //fmt.Printf("%v == %v",nend,nstart)
	      if nend == nstart {
	      	loop = false
	      	break
	      }
  	}
  }
  close(p.end_ch)
  return panics
}


//----------------------------------------------------------------------------------------
//https://stackoverflow.com/questions/36417199/how-to-broadcast-message-using-channel

type Broker[T any] struct {
  stopCh    chan struct{}
  publishCh chan T
  subCh     chan chan T
  unsubCh   chan chan T
}

func NewBroker[T any]() *Broker[T] {
  return &Broker[T]{
    stopCh:    make(chan struct{}),
    publishCh: make(chan T, 1),
    subCh:     make(chan chan T, 1),
    unsubCh:   make(chan chan T, 1),
  }
}

func (b *Broker[T]) Start() {
  subs := map[chan T]struct{}{}
  for {
    select {
    case <-b.stopCh:
    	for msgCh := range subs {
        close(msgCh)
    	}
      return
    case msgCh := <-b.subCh:
      subs[msgCh] = struct{}{}
    case msgCh := <-b.unsubCh:
      delete(subs, msgCh)
    case msg := <-b.publishCh:
      for msgCh := range subs {
        // msgCh is buffered, use non-blocking send to protect the broker:
        select {
	        case msgCh <- msg:
	        default:
        }
      }
    }
  }
}

func (b *Broker[T]) Stop() {
  close(b.stopCh)
}

func (b *Broker[T]) Sub() chan T {
  msgCh := make(chan T, 5)
  b.subCh <- msgCh
  return msgCh
}

func (b *Broker[T]) Unsub(msgCh chan T) {
  b.unsubCh <- msgCh
  close(msgCh)
}

func (b *Broker[T]) Pub(msg T) {
  b.publishCh <- msg
}

//----------------------------------------------------------------------------------------