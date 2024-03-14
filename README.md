# TTK4145 - ELEVATOR PROJECT, GROUP 63

## Functionality
There are some factors to take into consideration with this code. We have implemented a system with master-slave topology, and implemented process pairs for primary and backup in case the master shuts down. Therefore, the first computer that runs the program will become the master and run the primary process, and the other computers that has to be run X seconds after the first computer, will become the backups. By killing the master terminal, currently will both other computers take over as primary. Ideally, only one of them would become the primary and the other a backup, but we were unsuccessful in implementing it.


## Notes regarding the code
We were unsuccessful in implementing the code for taking care of disconnections. We therefore implement the pseudo code here to maintain the code quality in the original code.


# TO DO LIST
- Jeg tror ikke den lagrer MasterHallRequests dersom den faller ut! (Se på gammel kode fra backup, må browse gammel githistorie, jeg har slettet det her)
- Oppdatere MasterHallRequests riktig at betjente ordre fjernes når den regner ut
- Fikse funksjonene for å assigne ordre(hvordan IP legges inn, må kanskje fjerne at det står "peer-" foran og ta ut riktig verdi)
    Tror det enkelt kan løses ved å gjøre om ID som sendes bare til IP-adressen
- Packet loss og watchdogtimer (minste prioritet nå)
- Kodekvalitet(fjerne prints, sortere channels, fjerne overflødige funksjoner og gi navn til ting så det gir mening. Kommentarene skal bare gi kontekst, ikke beskrive hva funksjonen gjør)
- Acknowledge funksjon for UDP
- Konfigurere en json-fil for konstanter?
- 

### Funker nå:
- IP-adresse for master
- Skru ned tiden på master/backup og fyll ut tiden i README og passe på at kun en kjører primary
- Lage funksjoner for lys på hallrequests. Cablys funker.
- TROR DETTE SKAL FUNGERE NÅ:
  - Må gjøre så den ikke kjører hall order, men bare sender det inn! Fikses i FsmOnButtonPress

### Json-fil:
    // Struct for JSONConfig
    type JSONConfig struct {
    Port1 string            'json:"port1"'
    Port2 string            'json:"port2"'
    Port3 string            'json:"port3"'
    DoorOpenDuration int   'json:"doorOpenDuration"'
    N_FLOORS int            'json:"nFloors"'
    N_BUTTONS int          'json:"nButtons"'
    } 


    func(){

    Config := JSONConfig{
    Port1: "15657", // eller bare = den som står i koden
    Port2: "15658",
    Port3: "15659",
    DoorOpenDuration: 3,
    N_FLOORS: 4,
    N_BUTTONS: 3,
    }

    // Open the configuration file
    file, err := os.Open("config.json")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer file.Close()

    // Decode the JSON file into a Config struct
    var config Config
    decoder := json.NewDecoder(file)
    err = decoder.Decode(&config)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    }

## Andre funksjoner vi ikke bruker:

    func ChooseConnection() {
	// Sjekker om channel 1 er ledig
	conn, err := net.ListenPacket("udp",":29503")
	if err != nil {
		fmt.Println("Error listening to channel")
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(3*time.Second))
	_, _, err = conn.ReadFrom(buffer)

	if err != nil {

		// Channel 1
		fmt.Println("Sending to channel 1")
		go ChannelTaken()
		go bcast.RunBroadcast(ChanElevator1, Address1) //Kjøres bare en gang


	} else {

		// Channel 2
		fmt.Println("sending to channel 2")
		go bcast.RunBroadcast(ChanElevator2, Address2)
	}
	time.Sleep(1*time.Millisecond)
    }

    func ChannelTaken() {
	for {
		conn, err := net.Dial("udp", "10.100.23.255:29503")
		if err != nil {
			fmt.Println("Error dialing udp")
		}
		defer conn.Close()
		_, err = conn.Write([]byte("1"))

		time.Sleep(1*time.Second)
	}
    }

    func RecievingState(address string,state *elevio.Elevator) {

	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Println("failed to listen to udp mip")
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Failed to listen from UDP")
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil{
			fmt.Println("Failed to read")
			continue
		}

		var newState elevio.Elevator
		err = json.Unmarshal(buffer[:n],&newState)
		if err != nil {
			fmt.Println("Failed to deserialize")
			continue
		}
		
		CostMutex.Lock()
		*state = newState
		CostMutex.Unlock()
	}
    }

	func ButtonIdentifier(chanButtonRequests chan elevio.ButtonEvent, chanHallRequests chan elevio.ButtonEvent, chanCabRequests chan elevio.ButtonEvent) {
	for{
		select {
			case btnEvent := <-chanButtonRequests:
				if btnEvent.Button == elevio.BT_Cab{
					chanCabRequests <- btnEvent
				} else{
					chanHallRequests <- btnEvent
				}
			}
		}
}


