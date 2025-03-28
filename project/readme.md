# Elevator System

This is a real-time programming project where a distributed system is created to control n elevators across m floors. Our solution is a P2P solution using UDP broadcast to periodically update the worldview of all elevators.

## Installation and usage
*NOTE* The installation script only works for linux. To set up and run the project, you need the following tools:

- Go runtime
- Bash

To download the system, do the following

1. Clone the repository.
2. Run `bash build.sh`
3. Acquire an elevator simulator or a physical elevator. This step is omitted because it is not needed in Sanntidslaben.

To run the system, navigate into the `main` folder and run `./elev_sys`.

### Command line arguments

`id`: the id of the peer, default 1
`elevatorPort`: The port where the connection to the elevator server is made. Default 15657
`elevatorIp`: The IP address of the elevator server, default localhost
`communicationPort`: The port used for state updates, default 20060
`peerPort`: The port used for keep-alive signals, default 21060
`numFloors`: The number of floors in the elevator system, default 4

Example usage (simulating two elevators on 1 computer)
`./elev_sys -id=2 -elevatorPort=15658`


## System overview
A single elevator consists of the modules `elevatorStateMachine`, `state`, `communication` and `assigner`.

`elevatorStateMachine` runs the elevator based on orders from assigner. It calls on the handed-out `elevio` to set physical state.

`assigner` assigns orders to a single elevator based on the worldview state provided by `communication`. It is an interface for the handed-out `hall_request_assigner`.

`communication` handles sending, receiving and updating of worldview state from the other peers. It calls on procedures given in the handed out `Network-Go`.

`state` is simply a collection of data structures and methods used for communication.

