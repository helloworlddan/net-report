package main

import (
    "context"
    "fmt"
    "log"
    "net"
    "os"
    "time"
    "encoding/json"

    "github.com/sparrc/go-ping"
    "cloud.google.com/go/pubsub"
)

type PingResult struct {
    RTTMicros int64 `json:"rtt_micros"`
    PacketSizeBytes int `json:"packet_size_bytes"`
    PacketTTL int `json:"packet_ttl"`
    TargetIP string `json:"target_ip_addr"`
    TargetHostName string `json:"target_host_name"`
    SourceIP string `json:"source_ip_addr"`
    SourceHostName string `json:"source_host_name"`
    Timestamp string `json:"timestamp"`
    UnixTime int64 `json:"unix_time"`
}

func report(hostName string, topicName string) (err error){
    fmt.Printf("topic is %s\n", topicName)

    pinger, err := ping.NewPinger(hostName)
    if err != nil {
        return err
    }

    localIP, err := getOutboundIP()
    if err != nil {
        return err
    }

    localhost, err := os.Hostname()
    if err != nil {
        return err
    }

    ctx := context.Background()
    client, err := pubsub.NewClient(ctx, os.Getenv("GCP_PROJECT_ID"))
    if err != nil {
        return err
    }
    topic := client.Topic(topicName)

    pinger.OnRecv = func(pkt *ping.Packet) {
        stamp := time.Now()
        current := PingResult{
            RTTMicros: pkt.Rtt.Microseconds(),
            PacketSizeBytes: pkt.Nbytes,
            PacketTTL: pkt.Ttl,
            TargetIP:  pkt.IPAddr.String(),
            TargetHostName: pkt.Addr,
            SourceIP: localIP.String(),
            SourceHostName: localhost,
            Timestamp: stamp.Format("2006-01-02T15:04:05.999999"),
            UnixTime: stamp.Unix(),
        }

        jsonData, err := json.Marshal(current)
        if err != nil {
            fmt.Printf("failed to serialize json: %v\n", err)
        }
        result := topic.Publish(ctx, &pubsub.Message{Data: jsonData})
        _, err = result.Get(ctx)
        if err != nil {
            fmt.Printf("failed to publish message: %v\n", err)
        }
        fmt.Printf("published message: %s\n", string(jsonData))
    }

    pinger.SetPrivileged(true)
    pinger.Run()
    return nil
}

func getOutboundIP() (IP net.IP, err error){
    conn, err := net.Dial("udp", "8.8.8.8:53")
    if err != nil {
        return nil, err
    }
    defer conn.Close()
    localAddr := conn.LocalAddr().(*net.UDPAddr)
    return localAddr.IP, nil
}

func main () {
    err := report(os.Getenv("TARGET_HOST"), os.Getenv("REPORT_TOPIC"))
    if err != nil {
        log.Fatalf("fatal: %v", err)
    }
}
