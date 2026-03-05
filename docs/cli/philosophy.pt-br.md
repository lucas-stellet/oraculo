# Oraculo CLI — Filosofia

## 1. Propósito

O CLI do Oraculo é a **Camada de Confiança** do sistema.

O Oraculo opera em dois níveis fundamentalmente diferentes: trabalho interpretativo e trabalho determinístico. Skills e agentes cuidam do lado interpretativo — diálogo socrático, análise de codebase, raciocínio criativo. O CLI cuida do lado determinístico — persistir dados, aplicar schemas, validar operações, retornar resultados previsíveis.

O CLI traça uma linha dura entre esses dois níveis. O consumidor — seja um agente ou um humano — declara O QUE quer. O CLI garante o COMO com um resultado previsível e validado. Sem adivinhação, sem interpretação, sem ambiguidade.

## 2. Crença Central

**Consumidores devem confiar na ferramenta completamente — agentes e humanos igualmente.**

Quando um agente precisa persistir uma descoberta sobre a codebase, ele não deveria precisar saber que o armazenamento é SQLite, que índices FTS5 requerem triggers de sincronização, ou que o banco de dados fica em `.oraculo/oraculo.db`. Ele deveria chamar `oraculo memory store --domain payments --finding "Uses repository pattern"` e confiar que a ferramenta cuida do resto.

O CLI é o ponto no sistema onde o resultado é determinístico e validado. Ele elimina suposições: o consumidor não precisa adivinhar paths, construir schemas, lembrar formatos ou lidar com edge cases. A interface é o contrato, e o contrato é aplicado.

A mesma interface serve qualquer contexto. Dentro do Claude Code ou fora dele. Chamado por uma skill, um agente, um humano no terminal ou um pipeline de CI/CD. A confiança é incondicional e universal.

## 3. Fundações Teóricas

O design do CLI se fundamenta em três conceitos estabelecidos da engenharia de software.

### 3.1 Design by Contract (Bertrand Meyer)

Cada comando do CLI opera sob um contrato: pré-condições que devem ser atendidas, pós-condições que são garantidas e invariantes que são preservados. O CLI valida estritamente — operações que violam pré-condições são recusadas com erro claro, não silenciosamente aceitas com resultados imprevisíveis.

**Por que importa:** Agentes são raciocinadores poderosos mas operadores de sistema não confiáveis. Eles alucinam paths, esquecem schemas e produzem comandos sutilmente quebrados. Um contrato estrito significa que o agente ou recebe um resultado correto ou uma rejeição clara. Não há meio-termo onde dados ruins entram silenciosamente no sistema.

### 3.2 Pit of Success (Brad Abrams)

O caminho correto é o caminho fácil. O CLI é desenhado para que usá-lo corretamente requer menos esforço do que usá-lo incorretamente. O consumidor não precisa construir caminhos de arquivo, lembrar schemas de banco de dados ou deduzir formatos de saída. O CLI cuida de tudo isso — o consumidor apenas declara sua intenção.

**Por que importa:** Quando a coisa fácil de fazer também é a coisa certa de fazer, erros se tornam raros. O CLI não depende de consumidores lerem documentação cuidadosamente ou lembrarem convenções. Ele torna o errado difícil e o certo óbvio.

### 3.3 Trusted Computing Base (TCB)

A Trusted Computing Base é o núcleo mínimo do qual todo o sistema depende para correção. Se a TCB é confiável, tudo construído em cima dela pode ser construído com confiança. O CLI é a TCB do Oraculo — a fundação na qual skills e agentes confiam sem questionar.

**Por que importa:** Skills cuidam do diálogo socrático. Agentes cuidam da investigação de codebase. Ambos dependem do CLI para persistência, validação e operações estruturadas. Se o CLI é confiável, skills e agentes podem focar no trabalho interpretativo sem se preocupar com integridade de dados. Se o CLI não é confiável, nada construído em cima dele pode ser confiável.

O julgamento humano flui pela TCB via dois mecanismos: reviews de versão de documentos (requisitos de epic, definições de story) usam verdicts `approved` e `rejected`; gates de aprovação operacionais (design, execution-plan, qa-escalation) usam verdicts `approved`, `rejected` e `needs_revision`. O CLI é o caminho único e autoritativo pelo qual o julgamento humano entra no sistema, garantindo que nenhuma review ou aprovação possa contornar a validação ou entrar no banco de dados em estado inconsistente.

## 4. Princípios de Design

### 4.1 Distribuição Zero-Dependência

O CLI é um binário estático único construído em Go. Sem runtime, sem gerenciador de pacotes, sem bibliotecas de sistema. Ele roda em qualquer máquina onde o Claude Code roda. Isso é inegociável — se a ferramenta requer setup, agentes não podem depender dela.

Go foi escolhido especificamente para isso: `modernc.org/sqlite` fornece uma implementação SQLite pura em Go sem CGO. O binário compila, copia e roda. Nada mais.

### 4.2 Output Contextual

Comandos servem dois tipos de perguntas e retornam dois tipos de respostas:

- **Comandos operacionais** retornam JSON estruturado. São comandos que reportam status, criam recursos ou gerenciam estado. Agentes e scripts os parseiam de forma confiável.
- **Comandos de conteúdo** retornam conteúdo cru. São comandos que recuperam documentos, requisitos ou conhecimento feito para ser lido. Envolver texto legível por humanos em JSON adiciona ruído sem valor.

O formato de saída corresponde à necessidade do consumidor, não a uma convenção arbitrária.

### 4.3 Auto-Bootstrapping

O CLI nunca requer uma etapa de inicialização separada. O primeiro comando que toca o armazenamento cria o banco de dados, roda migrações e configura o schema automaticamente. `oraculo epic init my-project` em uma codebase nova faz tudo — sem pré-requisito `oraculo init`.

Isso importa porque agentes não conseguem lembrar de forma confiável de rodar comandos de setup. Se a ferramenta requer setup, ela eventualmente será chamada sem ele, e o agente vai desperdiçar um turno em um erro.

### 4.4 Operações Idempotentes

Toda operação de escrita é segura para repetir. `oraculo epic init my-project` chamado duas vezes não falha — retorna `"created": false` e segue em frente. Isso se alinha com como agentes trabalham: eles retentam, re-executam, às vezes esquecem que já fizeram algo. O CLI tolera isso.

### 4.5 Validação Estrita

O CLI valida pré-condições e recusa operações inválidas. Ele não é permissivo — é um guardião da integridade dos dados.

Se um comando requer um epic existente, e o epic não existe, o CLI rejeita a chamada com um erro claro. Ele não cria um placeholder, adivinha a intenção ou prossegue com dados parciais. Input inválido produz um erro, nunca um estado corrompido.

Isso é Design by Contract na prática: o CLI aplica as regras para que os consumidores não precisem.

### 4.6 Falhar Alto, Recuperar Silenciosamente

Erros retornam JSON estruturado com um exit code diferente de zero. O agente vê uma mensagem de erro clara e se adapta. Operações bem-sucedidas retornam resultados estruturados. Não há ambiguidade — o exit code diz ao consumidor se deve prosseguir ou ajustar.

## 5. Independência

O CLI funciona standalone. Ele não depende de estar dentro de uma sessão do Claude Code, de um ambiente de shell específico ou de qualquer contexto de runtime particular. Isso é um princípio filosófico, não uma conveniência prática.

Agentes usam o CLI dentro de skills. Humanos usam no terminal. Pipelines de CI/CD podem usar para automação. O mesmo binário, a mesma interface, os mesmos contratos — independente de quem ou o que está chamando.

Essa independência garante que a Camada de Confiança está sempre disponível. A confiabilidade do CLI não depende da confiabilidade de nada ao redor dele.

## 6. Modelo de Evolução

O CLI é opinativo e versionado. Ele conhece o modelo operacional do Oraculo — as fases, os artefatos, os schemas — e valida contra ele. O CLI não é um data store genérico; ele entende o domínio que serve.

Quando o modelo operacional evolui, o CLI evolui com ele. Mudanças em estruturas de artefatos, novas fases, novas regras de validação — estas se tornam novas versões do CLI com caminhos de migração explícitos.

Não há sistema de plugins ou mecanismo de extensão. O CLI cresce com o Oraculo como uma ferramenta única e coesa. Ele cobre todas as fases centrais: Discover, Plan, Execute e Validate — e os approval gates que as conectam.

## 7. O Que o CLI Não É

O CLI **não é uma ferramenta de gerenciamento de projetos de propósito geral**. Ele não substitui Jira, Linear ou qualquer tracker externo. Ele é a infraestrutura interna do Oraculo — o sistema nervoso que conecta skills, agentes e armazenamento persistente.

O CLI também **não é o ponto de entrada para o fluxo socrático**. Usuários iniciam a descoberta de produto através de skills (`/oraculo:epic`, `/oraculo:story`), que fornecem a experiência guiada e conversacional. O CLI é a infraestrutura na qual skills e agentes dependem — universal em acesso, mas não um substituto para o processo socrático.
