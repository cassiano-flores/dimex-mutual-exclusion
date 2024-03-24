// Construido como parte da disciplina: Sistemas Distribuidos - PUCRS - Escola Politecnica
// Professor: Fernando Dotti  (https://fldotti.github.io/)

// Autor:
// Cassiano Luis Flores Michel (20204012-7)

/*
  Modulo representando Algoritmo de Exclusao Mutua Distribuida: Semestre 2023/1
	Aspectos a observar:
		mapeamento de modulo para estrutura
	  inicializacao
	  semantica de concorrencia: cada evento eh atomico, modulo trata 1 por vez

	QUESTAO:
	  Implementar o nucleo do algoritmo ja descrito, ou seja, o corpo das funcoes reativas a cada entrada possivel:
	  	handleUponReqEntry()                // recebe do nivel de cima (app)
			handleUponReqExit()                 // recebe do nivel de cima (app)
			handleUponDeliverRespOk(msgOutro)   // recebe do nivel de baixo
			handleUponDeliverReqEntry(msgOutro) // recebe do nivel de baixo
*/

package DIMEX

import (
	PP2PLink "SD/PP2PLink"
	"fmt"
	"strconv"
	"strings"
)

// ------------------------------------------------------------------------------------
// ------- principais tipos
// ------------------------------------------------------------------------------------

type State int // enumeracao dos estados possiveis de um processo
const (
	noMX State = iota
	wantMX
	inMX
)

type dmxReq int // enumeracao dos estados possiveis de um processo
const (
	ENTER dmxReq = iota
	EXIT
)

type dmxResp struct { // mensagem do modulo DIMEX informando que pode acessar - pode ser somente um sinal (vazio)
	// mensagem para aplicacao indicando que pode prosseguir
}

type DIMEX_Module struct {
	Req       chan dmxReq  // canal para receber pedidos da aplicacao (REQ e EXIT)
	Ind       chan dmxResp // canal para informar aplicacao que pode acessar
	addresses []string     // endereco de todos, na mesma ordem
	id        int          // identificador do processo - eh o indice no array de enderecos acima
	st        State        // estado deste processo na exclusao mutua distribuida
	waiting   []bool       // processos aguardando tem flag true
	lcl       int          // relogio logico local
	reqTs     int          // timestamp local da ultima requisicao deste processo
	nbrResps  int          // contador de respostas
	dbg       bool         // flag para depuracao

	Pp2plink *PP2PLink.PP2PLink // acesso a comunicacao pra enviar por PP2PLinq.Req e receber por PP2PLinq.Ind
}

// ------------------------------------------------------------------------------------
// ------- inicializacao
// ------------------------------------------------------------------------------------

func NewDIMEX(_addresses []string, _id int, _dbg bool) *DIMEX_Module {

	p2p := PP2PLink.NewPP2PLink(_addresses[_id], _dbg)

	dmx := &DIMEX_Module{
		Req: make(chan dmxReq, 1),
		Ind: make(chan dmxResp, 1),

		addresses: _addresses,
		id:        _id,
		st:        noMX,
		waiting:   make([]bool, len(_addresses)),
		lcl:       0,
		reqTs:     0,
		dbg:       _dbg,

		Pp2plink: p2p}

	for i := 0; i < len(dmx.waiting); i++ {
		dmx.waiting[i] = false
	}
	dmx.Start()
	dmx.outDbg("Init DIMEX!")
	return dmx
}

// ------------------------------------------------------------------------------------
// ------- nucleo do funcionamento
// ------------------------------------------------------------------------------------

func (module *DIMEX_Module) Start() {

	go func() {
		for {
			select {
			case dmxR := <-module.Req: // vindo da aplicacao
				if dmxR == ENTER {
					module.outDbg("app pede mx")
					module.handleUponReqEntry() // ENTRADA DO ALGORITMO

				} else if dmxR == EXIT {
					module.outDbg("app libera mx")
					module.handleUponReqExit() // ENTRADA DO ALGORITMO
				}

			case msgOutro := <-module.Pp2plink.Ind: // vindo de outro processo
				// fmt.Printf("dimex recebe da rede: ", msgOutro.Message)
				if strings.Contains(msgOutro.Message, "respOk") {
					module.outDbg("         <<<---- responde! " + msgOutro.Message)
					module.handleUponDeliverRespOk(msgOutro) // ENTRADA DO ALGORITMO

				} else if strings.Contains(msgOutro.Message, "reqEntry") {
					module.outDbg("          <<<---- pede??  " + msgOutro.Message)
					module.handleUponDeliverReqEntry(msgOutro) // ENTRADA DO ALGORITMO

				}
			}
		}
	}()
}

// ------------------------------------------------------------------------------------
// ------- tratamento de pedidos vindos da aplicacao
// ------- UPON ENTRY
// ------- UPON EXIT
// ------------------------------------------------------------------------------------

func (module *DIMEX_Module) handleUponReqEntry() {
/*
	upon event [ dmx, Entry  |  r ]  do                  - EVENTO DE ENTRADA
		lts.ts++                                             - AUMENTA O CONTADOR DE TIMESTAMP
		myTs  := lts                                         - ATUALIZA O TIMESTAMP COM O VALOR ATUAL
		resps := 0                                           - ZERA CONTADOR DE RESPOSTAS
		para todo processo p                                 - PARA CADA PROCESSO
			trigger [ pl , Send | [ reqEntry, r, myTs ]          - ENVIA MENSAGEM DE REQUISICAO DE ENTRADA
		estado := queroSC                                    - ATUALIZA ESTADO PARA queroSC
*/
	module.lcl++
	myTs := module.lcl
	module.nbrResps = 0

	for p := 0; p < len(module.addresses); p++ {
		module.sendToLink(module.addresses[p], fmt.Sprintf("[reqEntry, %d, %d]", module.id, myTs), "  ")
	}

	module.st = wantMX
}

func (module *DIMEX_Module) handleUponReqExit() {
/*
	upon event [ dmx, Exit  |  r  ]  do                 - EVENTO DE SAIDA
		para todo [p, r, ts ] em waiting                    - PARA CADA PROCESSO EM waiting
			trigger [ pl, Send | p , [ respOk, r ]  ]           - ENVIA MENSAGEM DE RESPOSTA (respOk e r)
		estado  := naoQueroSC                               - ATUALIZA ESTADO PARA naoQueroSC
		waiting := {}                                       - LIMPA waiting
*/
	for p, waiting := range module.waiting {
		if (waiting) {
			module.sendToLink(module.addresses[p], fmt.Sprintf("[respOk, %d]", module.id), "  ")
		}
	}

	module.st = noMX
	module.waiting = make([]bool, len(module.addresses))
}

// ------------------------------------------------------------------------------------
// ------- tratamento de mensagens de outros processos
// ------- UPON respOk
// ------- UPON reqEntry
// ------------------------------------------------------------------------------------

func (module *DIMEX_Module) handleUponDeliverRespOk(msgOutro PP2PLink.PP2PLink_Ind_Message) {
/*
	upon event [ pl, Deliver | p, [ respOk, r ] ]       - EVENTO DE RECEBIMENTO DE RESPOSTA
		resps++                                             - AUMENTA CONTADOR DE RESPOSTAS
		se resps = N                                        - SE TODOS RESPONDERAM
			entao trigger [ dmx, Deliver | free2Access ]        - ENVIA MENSAGEM DE LIBERACAO
			estado := estouNaSC                                 - ATUALIZA ESTADO PARA estouNaSC
*/
	module.nbrResps++
	if (module.nbrResps == len(module.addresses)) {
		module.Ind <- dmxResp{}
		module.st = inMX
	}
}

func (module *DIMEX_Module) handleUponDeliverReqEntry(msgOutro PP2PLink.PP2PLink_Ind_Message) {
/*
	upon event [ pl, Deliver | p, [ reqEntry, r, rts ]  do                - EVENTO DE RECEBIMENTO DE REQUISICAO
		se (estado == naoQueroSC) OR (estado == QueroSC AND  myTs >  ts)      - SE NAO QUER ACESSAR OU QUER E TEM PRIORIDADE
			entao trigger [ pl, Send | p , [ respOk, r ]  ]                       - ENVIA MENSAGEM DE RESPOSTA (respOk e r)
		senao                                                                 - SENAO
			se (estado == estouNaSC) OR (estado == QueroSC AND  myTs < ts)        - SE ESTA NA SC OU QUER E TEM PRIORIDADE
				entao postergados := postergados + [p, r ]                            - ADICIONA A POSTERGADOS
				lts.ts := max(lts.ts, rts.ts)                                       - ATUALIZA O TIMESTAMP
*/
	parts    := strings.Split(msgOutro.Message, ",")
	othId, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
	othTs, _ := strconv.Atoi((string([]rune(strings.TrimSpace(parts[2]))[0])))

	if ((module.st == noMX) || ((module.st == wantMX) && (othTs > module.reqTs))) {
		module.sendToLink(module.addresses[othId], fmt.Sprintf("[respOk, %d]", module.id), "  ")
	} else {
		if ((module.st == inMX) || ((module.st == wantMX) && (othTs < module.reqTs))) {
			module.waiting[othId] = true
		}
		if (othTs > module.lcl) {
			module.lcl = othTs
		}
	}
}

// ------------------------------------------------------------------------------------
// ------- funcoes de ajuda
// ------------------------------------------------------------------------------------

func (module *DIMEX_Module) sendToLink(address string, content string, space string) {
	module.outDbg(space + " ---->>>>   to: " + address + "     msg: " + content)
	module.Pp2plink.Req <- PP2PLink.PP2PLink_Req_Message{
		To:      address,
		Message: content}
}

func before(oneId, oneTs, othId, othTs int) bool {
	if oneTs < othTs {
		return true
	} else if oneTs > othTs {
		return false
	} else {
		return oneId < othId
	}
}

func (module *DIMEX_Module) outDbg(s string) {
	if module.dbg {
		fmt.Println(". . . . . . . . . . . . [ DIMEX : " + s + " ]")
	}
}
