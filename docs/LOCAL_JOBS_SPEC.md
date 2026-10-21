# Especificação — Módulo de Empregos Locais

## Objetivo

Adicionar ao JobSearcher um módulo independente para localizar vagas não necessariamente ligadas à tecnologia. O módulo deve coexistir com o módulo atual de Tecnologia Remota, sem alterar suas regras de busca.

Esta especificação define o comportamento inicial. **Não executa buscas de vagas e não configura nenhuma fonte externa nesta etapa.**

## Configuração inicial de exemplo

| Campo | Valor inicial |
|---|---|
| Cargo desejado | Vendedora |
| Área profissional | Vendas e comércio |
| Experiência | Sem preferência |
| Escolaridade | Qualquer |
| Cidade | Maceió |
| Estado | Alagoas (AL) |
| País | Brasil |
| Modalidade | Presencial |
| Remoto/híbrido | Não aceitar |
| Persistência do perfil | Não persistir; manter apenas durante a requisição/sessão |

Esses valores são defaults de exemplo para desenvolver e testar o módulo, não um perfil pessoal salvo.

## Separação dos módulos

- **Tecnologia Remota:** mantém o comportamento existente, com busca global e filtro de vagas remotas compatíveis com a localização do candidato.
- **Empregos Locais:** usa localização fixa em Maceió/AL, aceita apenas vagas presenciais e permite cargos e áreas não técnicas.

Os módulos podem compartilhar clientes HTTP, normalização de vagas, deduplicação, ranking e componentes de interface, mas devem ter configurações e políticas de elegibilidade distintas. Nenhuma mudança no módulo local deve relaxar os filtros do módulo remoto, e vice-versa.

## Modelo de configuração proposto

A configuração do módulo local deve ser independente da configuração atual de tecnologia. Estrutura conceitual sugerida:

```go
type LocalJobsProfile struct {
    Title        string
    Category     string
    Experience   string // "any", "none", "required"
    Education    string // "any", "elementary", "high_school", "higher"
    City         string
    State         string
    Country       string
    WorkplaceType string // somente "onsite" nesta primeira versão
}
```

Defaults de exemplo:

```go
LocalJobsProfile{
    Title:         "Vendedora",
    Category:      "Vendas e comércio",
    Experience:    "any",
    Education:     "any",
    City:          "Maceió",
    State:         "AL",
    Country:       "Brasil",
    WorkplaceType: "onsite",
}
```

A estrutura acima é uma proposta de design; os tipos e nomes finais podem ser adaptados às convenções existentes no repositório durante a implementação.

## Regras de busca e elegibilidade

1. **Localização:** aceitar somente vagas cuja localização identificada seja Maceió/AL. Vagas nacionais, de outras cidades ou com localização desconhecida devem ser excluídas ou marcadas como não elegíveis — nunca presumidas como locais.
2. **Modalidade:** aceitar somente vagas presenciais confirmadas. Rejeitar remoto, híbrido e modalidade desconhecida.
3. **Cargo e categoria:** pesquisar o cargo digitado e sinônimos relevantes em português. Para "Vendedora", considerar variantes como "vendedor", "consultora de vendas", "consultor de vendas" e "atendente de loja" como termos de descoberta, mas o matching deve explicar quando o título é apenas relacionado, não equivalente.
4. **Experiência "Sem preferência":** não excluir vaga por exigir ou não experiência. Se a vaga informar esse requisito, mostrá-lo como informação; não aplicar penalidade de elegibilidade.
5. **Escolaridade "Qualquer":** não excluir vagas por escolaridade. Se a vaga informar escolaridade, apresentá-la como requisito.
6. **Informações ausentes:** não inventar modalidade, cidade, escolaridade, salário ou experiência. Mostrar "não informado" quando pertinente.
7. **Deduplicação:** remover vagas repetidas entre fontes antes de exibir os resultados, mantendo a melhor representação disponível.
8. **Ranking:** priorizar compatibilidade de cargo/categoria e requisitos informados. Não usar skills de programação ou pesos específicos de tecnologia para pontuar vagas locais.
9. **Links:** cada resultado deve apontar para a publicação original da vaga.
10. **Privacidade:** não solicitar currículo, nome, telefone, e-mail ou outros dados pessoais. Não salvar o perfil informado pelo formulário.

## Taxonomia inicial de categorias

A lista deve ser extensível e editável. Categorias iniciais sugeridas:

- Saúde e clínicas: recepcionista de clínica, atendente de clínica, auxiliar administrativo de clínica.
- Atendimento ao público: atendente, recepcionista, operador(a) de atendimento.
- Vendas e comércio: vendedor(a), consultor(a) de vendas, atendente de loja, operador(a) de caixa.
- Administrativo: auxiliar administrativo, assistente administrativo.
- Hotelaria e serviços: recepção, atendimento e funções de serviços compatíveis com a categoria escolhida.

Sinônimos servem para ampliar a descoberta, não para afirmar que duas ocupações são idênticas. A interface deve preservar o cargo desejado e explicar quando um resultado corresponde a um cargo relacionado.

## Fontes de vagas

Cada fonte precisa ser avaliada individualmente antes de ser ativada:

- Preferir APIs oficiais, feeds públicos e páginas de carreiras cuja coleta automatizada seja permitida.
- Respeitar termos de uso, limites de requisição, robots.txt quando aplicável, atribuição exigida e regras de licenciamento.
- Não contornar autenticação, CAPTCHA, paywall ou controles anti-bot.
- Não tratar páginas públicas do Indeed ou Glassdoor como autorização automática para scraping. Usar integração oficial somente se houver acesso e permissão adequados.
- Permitir que uma fonte falhe sem descartar resultados obtidos das outras fontes.
- Registrar no resultado a fonte original e a URL da vaga.

A lista concreta de fontes brasileiras deve ser validada por categoria e cobertura real de Maceió antes de prometer disponibilidade de vagas.

## Interface proposta

A tela inicial deve permitir escolher entre dois módulos:

1. **Tecnologia Remota** — interface e filtros atuais.
2. **Empregos Locais** — formulário com cargo, categoria, experiência, escolaridade e local fixo Maceió/AL; modalidade fixa em presencial.

No módulo local, exibir de forma visível: "Maceió, Alagoas · Presencial". A modalidade e a cidade não devem ser controles que possam ser alterados acidentalmente nesta primeira versão.

A seção de resultados deve aparecer apenas depois que o usuário iniciar uma busca. Antes disso, mostrar apenas a introdução e o formulário.

## Critérios de aceite

- [ ] O módulo local pode ser aberto sem alterar o módulo de Tecnologia Remota.
- [ ] O formulário começa com os valores de exemplo documentados acima, sem persistir o perfil após a busca.
- [ ] Uma vaga presencial em Maceió pode ser elegível se atender aos demais critérios.
- [ ] Uma vaga em Arapiraca, Recife ou outra cidade é rejeitada.
- [ ] Vagas remotas, híbridas ou com modalidade desconhecida são rejeitadas.
- [ ] Experiência "Sem preferência" não elimina vagas por requisito de experiência.
- [ ] Escolaridade "Qualquer" não elimina vagas por requisito de escolaridade.
- [ ] Títulos relacionados podem ser encontrados, mas a explicação do matching não os apresenta como equivalência exata.
- [ ] A falha de uma fonte não impede a exibição dos resultados das demais.
- [ ] Duplicatas são eliminadas e os links levam à publicação original.
- [ ] O módulo local não usa os pesos de matching de tecnologia.
- [ ] Testes unitários cobrem cidade, modalidade, experiência, escolaridade, aliases e isolamento entre módulos.
- [ ] `go test ./...` passa.

## Fora do escopo desta especificação

- Executar uma busca real agora.
- Candidatar-se automaticamente.
- Enviar mensagens a recrutadores ou empregadores.
- Fazer upload ou análise de currículo.
- Persistir dados pessoais do candidato.
- Garantir que haverá vagas disponíveis para todos os cargos ou fontes.
