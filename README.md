# TTK4145 - ELEVATOR PROJECT, GROUP 63

## Functionality
There are some factors to take into consideration with this code. We have implemented a system with master-slave topology, and implemented process pairs for primary and backup in case the master shuts down. Therefore, the first computer that runs the program will become the master and run the primary process, and the other computers that has to be run X seconds after the first computer, will become the backups. By killing the master terminal, currently will both other computers take over as primary. Ideally, only one of them would become the primary and the other a backup, but we were unsuccessful in implementing it.


## Notes regarding the code
We were unsuccessful in implementing the code for taking care of disconnections. We therefore implement the pseudo code here to maintain the code quality in the original code.


# TO DO LIST
- Skru ned tiden på master/backup og fyll ut tiden i README
- Oppdatere MasterHallRequests riktig at betjente ordre fjernes når den regner ut
- Lage funksjoner for lys
- Fikse funksjonene for å assigne ordre(hvordan IP legges inn, må kanskje fjerne at det står "peer-" foran og ta ut riktig verdi)
- Packet loss og watchdogtimer (minste prioritet nå)
- Kodekvalitet(fjerne prints, sortere channels, fjerne overflødige funksjoner og gi navn til ting så det gir mening. Kommentarene skal bare gi kontekst, ikke beskrive hva funksjonen gjør)
- Acknowledge funksjon for UDP
- Konfigurere en json-fil for konstanter?

### Funker nå:
- IP-adresse for master

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


