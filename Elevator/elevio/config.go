package elevio

// import(
// 	"fmt"
// 	"os"
// 	"encoding/json"
// )

// type JsonConfig struct {
// Port1 string            `json:"port1"`
// Port2 string            `json:"port2"`
// Port3 string            `json:"port3"`
// N_FLOORS int            `json:"nFloors"`
// N_BUTTONS int           `json:"nButtons"`
// } 


// func MakeJsonConfig(){
// config := JsonConfig{
//     Port1: "15657", // eller bare = den som står i koden
//     Port2: "15658",
//     Port3: "15659",
// 	N_FLOORS: N_FLOORS,
//     N_BUTTONS: N_BUTTONS,
// }

//     // Open the configuration file
//     file, err := os.Open("config.json")
//     if err != nil {
//         fmt.Println("Error:", err)
//         return
//     }
//     defer file.Close()

//     // Decode the JSON file into a Config struct
    
//     decoder := json.NewDecoder(file)
//     err = decoder.Decode(&config)
//     if err != nil {
//         fmt.Println("Error:", err)
//         return
//     }
// }


