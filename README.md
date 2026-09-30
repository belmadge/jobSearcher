# JobSearcher

Automação para buscar vagas de tecnologia compatíveis com o perfil configurado, priorizando vagas remotas elegíveis para o Brasil e vagas presenciais/híbridas em Maceió/Alagoas.

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

Para executar a busca completa:

```bash
go run ./cmd/jobsearch run
```

A busca consulta as fontes configuradas, aplica os filtros de perfil, senioridade, localização e atualização das vagas e gera os resultados.

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

### 8. Fluxo recomendado no dia a dia

O fluxo mais simples é:

```bash
go test ./...
go run ./cmd/jobsearch run
go run ./cmd/jobsearch report
```

Se os testes falharem, corrija os erros antes de executar a busca.

## Configuração

Os principais arquivos de configuração são:

- `config/profile.json` — perfil profissional, tecnologias, experiência e pesos de compatibilidade.
- `config/search.json` — localização, títulos procurados, score mínimo e período de atualização.
- `config/boards.json` — empresas/fontes Greenhouse e Lever habilitadas.
- `.env.example` — variáveis de ambiente opcionais.

### Regras principais da busca

O JobSearcher:

- prioriza vagas recentes;
- aceita vagas remotas quando a descrição permite trabalho a partir do Brasil/região elegível;
- aceita vagas presenciais ou híbridas em Maceió/Alagoas;
- rejeita vagas presenciais/híbridas fora de Maceió/Alagoas;
- avalia compatibilidade técnica e de senioridade;
- prioriza Software Engineer I e considera Software Engineer II como alvo secundário;
- evita funções fora do perfil;
- não realiza candidatura automática;
- não envia mensagens para recrutadores.

## Relatórios

Os resultados são separados em:

- **Alta compatibilidade** — score alto.
- **Compatibilidade possível** — atende ao score mínimo, mas possui mais lacunas.
- **Para analisar** — localização ainda não pôde ser determinada com segurança.

Cada vaga pode apresentar título, empresa, localização, modelo de trabalho, score, senioridade, motivos de compatibilidade, gaps e link direto.

## Comandos rápidos

| Objetivo | Comando |
|---|---|
| Testar projeto | `go test ./...` |
| Buscar vagas | `go run ./cmd/jobsearch run` |
| Ver relatório | `go run ./cmd/jobsearch report` |
| Busca sem salvar | `go run ./cmd/jobsearch run --dry-run` |
| Ver estatísticas | `go run ./cmd/jobsearch stats` |
| Ver ajuda | `go run ./cmd/jobsearch help` |

## Importante

O JobSearcher é uma ferramenta de pesquisa e triagem. Ele **não se candidata automaticamente às vagas** e **não envia mensagens para recrutadores**.


## Fontes de vagas

Atualmente o JobSearcher usa fontes com endpoints públicos ou APIs documentadas:

- Greenhouse — boards públicos de empresas configuradas.
- Lever — postings públicos de empresas configuradas.
- Remote OK — feed público remoto.
- Remotive — API pública de vagas remotas.
- Programathor — página pública brasileira de vagas de programação.
- Himalayas — API pública de vagas remotas, sem autenticação.

A busca não faz candidatura automática.

### Fontes avaliadas, mas não acopladas diretamente

**LinkedIn:** as páginas de vagas podem ser públicas, mas não foi encontrada uma API pública oficial de busca de vagas adequada para esta automação. Por isso, o projeto não depende de scraping direto do LinkedIn.

**Gupy:** possui API pública documentada para consulta de vagas, porém o fluxo documentado para consumir as vagas exige um Bearer Token gerado pela plataforma. Não vou exigir uma credencial de empresa/recrutador apenas para pesquisar vagas.

**Indeed:** a plataforma possui APIs e regras específicas de integração; não será usada como scraper de HTML.

O objetivo é ampliar a cobertura sem transformar o JobSearcher em um robô frágil ou dependente de credenciais de terceiros.


## Interface web

A primeira versão do JobSearcher também pode ser usada por uma interface local, sem upload de currículo e sem login.

    go run ./cmd/jobsearch web

Depois abra `http://localhost:8080`.

A configuração é feita durante a sessão: cargo/área desejada; skills; senioridade; anos de experiência; localização; e modelo de trabalho (remoto, híbrido ou presencial).

O perfil informado no formulário é mantido apenas em memória durante a busca. O JobSearcher não precisa armazenar currículo, nome, telefone, e-mail ou outros dados pessoais para realizar o matching.

O formulário usa o mesmo pipeline existente de fontes, filtros, localização, senioridade, matching, score e deduplicação do CLI. A aplicação continua apenas encontrando e apresentando vagas; não envia candidaturas nem mensagens para recrutadores.
