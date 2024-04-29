package main

import (
	"flag"
	"fmt"
	"root/assigner"
	"root/config"
	"root/distributor"
	"root/elevator"
	"root/elevio"
	"root/lights"
	"root/network/bcast"
	"root/network/peers"
	"strconv"
)

var Port int
var id int

func main() {

	port := flag.Int("port", 15657, "<-- Default value, override with command line argument -port=xxxxx")
	elevatorId := flag.Int("id", 0, "<-- Default value, override with command line argument -id=x")
	flag.Parse()

	Port = *port
	id = *elevatorId

	elevio.Init("localhost:"+strconv.Itoa(Port), config.NumFloors)

	fmt.Println("Elevator initialized with ID", id, "on port", Port)
	fmt.Println("System has", config.NumFloors, "floors and", config.NumElevators, "elevators.")

	newOrderC			:= make(chan elevator.Orders, config.Buffer)
	deliveredOrderC		:= make(chan elevio.ButtonEvent, config.Buffer)
	newStateC			:= make(chan elevator.State, config.Buffer)
	confirmedCsC		:= make(chan distributor.CommonState, config.Buffer)
	networkTx			:= make(chan distributor.CommonState, config.Buffer)
	networkRx			:= make(chan distributor.CommonState, config.Buffer)
	peersRx				:= make(chan peers.PeerUpdate, config.Buffer)
	peersTx				:= make(chan bool, config.Buffer)

	go peers.Receiver(config.PeersPortNumber, peersRx)
	go peers.Transmitter(config.PeersPortNumber, id, peersTx)

	go bcast.Receiver(config.BcastPortNumber, networkRx)
	go bcast.Transmitter(config.BcastPortNumber, networkTx)

	go distributor.Distributor(
		confirmedCsC,
		deliveredOrderC,
		newStateC,
		networkTx,
		networkRx,
		peersRx,
		id)

	go elevator.Elevator(
		newOrderC,
		deliveredOrderC,
		newStateC)

	for {
		select {
		case cs := <-confirmedCsC:
			newOrderC <- assigner.CalculateOptimalOrders(cs, id)
			lights.SetLights(cs, id)

		default:
			continue
		}
	}
}
