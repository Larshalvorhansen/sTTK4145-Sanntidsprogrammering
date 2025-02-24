package distributor

/*
Distributøren håndterer og distribuerer heis ordere
Den oppdaterer states, prosesserer ordere og passer på at ting er syronisert
Tror jeg...
*/

import (
	"fmt"
	"root/config"
	"root/elevator"
	"root/elevio"
	"root/network/peers"
	"time"
)

type StashType int

// Type som definerer none som 0, add som 1, remove som 2 og state som 3.
const (
	None StashType = iota
	Add
	Remove
	State
)

// Tar inn som funksjonsargumenter: sender inn f.eksempel CommonState som en input i networkRx.
// <-chan betyr recieve-only, chan<-betyr send-only
func Distributor(
	confirmedCsC chan<- CommonState,
	deliveredOrderC <-chan elevio.ButtonEvent,
	newStateC <-chan elevator.State,
	networkTx chan<- CommonState,
	networkRx <-chan CommonState,
	peersC <-chan peers.PeerUpdate,
	id int,
) {

	newOrderC := make(chan elevio.ButtonEvent, config.Buffer)

	go elevio.PollButtons(newOrderC)

	var stashType StashType
	var newOrder elevio.ButtonEvent
	var deliveredOrder elevio.ButtonEvent
	var newState elevator.State
	var peers peers.PeerUpdate
	var cs CommonState

	//Starter disconecttimer for å sjekke om det har vert en disconnect etter en hvis tid
	disconnectTimer := time.NewTimer(config.DisconnectTime)
	heartbeat := time.NewTicker(config.HeartbeatTime)

	idle := true
	offline := false

	for {
		select {
		case <-disconnectTimer.C:
			cs.makeOthersUnavailable(id)
			fmt.Println("Lost connection to network")
			offline = true

		case peers = <-peersC:
			cs.makeOthersUnavailable(id)
			idle = false

		case <-heartbeat.C:
			networkTx <- cs

		default:
		}

		switch {
		case idle:
			select {
			case newOrder = <-newOrderC:
				stashType = Add
				cs.prepNewCs(id)
				cs.addOrder(newOrder, id)
				cs.Ackmap[id] = Acked
				idle = false

			case deliveredOrder = <-deliveredOrderC:
				stashType = Remove
				cs.prepNewCs(id)
				cs.removeOrder(deliveredOrder, id)
				cs.Ackmap[id] = Acked
				idle = false

			case newState = <-newStateC:
				stashType = State
				cs.prepNewCs(id)
				cs.updateState(newState, id)
				cs.Ackmap[id] = Acked
				idle = false

			case arrivedCs := <-networkRx:
				disconnectTimer = time.NewTimer(config.DisconnectTime)
				if arrivedCs.SeqNum > cs.SeqNum || (arrivedCs.Origin > cs.Origin && arrivedCs.SeqNum == cs.SeqNum) {
					cs = arrivedCs
					cs.makeLostPeersUnavailable(peers)
					cs.Ackmap[id] = Acked
					idle = false
				}

			default:
			}

		case offline:
			select {
			case <-networkRx:
				if cs.States[id].CabRequests == [config.NumFloors]bool{} {
					fmt.Println("Regained connection to network")
					offline = false
				} else {
					cs.Ackmap[id] = NotAvailable
				}

			case newOrder := <-newOrderC:
				if !cs.States[id].State.Motorstop {
					cs.Ackmap[id] = Acked
					cs.addCabCall(newOrder, id)
					confirmedCsC <- cs
				}

			case deliveredOrder := <-deliveredOrderC:
				cs.Ackmap[id] = Acked
				cs.removeOrder(deliveredOrder, id)
				confirmedCsC <- cs

			case newState := <-newStateC:
				if !(newState.Obstructed || newState.Motorstop) {
					cs.Ackmap[id] = Acked
					cs.updateState(newState, id)
					confirmedCsC <- cs
				}

			default:
			}

		case !idle:
			select {
			case arrivedCs := <-networkRx:
				if arrivedCs.SeqNum < cs.SeqNum {
					break
				}
				disconnectTimer = time.NewTimer(config.DisconnectTime)

				switch {
				case arrivedCs.SeqNum > cs.SeqNum || (arrivedCs.Origin > cs.Origin && arrivedCs.SeqNum == cs.SeqNum):
					cs = arrivedCs
					cs.Ackmap[id] = Acked
					cs.makeLostPeersUnavailable(peers)

				case arrivedCs.fullyAcked(id):
					cs = arrivedCs
					confirmedCsC <- cs

					switch {
					case cs.Origin != id && stashType != None:
						cs.prepNewCs(id)

						switch stashType {
						case Add:
							cs.addOrder(newOrder, id)
							cs.Ackmap[id] = Acked

						case Remove:
							cs.removeOrder(deliveredOrder, id)
							cs.Ackmap[id] = Acked

						case State:
							cs.updateState(newState, id)
							cs.Ackmap[id] = Acked
						}

					case cs.Origin == id && stashType != None:
						stashType = None
						idle = true

					default:
						idle = true
					}

				case cs.equals(arrivedCs):
					cs = arrivedCs
					cs.Ackmap[id] = Acked
					cs.makeLostPeersUnavailable(peers)

				default:
				}
			default:
			}
		}
	}
}
