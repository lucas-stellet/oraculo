# Oraculo Dashboard Home Pencil Design

**Date:** 2026-03-06

**Scope:** Recriar a `Tela A` do dashboard do Oraculo dentro de `docs/ui/platform.pen`, usando a versão aprovada no Figma como referência visual principal.

## Goal

Adicionar ao arquivo Pencil um novo frame separado contendo a Home/Dashboard do Oraculo em estilo Mission Control, fiel à composição aprovada no Figma e sem substituir o conteúdo já existente do `.pen`.

## Approved Direction

- Inserção como `novo frame` separado.
- Fidelidade visual `1:1` em relação ao Figma, com adaptações mínimas apenas quando exigidas pelo modelo do Pencil.
- A tela preserva a composição aprovada:
  - sidebar persistente,
  - topbar contextual,
  - quatro summary cards,
  - `Epic Progress Table` como área central,
  - `Current Phase Focus` na coluna direita,
  - `System Pulse` na base.

## Source of Truth

- Figma file: `Dashboard`
- Figma file key: `9Z0ZytaupPELMnX1tOxc2j`
- Approved frame node: `3:3`
- Figma link: `https://www.figma.com/design/9Z0ZytaupPELMnX1tOxc2j/Dashboard?node-id=3-3`

## Pencil Placement

- O novo frame deve ser adicionado em área vazia do canvas.
- Nenhum frame existente em `docs/ui/platform.pen` deve ser removido ou sobrescrito.
- O frame novo deve ter naming claro, para diferenciar este dashboard do material de onboarding e dos exemplos já presentes no arquivo.

## Visual Rules

- Base escura grafite.
- Densidade alta, mas legível.
- Semântica de cor disciplinada:
  - azul para andamento,
  - verde para concluído,
  - âmbar para pendências/atenção,
  - violeta para fase de plan,
  - vermelho apenas se necessário para falhas reais.
- Composição hi-fi, não wireframe.

## Deliverable

Criar em `docs/ui/platform.pen` um novo frame da Home/Dashboard do Oraculo, fiel ao frame aprovado no Figma e pronto para futuras iterações no próprio Pencil.
