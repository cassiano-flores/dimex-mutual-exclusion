O arquivo adicionado **useDIMEX-f.go** gera o arquivo **mxOUT.txt** que faz append de:
  - **" | "** quando entra no MX
  - **" . "** quando sai de MX

Todos processos acessam o mesmo arquivo. Este sistema funciona na mesma m�quina ou em m�quinas que compartilham o sistema de arquivos (enxergam mesmos arquivos).

Se voc� deixar rodar por algum tempo, obter� um registro de milhares de entradas e sa�das do MX. No arquivo nunca poder� ser encontrada a sub-string **"||"** significando duas entradas consecutivas no MX sem uma sa�da entre elas.

Analogamente, a sub-string **".."** tamb�m n�o deve ser encontrada. Ou seja, o arquivo � composto somente por sequ�ncias de **"|."**

Ap�s a execu��o voc� pode abrir o arquivo **mxOUT.txt** em um editor txt comum e procurar por **"||"**, devendo dar zero ocorr�ncias.