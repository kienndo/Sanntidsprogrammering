package backup
import (
    "encoding/json"
    "fmt"
    "net"
    "os"
    "os/exec"
    "time"
)

const (
    udpPort     = 8000 // UDP port
    checkPeriod = 1 * time.Second
)

type Data struct {
    PrimaryAlive bool `json:"primary_alive"`
    LastNumber   int  `json:"last_number"`
}

var (
    OurData = Data{
        PrimaryAlive: true,
        LastNumber:   1,
    }
)

func RunPrimary() {
    fmt.Println("Running as Primary")

    if data, err := os.ReadFile("status.txt"); err == nil {
        if err := json.Unmarshal(data, &OurData); err != nil {
            fmt.Println("Error unmarshaling JSON:", err)
        }
    }

    StartBackupProcess()

    go func() {
        udpAddr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", udpPort))

        conn, _ := net.ListenUDP("udp", udpAddr)

        defer conn.Close()

        buf := make([]byte, 1024)
        for {
            n, _, _ := conn.ReadFromUDP(buf)

            var receivedData Data
            if err := json.Unmarshal(buf[:n], &receivedData); err != nil {
                fmt.Println("Error unmarshaling JSON:", err)
                continue
            }

            fmt.Println("Received message from backup:", receivedData)
        }
    }()

    for {
        fmt.Println(OurData.LastNumber)

        os.WriteFile("status.txt", SerializeData(OurData), 0666)

        SendUDPMessage("localhost", udpPort, OurData)

        time.Sleep(1 * time.Second)
    }
}

func RunBackup() {
    fmt.Println("Running as Backup")
    for {
        if PrimaryIsActive() {
            fmt.Println("Primary is active")
        } else {
            fmt.Println("Primary is inactive, taking over.")
            RunPrimary()
            return
        }
        time.Sleep(checkPeriod)
    }
}

func PrimaryIsActive() bool {
    info, _ := os.Stat("status.txt")

    return time.Since(info.ModTime()) < 2*checkPeriod
}

func StartBackupProcess() {
    cmd := exec.Command("gnome-terminal", "--", "go", "run", "main.go", "backup")

    if err := cmd.Start(); err != nil {
        fmt.Println("Failed to start backup process:", err)
    }
}

func SendUDPMessage(host string, port int, data Data) {
    jsonData, err := json.Marshal(data)
    if err != nil {
        fmt.Println("Error marshaling JSON:", err)
        return
    }

    udpAddr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))

    conn, err := net.DialUDP("udp", nil, udpAddr)
    if err != nil {
        fmt.Println("Error dialing UDP:", err)
        return
    }
    defer conn.Close()

    _, err = conn.Write(jsonData)
    if err != nil {
        fmt.Println("Error sending UDP message:", err)
        return
    }
}

func SerializeData(data Data) []byte {
    jsonData, err := json.Marshal(data)
    if err != nil {
        fmt.Println("Error marshaling JSON:", err)
        return nil
    }
    return jsonData
}

func main() {
    args := os.Args[1:]
    if len(args) > 0 && args[0] == "backup" {
        RunBackup()
    } else {
        RunPrimary()
    }
}
