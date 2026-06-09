# Relatório Técnico:

**Disciplina:** Sistemas Distribuídos \
**Projeto:** Motor de Jogo P2P com Sincronização via MQTT e UDP Hole Punching \
**Equipa:** Talles André Lopes Lima \
**Data:** 08/06/2026

## 1. Justificativa da Escolha: Sistemas Publicar-Assinar (Publish-Subscribe)
A opção selecionada foi a **Opção B: Sistemas Publicar-Assinar (Publish-Subscribe)**. Mas por obra do acaso mqtt já havia sido incorporado antes mesmo da etapa 4.

A escolha fundamenta-se na necessidade crítica de um mecanismo de descoberta de pares (peer discovery) que não dependa de endereçamento IP estático ou conhecimento prévio da topologia da rede pelos clientes. Ao utilizar o padrão Pub-Sub através de um broker MQTT (`broker.emqx.io`), alcançámos o **desacoplamento espacial**, onde os remetentes e destinatários não precisam de conhecer a identidade ou localização física um do outro para estabelecer a comunicação inicial.

O MQTT foi escolhido por ter um broker publico que poderia ser usado como ponto de partida,o cumprimento dos requisitos só veio a ser uma conhecidencia.

## 2. Análise de Desempenho e Complexidade
A introdução do *broker* MQTT como intermediário traz flexibilidade, mas introduz custos técnicos que foram mitigados no meu design:

* **Sobrecarga de Desempenho (Overhead):** A comunicação via broker introduz um salto adicional (*hop*) de rede. Para mitigar o impacto na latência do *gameplay*, o broker é utilizado apenas na fase de sinalização (descoberta). Uma vez estabelecido o canal UDP direto através do *Hole Punching*, o broker é descartado do fluxo de dados principal, garantindo que o movimento do jogador ocorra em comunicação direta P2P, mantendo a performance exigida em tempo real.
* **Complexidade de Gerenciamento:** A gerência de estado no broker exige tratamento de mensagens retidas (*retained messages*) para que novos jogadores encontrem a sala. A complexidade foi gerenciada através de uma implementação *stateless* baseada em eventos, onde o sistema se autorrecupera caso o broker seja desconectado temporariamente.

## 3. Demonstração de Desacoplamento e Robustez

### Desacoplamento Espacial
O sistema prova o desacoplamento espacial através do fluxo de entrada: um novo jogador entra na sala subscrevendo-se ao tópico de sinalização (`game/signal/sala1`) sem conhecer o IP ou a porta do *host*. A troca de endereços ocorre via broker, eliminando a configuração manual de endereços remotos.

### Robustez e Tratamento de Falhas
* **Falha do Broker:** A arquitetura permite que, caso o broker MQTT falhe temporariamente, os pares que já realizaram o *Hole Punching* continuem a trocar pacotes UDP diretamente, mantendo a sessão de jogo ativa.
* **Recuperação de Estado:** Ao utilizar mensagens retidas no MQTT para a criação de salas (`topicSalas`), garantimos que, se o criador da sala deixar de existir, a informação de existência daquela sessão seja persistida pelo broker e disponibilizada automaticamente aos clientes que tentarem se conectar.
* **Silent Discard:** Implementamos uma política de tratamento de mensagens onde mensagens malformados (fora do formato esperado pela nossa especificação de protocolo) são descartados silenciosamente. Isso garante que o motor de renderização não sofra *panics* ou encerramentos inesperados devido a pacotes corrompidos, garantindo estabilidade durante a execução.


## 4. Especificação do Protocolo TUDP (RFC T)

```text
Network Working Group                                 T. A. L. Lima
Request for Comments: T                               UFC - Campus Quixadá
Category: Informational                               Junho 2026

             TUDP: T User Datagram Protocol para 
                 Sincronização de Entidades em Tempo Real
```

### 4.1. Resumo
Este documento especifica o TUDP (T User Datagram Protocol), um protocolo de aplicação sobre UDP desenhado para transmissão de estado de entidades em jogos multijogador P2P de alta performance. O foco do protocolo é a minimização de overhead de serialização, optando por uma sintaxe em texto simples codificada em ASCII/UTF-8, separada por espaços, permitindo uma leitura direta e inspeção de pacotes sem ferramentas complexas de desempacotamento.

### 4.2. Convenções
As palavras-chave "DEVE", "NÃO DEVE", "REQUERIDO", "DEVERÁ", "NÃO DEVERÁ", "DEVERIA", "NÃO DEVERIA", "RECOMENDADO", "PODE", e "OPCIONAL" neste documento devem ser interpretadas conforme descrito na RFC 2119.

### 4.3. Visão Geral do Protocolo
O TUDP opera num modelo sem estado (*stateless*) via UDP, onde cada datagrama contém a informação absoluta do estado atual de uma entidade. Como a entrega de pacotes UDP não é garantida, o protocolo assume que mensagens perdidas serão sobrepostas pela próxima mensagem de estado.

### 4.4. Especificação da Mensagem
Todas as mensagens TUDP DEVEM terminar implicitamente no fim do datagrama UDP (sem necessidade de terminadores como `\r\n`). Os campos DEVEM ser separados por exatamente um espaço em branco (ASCII `0x20`).

A estrutura base de uma mensagem é:
```text
<COMANDO> <ID_ENTIDADE> <PAYLOAD...>
```

#### 4.4.1. O Comando MOVE
O comando `move` é utilizado para atualizar a posição espacial e a direção de uma entidade no plano 2D.

**Sintaxe:**
```text
move <ID> <X> <Y> <D>
```

**Restrições de Formato:**
* **`<ID>`**: String alfanumérica. NÃO DEVE conter espaços em branco (Ex: `"go-123456"`).
* **`<X>`**: Número inteiro de 32-bits (Base 10) representando a coordenada X.
* **`<Y>`**: Número inteiro de 32-bits (Base 10) representando a coordenada Y.
* **`<D>`**: Número inteiro (0 a 4) representando a direção (Ex: `0` = Stop, `1` = Up, `2` = Down, `3` = Left, `4` = Right).

**Exemplo de datagrama válido:**
```text
move go-987654321 150 200 2
```

### 4.5. Tratamento de Erros e Segurança
* Se um receptor ler um datagrama malformado (ex: número incorreto de argumentos para um comando, ou tipos de dados inválidos), ele **DEVE** descartar o pacote silenciosamente para evitar quebras no loop principal de renderização.
* O delimitador estrito de campo é apenas o espaço (ASCII `0x20`). O uso de IDs com espaços quebra a gramática e invalida o pacote.

### 4.6. Considerações de Performance
Ao evitar bibliotecas de serialização complexas (como JSON ou Protobuf), o TUDP reduz o overhead computacional, sendo ideal para sistemas que priorizam arquiteturas minimalistas e compilação estática autónoma, permitindo um processamento de strings altamente otimizado ao nível da linguagem de implementação.