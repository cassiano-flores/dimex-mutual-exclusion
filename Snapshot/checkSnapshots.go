package main

import (
	"SD/DIMEX"
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Snapshot struct {
	Type             string                   `json:"Type"`
	Id               int                      `json:"Id"`
	IdProcess        int							        `json:"IdProcess"`
	State            int                      `json:"State"`
	Waiting          []bool                   `json:"Waiting"`
	SnapshotSaved    bool							        `json:"SnapshotSaved"`
	ChannelStates    map[int]SnapshotResponse `json:"ChannelStates"`
	ReceivedMessages []string                 `json:"ReceivedMessages"`
	SentMessages     []string                 `json:"SentMessages"`
}

type SnapshotResponse struct {
	Type      string `json:"Type"`
	Id        int    `json:"Id"`
	IdProcess int    `json:"IdProcess"`
	State     int    `json:"State"`
	Waiting   []bool `json:"Waiting"`
}

// conta o número de processos na seção crítica (SC)
func inv1(snapshot DIMEX.Snapshot) bool {
	count := 0
	for _, state := range snapshot.ChannelStates {
    if (state.State == 2) {
			count++
    }
	}

	return count <= 1
}

// verifica se todos os processos estão em "noMX", se todos os waitings são falsos e se não há mensagens
func inv2(snapshot DIMEX.Snapshot) bool {
	// para inv2 só me interessa quando todos os processos estão em noMX, se não estão, não preciso verificar
	if (snapshot.State != 0) {  // 0 = noMX
		return true
	}

	for _, state := range snapshot.ChannelStates {
		if state.State != 0 {
			return true
		}
	}

	// chegou aqui, todos os processos estão em noMX, verificar waitings e mensagens
	for _, waiting := range snapshot.Waiting {
		if (waiting) {
			return false
		}
	}

	if len(snapshot.ReceivedMessages) > 0 || len(snapshot.SentMessages) > 0 {
		return false
	}

	return true
}

// verifica se, para cada processo q que está marcado como waiting em p, p está na SC ou quer a SC
func inv3(snapshot DIMEX.Snapshot) bool {
	// verifica waitings do processo principal
	if snapshot.State != 1 && snapshot.State != 2 {  // 1 = wantMX | 2 = inMX
		for _, waiting := range snapshot.Waiting {
			if (waiting) {
				return false
			}
		}
	}
	return true
}

// verifica se, para cada processo q que quer a seção crítica, a soma das mensagens recebidas, das mensagens em trânsito e dos flags waiting para p em outros processos é igual a N-1
func inv4(snapshot DIMEX.Snapshot) bool {
	totalProcesses := 3  // número total de processos

	if (snapshot.State == 1) {  // 1 = wantMX
		sum := len(snapshot.ReceivedMessages) + len(snapshot.SentMessages)
		for _, waiting := range snapshot.Waiting {
			if (waiting) {
				sum++
			}
		}
		if (sum != totalProcesses-1) {
			return false
		}
	}

	return true
}

func main() {
	// abre o arquivo snapshots.txt
	file, err := os.Open("snapshots.txt")
	if (err != nil) {
		fmt.Println("File opening error", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var line string
	var count int
	var snapshotString string

	// lê o arquivo linha por linha
	for scanner.Scan() {
		line = scanner.Text()
		count += strings.Count(line, "}")
		snapshotString += line

		// a cada quarta "}", envia a string lida
		if (count == 4) {
			snapshot := DIMEX.StringToSnapshot(snapshotString)
			//////////////////////////////////////////////////////////////////
			// verifica as invariantes
			if !inv1(snapshot) {
				fmt.Printf("Invariant 1 violated in snapshot %d\n", snapshot.Id)
			}
			if (!inv2(snapshot)) {
				fmt.Printf("Invariant 2 violated in snapshot %d\n", snapshot.Id)
			}
			if (!inv3(snapshot)) {
				fmt.Printf("Invariant 3 violated in snapshot %d\n", snapshot.Id)
			}
			// if (!inv4(snapshot)) {
				// fmt.Printf("Invariant 4 violated in snapshot %d\n", snapshot.Id)
			// }
      //////////////////////////////////////////////////////////////////
			count = 0
			snapshotString = ""
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("File reading error", err)
	}
}
