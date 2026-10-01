# Security

## Reportando uma vulnerabilidade

Não publique detalhes de uma vulnerabilidade de segurança em uma issue pública. Abra um contato privado com os mantenedores antes da divulgação.

Inclua descrição, passos para reproduzir, impacto, versão/commit afetado e sugestão de correção, se houver.

Nunca envie senhas, tokens, cookies, chaves privadas ou outros segredos no relatório.

## Dados e credenciais

O JobSearcher não exige upload de currículo ou login para a busca web. A configuração da interface web fica em memória durante a sessão.

Variáveis de ambiente, tokens e credenciais locais nunca devem ser commitados. Use arquivos locais ignorados pelo Git.

## Dependências

Mantenha dependências no `go.mod` e `go.sum`. Integrações externas devem usar HTTPS e respeitar os contratos das fontes.
