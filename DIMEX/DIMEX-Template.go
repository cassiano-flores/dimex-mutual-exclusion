// Construido como parte da disciplina: Sistemas Distribuidos - PUCRS - Escola Politecnica
// Professor: Fernando Dotti  (https://fldotti.github.io/)

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
	nbrResps  int
	dbg       bool

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
				//fmt.Printf("dimex recebe da rede: ", msgOutro)
				if strings.Contains(msgOutro.Message, "respOK") {
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
	resps := 0

	for _, p := range module.addresses {
		module.Pp2plink.Req <- PP2PLink.Message{
			Dest: p,
			Message: fmt.Sprintf("[reqEntry, %d, %d]", module.id, myTs),
		}
	}

	module.st = queroSC
}

func (module *DIMEX_Module) handleUponReqExit() {
/*
	upon event [ dmx, Exit  |  r  ]  do                 - EVENTO DE SAIDA
		para todo [p, r, ts ] em waiting                    - PARA CADA PROCESSO EM waiting
			trigger [ pl, Send | p , [ respOk, r ]  ]           - ENVIA MENSAGEM DE RESPOSTA (respOk e r)
		estado  := naoQueroSC                               - ATUALIZA ESTADO PARA naoQueroSC
		waiting := {}                                       - LIMPA waiting
*/
	for _, wr := range module.waiting {
		module.Pp2plink.Req <- PP2PLink.Message{
			Dest: wr,
			Message: fmt.Sprintf("[respOk, %d]", module.id),
		}
	}

	module.st = naoQueroSC
	module.waiting = []bool{}
}

// ------------------------------------------------------------------------------------
// ------- tratamento de mensagens de outros processos
// ------- UPON respOK
// ------- UPON reqEntry
// ------------------------------------------------------------------------------------

func (module *DIMEX_Module) handleUponDeliverRespOk(msgOutro PP2PLink.PP2PLink_Ind_Message) {
/*
	upon event [ pl, Deliver | p, [ respOk, r ] ]
		resps++
		se resps = N
			entao trigger [ dmx, Deliver | free2Access ]
			estado := estouNaSC

*/
	module.resps++
	if module.resps == N {
		module.Pp2plink.Req <- PP2PLink.PP2PLink_Req_Message{
			To:      "dmx",
			Message: "free2Access",
		}
		module.estado = estouNaSC
	}
}

func (module *DIMEX_Module) handleUponDeliverReqEntry(msgOutro PP2PLink.PP2PLink_Ind_Message) {
// outro processo quer entrar na SC
/*
	upon event [ pl, Deliver | p, [ reqEntry, r, rts ]  do
		se (estado == naoQueroSC) OR (estado == QueroSC AND  myTs >  ts)
			entao trigger [ pl, Send | p , [ respOk, r ]  ]
		senao
			se (estado == estouNaSC) OR (estado == QueroSC AND  myTs < ts)
				entao postergados := postergados + [p, r ]
				lts.ts := max(lts.ts, rts.ts)
*/
	if module.estado == naoQueroSC || (module.estado == QueroSC && module.myTs > ts) {
		module.Pp2plink.Req <- PP2PLink.PP2PLink_Req_Message{
			To:      msgOutro.From,
			Message: "[respOk, r]",
		}
	} else {
		if module.estado == estouNaSC || (module.estado == QueroSC && module.myTs < ts) {
			module.postergados = append(module.postergados, [msgOutro.From, r])
		}
		module.lts.ts = max(module.lts.ts, rts.ts)
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
