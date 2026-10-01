# Contributing

Obrigado por contribuir com o JobSearcher.

## Antes de abrir um PR

1. Rode `go test ./...`.
2. Mantenha o comportamento global e exclusivamente remoto do produto.
3. Não adicione upload de currículo, armazenamento de dados pessoais ou candidatura automática sem uma decisão explícita do projeto.
4. Evite scraping frágil quando existir uma API ou feed público apropriado.
5. Adicione testes para novas regras de matching, filtros ou fontes.
6. Não inclua tokens, credenciais, cookies, dados pessoais ou arquivos locais no commit.

## Pull requests

Descreva o problema, a mudança, os testes executados e eventuais limitações.

## Novas fontes

Uma nova fonte deve implementar `sources.JobSource`, respeitar `context.Context`, limitar respostas, normalizar os campos básicos e ter testes para sucesso e erro.

O JobSearcher encontra e compara vagas. Ele não envia candidaturas nem mensagens para recrutadores.
