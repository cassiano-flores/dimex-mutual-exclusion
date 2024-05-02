package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Snapshot struct {
	Type          string            `json:"Type"`
	Id            int               `json:"Id"`
	State         int               `json:"State"`
	Waiting       []bool            `json:"Waiting"`
	ChannelStates map[int][]string  `json:"ChannelStates"`
}

// conta o número de processos na seção crítica (SC)
func inv1(snapshot Snapshot) bool {
	count := 0
	for _, state := range snapshot.ChannelStates {
		for _, s := range state {
			if s == "SC" {
				count++
			}
		}
	}
	// Retorna verdadeiro se no máximo um processo está na SC
	return count <= 1
}

func main() {
    // Lê o arquivo snapshots.txt
    data, err := os.ReadFile("snapshots.txt")
    if err != nil {
        fmt.Println("File reading error", err)
        return
    }

    // Parseia o conteúdo do arquivo para um slice de Snapshots
    var snapshots []Snapshot
    err = json.Unmarshal(data, &snapshots)
    if err != nil {
        fmt.Println("Error parsing JSON", err)
        return
    }

    // Para cada snapshot, verifica as invariantes
    for _, snapshot := range snapshots {
        if !inv1(snapshot) {
            fmt.Printf("Invariant 1 violated in snapshot %d\n", snapshot.Id)
        }
        // Adicione aqui chamadas para outras funções de verificação de invariantes
    }
}
