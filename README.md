# How to run the elevators

1. Set the parameters in config.go to determine the number of elevators, number of floors etc.
2. Run the command `go run main.go -port=xxxx -id=x`. Alternatively run the `program` executable with the same command line arguments.\
The id's must be integers and the first id has to be 0. Additional elevators increment the id by 1.
3. Enjoy the ride(s)

Tree structure of this repo: 
.
├── README.md
├── assigner
│   ├── assigner.go
│   └── executables
│       ├── hall_request_assigner
│       ├── hall_request_assigner.exe
│       └── hall_request_assigner_mac
├── config
│   └── config.go
├── distributor
│   ├── commonstate.go
│   └── distributorFsm.go
├── elevator
│   ├── direction.go
│   ├── door.go
│   ├── elevatorFsm.go
│   └── orders.go
├── elevio
│   └── elevio.go
├── go.mod
├── lights
│   └── lights.go
├── main.go
├── network
│   ├── bcast
│   │   └── bcast.go
│   ├── conn
│   │   ├── bcast_conn_darwin.go
│   │   ├── bcast_conn_linux.go
│   │   └── bcast_conn_windows.go
│   └── peers
│       └── peers.go
└── program

12 directories, 22 files

./assigner/assigner.go                            77
./assigner/executables/hall_request_assigner      5650
./assigner/executables/hall_request_assigner.exe  8785
./assigner/executables/hall_request_assigner_mac  4999
./config/config.go                                19
./distributor/commonstate.go                      101
./distributor/distributorFsm.go                   187
./elevator/direction.go                           28
./elevator/door.go                                78
./elevator/elevatorFsm.go                         205
./elevator/orders.go                              41
./elevio/elevio.go                                199
./go.mod                                          2
./lights/lights.go                                28
./main.go                                         72
./network/bcast/bcast.go                          148
./network/conn/bcast_conn_darwin.go               43
./network/conn/bcast_conn_linux.go                39
./network/conn/bcast_conn_windows.go              68
./network/peers/peers.go                          93