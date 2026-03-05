# Oraculo UI — Filosofia

## 1. Propósito

A UI é a **superfície de observação e controle** do Oraculo para humanos.

Os agentes do Oraculo trabalham de forma autônoma — decompondo tarefas, escrevendo código, executando QA — tudo coordenado pelo orquestrador através do CLI Trust Layer. O humano precisa ver o que está acontecendo, entender o porquê, e intervir nos momentos críticos. A UI torna isso possível sem interromper a operação do sistema.

A UI não introduz novos fluxos de dados. Ela consome dados que já existem na camada de armazenamento do CLI, apresentando-os de forma visual e interativa. Agentes e CLI funcionam de forma idêntica com ou sem o dashboard aberto. A UI é um complemento à compreensão humana — nunca uma dependência para o funcionamento correto do sistema.

## 2. Crença Central

**O papel do humano é observar, entender e aprovar — a UI existe para tornar cada um desses gestos natural.**

O sistema já funciona. Os agentes executam. O CLI aplica os contratos. O DAG gerencia a ordenação. O que o humano precisa não é de mais controle — é de mais clareza. A UI oferece uma janela para um sistema em execução, revelando a informação certa no momento certo para que o humano tome decisões confiantes sem desacelerar a máquina.

## 3. Fundamentos Teóricos

O design da UI é embasado em três conceitos consolidados de interação humano-computador e engenharia de sistemas.

### 3.1 Situational Awareness (Mica Endsley)

Situational awareness opera em três níveis: percepção do estado atual, compreensão do seu significado, e projeção do que acontece a seguir. A UI precisa servir aos três. Ela mostra quais agentes estão ativos e o que estão fazendo (percepção). Mostra como as tarefas se relacionam no DAG e o que está bloqueando o progresso (compreensão). Mostra o que acontecerá quando a tarefa atual concluir e o que exige atenção humana a seguir (projeção).

**Por que importa:** Um dashboard que exibe dados brutos sem contexto força o humano a construir um modelo mental do zero a cada olhada. A UI precisa fazer esse trabalho de síntese — apresentando não apenas fatos, mas significado.

### 3.2 Information Radiator (Alistair Cockburn)

Um information radiator é um display posicionado onde as pessoas podem vê-lo sem esforço. Não exige login, navegação ou ação ativa. A informação simplesmente está lá — sempre atualizada, sempre visível. A UI funciona como o information radiator do Oraculo — um display persistente e ambiente do estado do sistema que o humano absorve de relance.

**Por que importa:** Se verificar o status do projeto exige navegar menus, executar queries ou interpretar logs, ele não será verificado com frequência suficiente. A UI precisa tornar o status visível por padrão — não sob demanda.

### 3.3 Human-in-the-Loop

O sistema gerencia a execução de forma autônoma, mas certas decisões exigem julgamento humano: aprovar requisitos de épico, aceitar definições de story, resolver escalações de QA e sobrepor recomendações de agentes. A UI é a superfície onde essas decisões são apresentadas, revisadas e confirmadas. O humano está no loop nos gates de aprovação — não no caminho de execução.

**Por que importa:** A disciplina socrática do Oraculo exige que artefatos críticos recebam validação humana antes de o trabalho avançar. A UI precisa tornar a aprovação confortável e bem fundamentada — apresentando todo o contexto relevante para que o humano decida com rapidez e confiança, sem precisar buscar informações de suporte.

## 4. Princípios de Design

### 4.1 Observar, Não Interferir

A UI mostra o que o sistema está fazendo sem se tornar um gargalo. Os agentes não aguardam a UI renderizar, reconhecer ou sincronizar. O dashboard se atualiza ao estado do sistema de forma assíncrona. Se a UI travar, ficar offline ou atrasar, o sistema continua operando normalmente. A camada de observação nunca constrange a camada de execução.

### 4.2 O CLI Permanece o Trust Layer

A UI nunca escreve diretamente no armazenamento ou no sistema de arquivos. Toda mutação passa pelo CLI, preservando todas as validações, contratos e invariantes garantidos pelo Trust Layer. O dashboard traduz ações do usuário em operações de CLI — é uma ponte, não um atalho. Isso significa que toda ação disponível na UI também está disponível no terminal, e ambos os caminhos produzem resultados idênticos.

### 4.3 Aprovação É o Momento do Humano

O propósito principal da UI é tornar a revisão humana confortável, bem fundamentada e decisiva. Renderização rica de artefatos em markdown, exibição contextual de decisões relacionadas, diff views entre versões de documentos, e ações claras de aceitar/rejeitar — tudo serve à decisão de revisão. Para reviews de documentos (requisitos de epic, definições de story), o dashboard mostra snapshots versionados com diffs entre versões e aceita verdicts `approved`/`rejected`. Para gates operacionais (design, execution-plan, qa-escalation), o dashboard também suporta `needs_revision`. A UI não apressa o humano. Ela apresenta a informação completa e aguarda. É aqui que a disciplina socrática se torna tangível: o humano vê o raciocínio completo antes de se comprometer.

### 4.4 Tempo Real Sem Acoplamento

Conexões ao vivo fornecem atualizações conforme agentes progridem, tarefas concluem e veredictos de QA chegam. Mas atualizações ao vivo são uma conveniência, não um contrato. O sistema produz resultados corretos independentemente de alguém estar assistindo. A UI pode reconectar, atualizar e reconstruir seu estado inteiramente a partir do CLI a qualquer momento. Nenhuma informação existe apenas no live stream.

### 4.5 Mission Control, Não Cockpit

A metáfora é o Mission Control da NASA — não a cabine de pilotagem de um avião. O humano observa telemetria abrangente, toma decisões estratégicas e intervém quando uma escalação exige. O humano não pilota agentes individualmente, não edita código pelo dashboard, não microgerencia a execução de tarefas. O orquestrador pilota o avião. A UI fornece a situational awareness que torna possível a supervisão estratégica.

### 4.6 Tudo em Um

A UI não é uma aplicação separada. Ela vive dentro do mesmo binário que o CLI — um único artefato que contém a interface de linha de comando, o servidor do dashboard, a camada de comunicação com os agentes e os assets do frontend. Não há instalação separada, runtime adicional nem processo extra para gerenciar. Se o dashboard exigir sua própria configuração, sua própria stack tecnológica ou seu próprio deploy, ele se torna um fardo que os times vão ignorar. A UI herda o princípio de distribuição sem dependências do CLI: um binário, tudo incluído.

### 4.7 Rode o Agente, Nós Cuidamos do Resto

O humano nunca deveria precisar iniciar o dashboard manualmente. Quando uma sessão começa, o dashboard sobe automaticamente — o servidor fica ativo, o browser abre, e a interface está pronta antes de o humano fazer sua primeira pergunta. O humano executa um comando, e o dashboard já está lá, esperando. Esse é o Pit of Success aplicado à UI: a configuração correta é a configuração padrão. Sem aba de terminal separada, sem inicialização manual do servidor, sem "lembre-se de abrir o dashboard primeiro." O sistema cuida de si mesmo para que o humano possa focar no trabalho.

## 5. Relação com as Outras Camadas

A UI ocupa a camada mais externa da arquitetura do Oraculo. Ela depende de todas as camadas abaixo, mas nada depende dela.

**CLI Trust Layer** é a única interface da UI com o estado do sistema. O servidor do dashboard chama comandos do CLI e apresenta sua saída. A UI herda todas as garantias de validação do CLI justamente por nunca os contornar.

**Agentes** são observados, mas nunca dirigidos através da UI. O humano vê a atividade dos agentes, o progresso das tarefas e os resultados de QA. A intervenção acontece através dos gates de aprovação definidos pelo workflow — não por comandos ad-hoc a agentes em execução.

**Commands e Skills** definem os workflows que produzem os artefatos que a UI exibe. A UI renderiza requisitos de épico, definições de story e status de tarefas — tudo gerado por skills e persistido através do CLI. A UI não define novos workflows; ela torna os existentes visíveis.

**Tanto humanos quanto agentes passam pelo mesmo CLI.** A UI torna os dados do CLI visuais e interativos. É uma lente diferente sobre a mesma fonte de verdade — não um caminho paralelo de dados.

## 6. O Que a UI Não É

A UI **não é uma ferramenta de gestão de projetos**. Ela não substitui Jira, Linear ou qualquer tracker externo. Ela mostra o estado interno do Oraculo — épicos, stories, tarefas, atividade de agentes — para os humanos que estão trabalhando ativamente com o sistema.

A UI **não é um editor de código**. Humanos não escrevem nem modificam código pelo dashboard. O código é produzido por agentes sob disciplina de TDD e revisado pelo QA. A UI mostra resultados e diffs — não arquivos-fonte editáveis.

A UI **não é necessária para o Oraculo funcionar**. Toda capacidade que a UI expõe existe primeiro no CLI. Um time pode rodar o Oraculo inteiramente pelo terminal. A UI adiciona visibilidade e conforto — não adiciona capacidade.

A UI **não é uma fonte de dados**. Ela lê do CLI e exibe o que encontra. Ela nunca gera, transforma ou armazena dados autoritativos. Se a UI discordar do CLI, a UI está errada.
