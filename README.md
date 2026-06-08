# Relatório Técnico: Evolução de Arquitetura com Comunicação Indireta

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