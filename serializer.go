// Package serializer implement mono-thread
// executor (like mutex but with channels)
package serializer

import (
	log "github.com/Sirupsen/logrus"
)

// Serializer allows you to encapsulate
// in a single thread your calls.
// panics are not propagated and kill
// the innermost goroutine so handle them
// properly.
type Serializer interface {
	Close()
	Serialize(fnc func() interface{}) interface{}
}

type operation struct {
	fnc       func() interface{}
	cResponse chan (interface{})
}

type rOperation chan (operation)
type rKill chan (chan (bool))

type concurrentService struct {
	isAlive    bool
	cOperation rOperation
	cKill      rKill
}

// New creates a new Serializer
// and starts its goroutine so it will be
// ready to accept requests.
func New() Serializer {
	cs := &concurrentService{
		isAlive:    true,
		cOperation: make(rOperation),
		cKill:      make(rKill),
	}

	go func() {
		defer func() {
			// in case of panic, kill the goroutine
			// and signal its state
			if r := recover(); r != nil {
				log.WithFields(log.Fields{
					"recover": r,
				}).Warn("ConcurrentHandler::goroutine panic")
			}
			cs.isAlive = false
		}()

		for { // neverending for
			select {
			case r := <-cs.cKill:

				log.WithFields(log.Fields{
					"kill": r,
				}).Debug("ConcurrentHandler::goroutine kill received")

				cs.isAlive = false
				close(cs.cOperation)
				close(cs.cKill)
				r <- true
				return
			case op := <-cs.cOperation:

				log.WithFields(log.Fields{
					"op": op,
				}).Debug("ConcurrentHandler::goroutine operation received")

				ret := op.fnc()
				op.cResponse <- ret
			}
		}
	}()

	return cs
}

// Serialize executes the fnc passed in the single
// goroutine.
func (cs *concurrentService) Serialize(fnc func() interface{}) interface{} {
	log.WithFields(log.Fields{
		"fnc": fnc,
	}).Debug("ConcurrentHandler::Serialize")

	if !cs.isAlive {
		panic("Attempting to use a closed ConcurrentHandler")
	}

	cResp := make(chan (interface{}))
	cs.cOperation <- operation{fnc: fnc, cResponse: cResp}

	return <-cResp
}

// Close disposes the internal goroutine
// and locks until the message has been received
func (cs *concurrentService) Close() {
	if !cs.isAlive {
		panic("Attempting to use a closed ConcurrentHandler")
	}

	cs.isAlive = false

	cResp := make(chan (bool))

	cs.cKill <- cResp
	<-cResp
}
