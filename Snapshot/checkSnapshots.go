package main

import (
	"SD/DIMEX"
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Snapshot struct {
	Type          string                   `json:"Type"`
	Id            int                      `json:"Id"`
	IdProcess     int							         `json:"IdProcess"`
	State         int                      `json:"State"`
	Waiting       []bool                   `json:"Waiting"`
	SnapshotSaved bool							       `json:"SnapshotSaved"`
	ChannelStates map[int]SnapshotResponse `json:"ChannelStates"`
}

type SnapshotResponse struct {
	Type          string `json:"Type"`
	Id            int    `json:"Id"`
	IdProcess     int    `json:"IdProcess"`
	State         int    `json:"State"`
	Waiting       []bool `json:"Waiting"`
}

// conta o número de processos na seção crítica (SC)
func inv1(snapshot DIMEX.Snapshot) bool {
	count := 0
	
	if (snapshot.State == 2) {  // 2 = inMX
		count++
	}

	for _, state := range snapshot.ChannelStates {
    if (state.State == 2) {
			count++
    }
	}

	return count <= 1
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
		/* 
				if (!inv2(snapshot)) {
					fmt.Printf("Invariant 2 violated in snapshot %d\n", snapshot.Id)
				}
				if (!inv3(snapshot)) {
					fmt.Printf("Invariant 3 violated in snapshot %d\n", snapshot.Id)
				}
				if (!inv4(snapshot)) {
					fmt.Printf("Invariant 4 violated in snapshot %d\n", snapshot.Id)
				}
		*/
			count = 0
			snapshotString = ""
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("File reading error", err)
	}
}
