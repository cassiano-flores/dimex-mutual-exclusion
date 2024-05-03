Em um terminal Powershell no diretório raiz, execute o comando **.\script.ps1** para rodar os processos. Após finalizar, para testar as invariantes, basta ir ao diretório **.\Snapshot** e executar o comando **go run .\checkSnapshots.go**

**1° parte**

O arquivo adicionado **useDIMEX-f.go** gera o arquivo **mxOUT.txt** que faz append de:

-   **" | "** quando entra no MX
-   **" . "** quando sai de MX

Todos processos acessam o mesmo arquivo. Este sistema funciona na mesma máquina ou em máquinas que compartilham o sistema de arquivos (enxergam mesmos arquivos).

Se você deixar rodar por algum tempo, obterá um registro de milhares de entradas e saídas do MX. No arquivo nunca poderá ser encontrada a sub-string **"||"** significando duas entradas consecutivas no MX sem uma saída entre elas.

Analogamente, a sub-string **".."** também não deve ser encontrada. Ou seja, o arquivo é composto somente por sequências de **"|."**

Após a execução você pode abrir o arquivo **mxOUT.txt** em um editor txt comum e procurar por **"||"**, devendo dar zero ocorrências.

**2° parte**

- Implemente o algoritmo de snapshot junto ao DiMEx, usando o algoritmo de Chandy-Lamport discutido em aula.

- Note que na exclusão mútua, todo processo tem um canal com cada outro e que este canal preserva ordem e não perde mensagens. Assim, as suposições do algoritmo de Chandy-Lamport são satisfeitas.

- Conforme o algoritmo, o módulo DiMEx pode receber/tratar também uma mensagem de snapshot.

- Cada snapshot tem um identificador único criado no processo que inicia o mesmo.   

- Todo processo, ao gravar seu estado, grava este identificador junto.

- O estado deve incluir suas variáveis e o estado dos canais, conforme o algoritmo de snapshot. Os (diversos) snapshots completos devem ser avaliados junto ao funcionamento do sistema.

**Realize as seguintes etapas, e demonstre os resultados no dia da apresentação:**

0) Rode o DIMEX com no mínimo 3 processos;

1) Faça um processo iniciar snapshots sucessivos, cada um com um identificador (1, 2, 3 ...) concorrentemente aos seus acessos como um processo usuário do DIMEX;

2) Colha uma sequencia de snapshots, algumas centenas. Eles podem estar em arquivos separados, um para cada processo;

3) Escreva uma ferramenta que avalia para cada snapshot se os estados dos processos estão consistentes.

- Para cada snapshot SnId a ferramenta lê os estados gravados por cada processo, respectivo ao snapshot SnId, e avalia se o mesmo está correto.

- Para isso voce tem que enunciar invariantes do sistema. Invariante é algo que deve ser verdade em qualquer estado.

- Inv 1: no máximo um processo na SC.
- Inv 2: se todos processos estão em "não quero a SC", então todos waitings tem que ser falsos e não deve haver mensagens.
- Inv 3: se um processo q está marcado como waiting em p, então p está na SC ou quer a SC.
- Inv 4: se um processo q quer a seção crítica (nao entrou ainda), então o somatório de mensagens recebidas, de mensagens em transito e de, flags waiting para p em outros processos deve ser igual a N-1 (onde N é o número total de processos) inv ... etc.

- Cada invariante é um teste sobre um snapshot, uma funcao_InvX(snapshot) retorna um bool com o resultado
- Cada snapshot é avaliado para todas invariantes. A ferramenta avisa invariantes violadas e o snapshot.

4) Rode o sistema e avalie com a ferramenta:
- se ela gerou avisos de violacao de invariantes, avalie seu algoritmo (ou o algoritmo de snapshot) para o DIMEX supostamente correto, as invariantes devem todas passar.

5) Insira falhas no DIMEX:
- por exemplo, altere a condicao de resposta para violar a SC, altere a mesma condicao para bloquear

6) Detecte estes casos com a análise de snapshots.
