// Construido como parte da disciplina: Sistemas Distribuidos - PUCRS - Escola Politecnica
// Professor: Fernando Dotti  (https://fldotti.github.io/)

/*
	Uso para exemplo:
	  go run useDIMEX.go 0 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002 ")
	  go run useDIMEX.go 1 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002 ")
	  go run useDIMEX.go 2 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002 ")

	LANCAR N PROCESSOS EM SHELL's DIFERENTES, UMA PARA CADA PROCESSO.
	para cada processo fornecer: seu id unico (0, 1, 2 ...) e a mesma lista de processos.
	o endereco de cada processo eh o dado na lista, na posicao do seu id.
	no exemplo acima o processo com id=1 usa a porta 6001 para receber e as portas
	5000 e 7002 para mandar mensagens respectivamente para processos com id=0 e 2

	Esta versao supoe que todos processos tem acesso a um mesmo arquivo chamado "mxOUT.txt"
	Todos processos escrevem neste arquivo, usando o protocolo dimex para exclusao mutua.
	Os processos escrevem "|." cada vez que acessam o arquivo. Assim, o arquivo com conteudo
	correto devera ser uma sequencia:

	|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.
	|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.
	|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.
	etc etc ...

	ate o usuario interromper os processos (ctl c).
	Qualquer padrao diferente disso, revela um erro: |.|.|.|.|.||..|.|.|.  por exemplo.

	Se voce retirar o protocolo dimex vai ver que o arquivo podera entrelacar "|."
	dos processos de diversas diferentes formas. Ou seja, o padrao correto acima eh garantido pelo dimex.
	Ainda assim, isto eh apenas um teste. E testes sao frageis em sistemas distribuidos.
*/

package main

import (
	"SD/DIMEX"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Please specify at least one address:port!")
		fmt.Println("go run useDIMEX.go 0 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002 ")
		fmt.Println("go run useDIMEX.go 1 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002 ")
		fmt.Println("go run useDIMEX.go 2 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002 ")
		return
	}

	id, _ := strconv.Atoi(os.Args[1])
	addresses := os.Args[2:]
	// fmt.Print("id: ", id, "   ") fmt.Println(addresses)

	var dmx *DIMEX.DIMEX_Module = DIMEX.NewDIMEX(addresses, id, true)
	fmt.Println(dmx)

	// abre arquivo que TODOS processos devem poder usar
	file, err := os.OpenFile("./mxOUT.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close() // garante que o arquivo esta fechado no final da funcao

	// espera para facilitar inicializacao de todos processos (a mao)
	time.Sleep(3 * time.Second)

	for {
		// SOLICITA ACESSO AO DIMEX
		fmt.Println("[ APP id: ", id, " PEDE   MX ]")
		dmx.Req <- DIMEX.ENTER
		//fmt.Println("[ APP id: ", id, " ESPERA MX ]")
		// ESPERA LIBERACAO DO MODULO DIMEX
		<-dmx.Ind //

		// A PARTIR DAQUI ESTA ACESSANDO O ARQUIVO SOZINHO
		_, err = file.WriteString("|") // marca entrada no arquivo
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}

		fmt.Println("[ APP id: ", id, " *EM*   MX ]")

		_, err = file.WriteString(".") // marca saida no arquivo
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}

		// AGORA VAI LIBERAR O ARQUIVO PARA OUTROS
		dmx.Req <- DIMEX.EXIT //
		fmt.Println("[ APP id: ", id, " FORA   MX ]")
	}
}
