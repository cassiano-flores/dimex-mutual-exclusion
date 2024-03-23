O arquivo adicionado **useDIMEX-f.go** gera o arquivo **mxOUT.txt** que faz append de:
  - **" | "** quando entra no MX
  - **" . "** quando sai de MX

Todos processos acessam o mesmo arquivo. Este sistema funciona na mesma máquina ou em máquinas que compartilham o sistema de arquivos (enxergam mesmos arquivos).

Se você deixar rodar por algum tempo, obterá um registro de milhares de entradas e saídas do MX. No arquivo nunca poderá ser encontrada a sub-string **"||"** significando duas entradas consecutivas no MX sem uma saída entre elas.

Analogamente, a sub-string **".."** também não deve ser encontrada. Ou seja, o arquivo é composto somente por sequências de **"|."**

Após a execução você pode abrir o arquivo **mxOUT.txt** em um editor txt comum e procurar por **"||"**, devendo dar zero ocorrências.