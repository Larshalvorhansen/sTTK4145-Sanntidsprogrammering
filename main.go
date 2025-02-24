package main

//Importerer resten av koden og relevante pakker
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

//Definerer port og id som global port og lokal id variabler
var Port int
var id int

func main() {
	//Setter port og elevator id slik at de kan overskrives fra terminal
	port := flag.Int("port", 15657, "<-- Default value, override with command line argument -port=xxxxx")
	elevatorId := flag.Int("id", 0, "<-- Default value, override with command line argument -id=x")
	flag.Parse()

	Port = *port
	id = *elevatorId

	//Initialiserer en "Driver" med portnummer og antall etasjer. Antall etasjer er definert i config.go
	elevio.Init("localhost:"+strconv.Itoa(Port), config.NumFloors)

	//Bare user interface printing til ternimal med info:
	fmt.Println("Elevator initialized with id", id, "on port", Port)
	fmt.Println("System has", config.NumFloors, "floors and", config.NumElevators, "elevators.")

	//Definerer mange ting i main utifra diverse typer definert i importerte filer og allokerer plass utifra størrelsen på config buffer(=1024 i dette tilfellet). 
	newOrderC 		:= make(chan elevator.Orders, config.Buffer)
	deliveredOrderC	:= make(chan elevio.ButtonEvent, config.Buffer)
	newStateC 		:= make(chan elevator.State, config.Buffer)
	confirmedCsC	:= make(chan distributor.CommonState, config.Buffer)
	networkTx 		:= make(chan distributor.CommonState, config.Buffer)
	networkRx 		:= make(chan distributor.CommonState, config.Buffer)
	peersRx 		:= make(chan peers.PeerUpdate, config.Buffer)
	peersTx 		:= make(chan bool, config.Buffer)

//Starter gorutiner for peers? og bcast? som er tråder som kjører samtidig
	//Peers er for hver enkelt "peer" siden vi kjører peer to peer
	go peers.Receiver(config.PeersPortNumber, peersRx)
	go peers.Transmitter(config.PeersPortNumber, id, peersTx)

	//bcast(Broadcast) er for tingen som sender melsinger mellom peers? 
	go bcast.Receiver(config.BcastPortNumber, networkRx)
	go bcast.Transmitter(config.BcastPortNumber, networkTx)

	//distributøren er en fsm som hånderer og distribuerer ting
	go distributor.Distributor(
		confirmedCsC,
		deliveredOrderC,
		newStateC,
		networkTx,
		networkRx,
		peersRx,
		id)

	// Elevator er en funkjson i filen elevatorFSM som tar inn ordere osv og oppfører seg utifra det
	go elevator.Elevator(
		newOrderC,
		deliveredOrderC,
		newStateC)

	//Evig løkke som sjekker om det kommer nye meldinger på kanalen "confirmedCsC"
	//(Gjør at programmet ikke stopper)
	for {
		select {
		case cs := <-confirmedCsC:
			newOrderC <- assigner.CalculateOptimalOrders(cs, id)
			lights.SetLights(cs, id)

		default://Hvis ikke case cs så bare fortsetter den i for løkken
			continue
		}
	}
}
