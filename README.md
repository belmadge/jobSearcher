# JobSearcher

Automação para buscar e comparar vagas de tecnologia compatíveis com um perfil configurado, priorizando vagas **100% remotas** em qualquer país ou região.

## Interface web

A primeira versão do JobSearcher pode ser usada por uma interface local, sem upload de currículo e sem login.

```bash
go run ./cmd/jobsearch web
```

Depois abra `http://localhost:8080`.

A configuração é feita durante a sessão:

- cargo / área desejada;
- skills;
- senioridade;
- anos de experiência.

A busca é sempre **global e exclusivamente remota**.

O perfil informado no formulário é mantido apenas em memória durante a busca. O JobSearcher não precisa armazenar currículo, nome, telefone, e-mail ou outros dados pessoais para realizar o matching.

A interface usa o mesmo pipeline de fontes, filtros, matching, score e deduplicação do CLI. A aplicação apenas encontra e apresenta vagas; não envia candidaturas nem mensagens para recrutadores.

## Como usar

### 1. Pré-requisitos

- Go instalado.
- Repositório clonado localmente.
- Terminal aberto na pasta do projeto.

Confira a versão do Go:

```bash
go version
```

### 2. Rodar os testes

Sempre que houver alteração no projeto, rode:

```bash
go test ./...
```

Esse comando executa os testes de todos os pacotes do projeto.

### 3. Fazer uma busca de vagas

Para executar a busca configurada no CLI:

```bash
go run ./cmd/jobsearch run
```

A busca consulta as fontes disponíveis, aplica os filtros de perfil, cargo, senioridade, compatibilidade e atualização das vagas e gera os resultados.

### 4. Visualizar o último relatório

Depois de executar a busca:

```bash
go run ./cmd/jobsearch report
```

O relatório também fica salvo em:

```text
reports/YYYY-MM-DD.md
reports/YYYY-MM-DD.json
```

### 5. Rodar uma busca sem salvar no banco

Para executar em modo de teste:

```bash
go run ./cmd/jobsearch run --dry-run
```

### 6. Executar apenas uma fonte

É possível limitar a busca a uma fonte:

```bash
go run ./cmd/jobsearch run --source=greenhouse
go run ./cmd/jobsearch run --source=lever
go run ./cmd/jobsearch run --source=remoteok
go run ./cmd/jobsearch run --source=remotive
go run ./cmd/jobsearch run --source=programathor
go run ./cmd/jobsearch run --source=himalayas
```

Para executar todas as fontes:

```bash
go run ./cmd/jobsearch run --source=all
```

### 7. Ver estatísticas do banco

```bash
go run ./cmd/jobsearch stats
```

### 8. Fluxo recomendado

```bash
go test ./...
go run ./cmd/jobsearch run
go run ./cmd/jobsearch report
```

Se os testes falharem, corrija os erros antes de executar a busca.

## Configuração

Os principais arquivos de configuração são:

- `config/profile.json` — perfil profissional usado pelo CLI: cargos, tecnologias, experiência e pesos.
- `config/search.json` — score mínimo, período de atualização e limite por fonte.
- `config/boards.json` — empresas/fontes Greenhouse e Lever habilitadas.
- `.env.example` — variáveis de ambiente opcionais.

### Regras principais da busca

O JobSearcher:

- busca globalmente;
- aceita somente vagas classificadas como **remotas**;
- aplica filtros de cargo e aliases relacionados;
- compara skills exatas e tecnologias relacionadas;
- considera responsabilidades, senioridade e experiência;
- calcula um score de compatibilidade;
- classifica os resultados em faixas de compatibilidade;
- explica os principais sinais usados no matching;
- deduplica resultados entre fontes;
- limita a quantidade de resultados por fonte após o ranking;
- não realiza candidatura automática;
- não envia mensagens para recrutadores.

A configuração de cargo, skills, senioridade e experiência é feita pelo usuário. O projeto não exige upload de currículo.

## Matching

O score combina diferentes sinais do perfil e da vaga, incluindo:

- **Cargo** — proximidade entre o cargo desejado e o título da vaga.
- **Skills** — tecnologias explicitamente encontradas.
- **Tecnologias relacionadas** — tecnologias equivalentes ou próximas dentro de uma mesma categoria.
- **Responsabilidades** — alinhamento entre as atividades descritas e a área procurada.
- **Senioridade** — compatibilidade entre a senioridade informada e a vaga.
- **Experiência** — comparação entre os anos informados e os requisitos explícitos da vaga.
- **Cloud, domínio, linguagem e IA** — conforme os pesos configurados no perfil.

As faixas de compatibilidade são:

- **Forte** — 75% a 100%.
- **Compatível** — 60% a 74%.
- **Possível** — 40% a 59%.
- **Baixa** — abaixo de 40%.

Os resultados também podem apresentar os principais motivos do match e requisitos ausentes identificados.

## Relatórios

Cada vaga pode apresentar:

- título;
- empresa;
- localização;
- modelo de trabalho;
- score;
- compatibilidade;
- cargo e skills;
- senioridade;
- experiência;
- responsabilidades;
- destaques do matching;
- requisitos encontrados e ausentes;
- fonte;
- link direto para a vaga.

## Fontes de vagas

Atualmente o JobSearcher usa fontes com endpoints públicos ou APIs documentadas:

- Greenhouse — boards públicos de empresas configuradas.
- Lever — postings públicos de empresas configuradas.
- Remote OK — feed público de vagas remotas.
- Remotive — API pública de vagas remotas.
- Programathor — página pública brasileira de vagas de programação.
- Himalayas — API pública de vagas remotas, sem autenticação.

### Fontes avaliadas, mas não acopladas diretamente

**LinkedIn:** as páginas de vagas podem ser públicas, mas o projeto não depende de scraping direto do LinkedIn.

**Gupy:** possui API documentada, mas o fluxo de consulta exige credenciais específicas. O projeto não exige uma credencial de empresa/recrutador apenas para pesquisar vagas.

**Indeed:** possui APIs e regras específicas de integração; não será usado como scraper de HTML.

O objetivo é ampliar a cobertura sem transformar o JobSearcher em um robô frágil ou dependente de credenciais de terceiros.

## Comandos rápidos

| Objetivo | Comando |
|---|---|
| Testar projeto | `go test ./...` |
| Interface web | `go run ./cmd/jobsearch web` |
| Buscar vagas | `go run ./cmd/jobsearch run` |
| Ver relatório | `go run ./cmd/jobsearch report` |
| Busca sem salvar | `go run ./cmd/jobsearch run --dry-run` |
| Ver estatísticas | `go run ./cmd/jobsearch stats` |
| Ver ajuda | `go run ./cmd/jobsearch help` |

## Importante

O JobSearcher é uma ferramenta de pesquisa e triagem. Ele **não se candidata automaticamente às vagas** e **não envia mensagens para recrutadores**.

A configuração usada pelo CLI fica nos arquivos locais de configuração. A interface web, por sua vez, permite informar o perfil durante a sessão sem exigir upload ou armazenamento de currículo.
