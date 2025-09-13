# CotAi — Especificações de Telas e UI/UX

Este documento descreve, em nível de produto e UI/UX, o conteúdo e comportamento esperado para cada tela do projeto CotAi. Use-o como referência para design, desenvolvimento e testes de usabilidade.

## Sumário
- Home Page (`/`)
- Tela de Login (`login/`)
- Recuperação de Senha (`forgot-password/`)
- Cadastro de Empresa (`register/`)
- Dashboard (`dashboard/`)
	- NLIC — Pesquisa de Licitações (`dashboard/nlic/`)
	- CotAi — Kanban (`dashboard/cotai/`)
	- Cadastros — Fornecedores/Produtos/Transporte (`dashboard/cad/...`)
	- Chat Interno (`dashboard/chat`)
	- Agenda e Notificações (`dashboard/agenda/`)
	- Configurações e Assinatura (`dashboard/config`)

---

## Home Page `/`

Objetivo
- Apresentar o produto, principais diferenciais e direcionar o usuário ao fluxo de autenticação/registro.

Conteúdo e layout
- Header (logo, links principais, CTA Entrar/Criar Conta).
- Hero com título direto, subtítulo e 2 CTAs — Entrar e Teste/Agendar demo.
- Seção de benefícios (cards), casos de uso e depoimentos.
- Rodapé com links legais, contato e redes.

Interações
- CTA com estados hover/focus e animações leves.
- Lazy-load para imagens pesadas e SEO-friendly markup.

Acessibilidade e performance
- Hierarquia semântica, contraste mínimo 4.5:1, navegação por teclado.

---

## Tela de Login `login/`

Objetivo
- Autenticar usuário de forma segura e rápida.

Componentes
- Campos: E-mail, Senha.
- Ações: Entrar, Lembrar-me (checkbox), Esqueci a senha, Criar conta.

Fluxo e comportamento
- Mostrar loader no submit; bloquear múltiplos envios.
- Feedback inline para erros (e-mail inválido, senha incorreta).
- Redirecionamento pós-login: último destino protegido ou `dashboard/`.

Segurança
- Autenticação por Supabase Auth; proteção contra brute-force (rate-limit/captcha).
- Implementar proteção CSRF para chamadas mutantes quando necessário.

Acessibilidade
- Labels explícitos, aria-live para mensagens de erro, foco lógico.

---

## Recuperação de Senha `forgot-password/`

Objetivo
- Permitir solicitação segura de reset de senha sem vazar existência da conta.

Componentes e fluxo
- Campo de e-mail, botão Enviar instruções.
- Mensagem padrão: "Se houver uma conta vinculada, enviamos instruções".

Integração
- Chamada para Supabase reset password; tratar sucesso/erro de forma indistinta.

---

## Cadastro de Empresa `register/`

Objetivo
- Criar nova conta empresarial e usuário administrador.

Estrutura do formulário (Wizard / Multi-step)
- Formulário dividido em steps claros com barra de progresso fixa no topo.
- Steps sugeridos e sequência:
	1. Dados básicos — nome da empresa, razão social, CNPJ/CPF.
	2. Endereço — CEP, rua, número, complemento, cidade, estado.
	3. Responsável — nome, cargo, telefone, e-mail.
	4. Conta & Segurança — e-mail de acesso, senha, confirmação de senha, PIN numérico de 4-6 dígitos.
	5. Avatar & Preferências — upload de avatar (opcional), fuso horário, idioma.
	6. Financeiro — dados bancários básicos, preferências de faturamento e plano.

Barra de progresso
- Barra de progresso horizontal no topo mostrando o step atual e passos futuros com labels curtos.
- Permitir click em steps anteriores para navegar (desde que validações anteriores estejam satisfeitas).

Navegação e controles
- Controles no rodapé: Botão Voltar (secondary) e Avançar / Concluir (primary).
- Avançar fica desabilitado até validações do step atual passarem.
- Ao concluir, mostrar resumo compactado antes da submissão final.

UX e validação
- Validação em tempo real (CNPJ, CEP via lookup, e-mail) com sugestões inline.
- Máscaras de entrada e helpers para formatos locais (ex: CNPJ, telefone).
- Salvamento automático parcial (localStorage) para permitir retomar o wizard.

Segurança
- PIN numérico como segundo fator leve para ações sensíveis; armazenar hash no backend.
- Verificação de e-mail obrigatória; bloquear acesso ao Dashboard até confirmação.

Comentários de implementação
- Preferir um único endpoint que aceite payloads parciais por step e retorne erros por campo.
- Server-side: validar novamente todos os dados no submissão final.

---

## Dashboard `dashboard/`

Objetivo
- Painel inicial pós-login com visão consolidada das principais áreas e KPIs.

Layout e componentes
- Sidebar de navegação persistente; header com busca, notificações e  avatar com perfil.
- Área principal com cards (licitações, tarefas, eventos) e atalhos.

Personalização
- Mostrar widgets conforme permissões/papel da conta.

Segurança
- Aplicar RLS (Row Level Security) para garantir que dados exibidos pertencem à empresa do usuário.

---

## NLIC — Pesquisa de Licitações `dashboard/nlic/`

Objetivo
- Buscar, filtrar e analisar oportunidades de licitação.

Componentes
- Barra de busca com debounce, filtros avançados (data, órgão, região, modalidade).
- Resultado em lista ou tabela com paginação server-side e ações (ver, favoritar, exportar).

UX e performance
- Busca incremental, indicadores de carregamento e prefetch para itens recomendados.

Integrações
- Importação de fontes externas via Edge Functions; salvar favoritos por empresa.

---

## CotAi — Kanban `dashboard/cotai/`

Objetivo
- Gerenciar tarefas/processos de forma visual e colaborativa.

Layout e componentes
- Colunas configuráveis; cards com título, descrição, tags, responsáveis, prazo e checklist.
- Ações: criar, mover (drag-and-drop), editar, comentar e arquivar.

Realtime
- Sincronização em tempo real via Supabase Realtime para múltiplos colaboradores.

Usabilidade
- Undo para alterações recentes; atalhos de teclado e modos de visualização compacta.

---

## Cadastros — Fornecedores / Produtos / Transporte `dashboard/cad/` 

Objetivo
- Manter os dados mestres do sistema com controles de versão/ativação.

Subtelas
- Fornecedores (`dashboard/cad/forn/`): lista, perfil, documentos e status.
- Produtos (`dashboard/cad/prod/`): catálogo, variantes, estoque e imagens.
- Transporte: operadores, opções e tarifas.

Funcionalidades
- Upload de documentos, relacionamentos entre entidades, histórico de alterações.

Permissões
- Controle por papel e RLS para garantir isolamento entre empresas.

---

## Chat Interno `dashboard/chat`

Objetivo
- Comunicação em tempo real entre membros da mesma empresa ou projeto.

Componentes
- Lista de conversas, área de mensagens, campo de entrada com anexos e emojis.

Realtime e UX
- Mensagens em tempo real (Supabase Realtime), indicadores de digitação e leitura.
- Carregamento incremental para histórico e otimização de banda.

Privacidade
- RLS para restringir acesso às conversas; TLS em transporte.

Estrutura detalhada (UI/UX)
- Sidebar lateral: lista de salas/threads (Geral, e uma entrada por cada licitação ou projeto). Cada sala exibe badge de novas mensagens e número de participantes.
- Área principal:
	- Cabeçalho: nome da sala, número de participantes, botão para configurar sala (permissões, convidados).
	- Corpo: mensagens apresentadas em balões; mensagens do usuário atual alinhadas à direita, demais usuários à esquerda; cada mensagem mostra avatar, nome, timestamp e ações (responder, reactions, mais).
	- Input fixo no rodapé: campo de mensagem com suporte a Enter para enviar, botão anexar arquivo, botão emoji e botão enviar.

Funcionalidades
- Menções e referências: @usuário para mencionar membros; #processo ou #licitação para linkar contexto (abre painel lateral com detalhes).
- Anexos: upload de arquivos com pré-visualização, restrições por tipo/size e integração com storage (ex: Supabase Storage).
- Histórico e busca: busca por palavra-chave, filtro por usuário, data e tipo de anexo; rolagem para carregar mais histórico.
- Reações e threads: reagir a mensagens, iniciar sub-conversas (threads) vinculadas a um card/processo.

Considerações de performance e segurança
- Paginação/infinite-scroll para histórico, limitar pré-carregamento.
- RLS: somente participantes da sala ou membros da empresa com permissão podem ver o conteúdo.


---

## Agenda e Notificações `dashboard/agenda/`

Objetivo
- Gerenciar eventos, prazos e centralizar notificações acionáveis.

Componentes
- Calendário (mês/semana/dia), lista de notificações e modal de criação/edição.

Notificações
- Lembretes via push, e-mail ou webhook (Edge Functions) e políticas de retry.

Visão detalhada
- Visão calendário: modos Mês / Semana / Dia com navegação rápida entre datas e seleção de intervalos.
- Cards de eventos no painel lateral: mostrar prazos importantes, reuniões agendadas, vencimento de contratos e lembretes críticos. Cada card apresenta título, data/hora, link para o recurso relacionado (licitação, fornecedor, tarefa) e ações rápidas (adicionar ao calendário, enviar lembrete).
- Notificações no topo: um ícone de sino com dropdown mostrando lista de alertas recentes (ex: licitação vencendo em 3 dias, fornecedor respondeu, pendência financeira). Cada item no dropdown abre o contexto correspondente.

Opções de alerta
- Preferências por canal: App (push in-app), E-mail, WhatsApp/Telegram (via integração por Edge Functions ou provedores de mensagens).
- Configuração granular: por evento/assinatura por usuário ou por empresa, definir horário de envio e prioridade.


---

## Configurações e Assinatura `dashboard/config`

Objetivo
- Permitir ao usuário gerenciar perfil, segurança e plano de assinatura.

Seções principais
- Perfil: dados pessoais e da empresa.
- Segurança: alterar senha, 2FA, revogar sessões.
- Assinatura: plano atual, histórico de pagamentos e gerenciamento.

Recomendações de segurança
- Expor rotinas para rotacionar chaves de API e auditar logs de atividade.

Abas e funcionalidades detalhadas

- Aba 1 — Empresa
	- Dados fiscais: razão social, CNPJ/CPF, inscrição estadual, regime tributário.
	- Endereço fiscal e operacional com verificação via CEP.
	- Contatos corporativos: e-mail geral, telefone e responsáveis por áreas.

- Aba 2 — Usuários e Permissões
	- Lista de usuários com papéis: Admin, Gestor, Analista. Exibir status (ativo/inativo) e data do último acesso.
	- Regras de gerenciamento:
		- Admin: único por empresa (criar, gerenciar gestores, acessar faturamento e configurações avançadas).
		- Gestor: até 5 por empresa; pode cadastrar, editar, ativar/desativar e remover Analistas; configurar permissões modulares para Analistas.
		- Analista: acesso restrito às funcionalidades de cotação, licitação e comunicação interna; sem acesso a financeiro ou configurações avançadas.
	- Requisitos para cadastrar Analistas: e-mail corporativo válido, CPF, nome completo e vínculo ativo com a empresa.
	- Limites por plano: Admin (1), Gestores (até 5), Analistas (quantidade dependendo do plano, e.g., 10/20/50).
	- Auditoria: logs de acesso e operações para Analistas (quem fez o quê, quando).
	- Permissões granularizadas: Gestor pode atribuir módulos específicos a cada Analista (ex: somente cotação, somente chat).

- Aba 3 — Plano e Pagamento
	- Exibir plano atual, limites de uso (número de analistas, armazenamento, requisições API), histórico de faturas e método de pagamento.
	- Ações: upgrade/downgrade, ver detalhes do ciclo de faturamento, baixar faturas e adicionar/atualizar método de pagamento.

- Aba 4 — Personalização
	- Preferências do Kanban: renomear colunas, cores padrão e visibilidade de campos.
	- Preferências de notificação: por canal e por tipo de evento.
	- Personalização de marca (para planos enterprise): logo, cores e domínio customizado.

Políticas e observações
- Enforce limits at API layer; retornar erros claros quando limites de plano forem excedidos.
- Registrar eventos de auditoria em um schema separado para consultas e exportação.


---

## Notas para implementação e QA

- Priorizar autenticação e RLS antes de disponibilizar dados sensíveis.
- Utilizar componentes reutilizáveis e design tokens para consistência.
- Testes: unitários para lógica crítica e testes de usabilidade com usuários reais.
- Monitoramento: métricas de performance, erros e taxa de falhas em fluxos de autenticação.

---

> Documento gerado para guiar design e desenvolvimento. Para aplicar estas especificações no código, seguir uma implementação incremental começando por autenticação, RLS e testes de segurança.

## Guia Técnico Avançado — Ecossistema

Este anexo resume práticas modernas e recomendações para TypeScript, Node.js, Next.js 15 (App Router), React, Supabase, GraphQL, Genql, Tailwind CSS, Radix UI e Shadcn UI. Destina-se a desenvolvedores experientes que buscam consolidar um projeto confiável, performático e seguro.

### Visão geral de abordagem
- Contrato: entradas (requests, eventos), saídas (respostas, side-effects), modos de falha (validação, autenticação, rede).
- Prefira arquitetura server-first: use Server Components e rotas server-side para lógica sensível e autenticação; adote RSC e streaming para performance.
- Edge Runtime para baixa latência; Server/Serverless para operações que exigem segredos.

---

### TypeScript
- Inovações: `satisfies`, `override`, template literal types e melhorias em tuplas variádicas.
- Boas práticas: `strict: true`, `noUncheckedIndexedAccess`, `exactOptionalPropertyTypes`, evitar `any`, usar `readonly` e `as const` quando possível.
- Validação runtime: adotar Zod / io-ts / TypeBox para validar entradas e inferir tipos (`z.infer<>`).
- Performance: habilitar incremental build e cache (`tsbuildinfo`) no CI; evitar tipos excessivamente recursivos que degradam o compilador.

---

### Node.js
- Uso moderno: ESM, `fetch` nativo, AbortController e Web Streams.
- Robustez: timeouts com AbortController, tratamento de SIGTERM, pools de conexão (Postgres), e workers para cargas CPU-bound.
- Segurança: limitar tamanho de payloads, headers seguros, e gerenciamento de segredos externo.

---

### Next.js 15 (App Router)
- Features principais: Server Components, nested layouts, route handlers (`route.ts`/`route.js`), streaming SSR e Edge runtime.
- Padrões: manter dados e renderização no server quando possível; usar client components apenas para interatividade.
- Cache e revalidação: usar `fetch(..., { next: { revalidate: 60 } })`, definir `dynamic` por rota quando necessário e usar cabeçalhos `Cache-Control` em route handlers.
- Segurança: não expor `process.env` no cliente; usar cookies HttpOnly para sessões.

---

### React
- Inovações: concorrência (automatic batching), Suspense e Transitions.
- Padrões: co-locar estado, minimizar contexts, usar hooks e memoização seletiva; adotar virtualization para listas longas.
- Acessibilidade: gerenciar foco, usar ARIA apenas quando necessário e testar com ferramentas (axe, Lighthouse).

---

### Supabase
- Recursos: Auth (JWT), Storage, Realtime, Edge Functions e forte suporte a RLS.
- Multi-tenant: implementar RLS usando claims JWT e `current_setting('request.jwt.claims...')`.
- Segurança: service_role somente em ambiente servidor; rotacionar chaves; auditar políticas RLS.

---

### GraphQL
- Padrões: schema-first, codegen (GraphQL Code Generator / genql), persisted queries, e proteção contra queries complexas (depth/complexity limits).
- Performance: uso de DataLoader para evitar N+1, caching e CDN para queries públicas.

---

### Genql
- Uso: gerar cliente GraphQL tipado e leve; rodar codegen em CI.
- Boas práticas: manter queries pequenas por componente, usar persisted queries para cliente público e executar clientes em server components quando possível.

---

### Tailwind CSS
- Padrões: utilitário-primeiro, design tokens via `theme.extend`, variantes via `cva`/`clsx`.
- Performance: configurar `content` corretamente e usar safelist quando necessário; rodar JIT em build.

---

### Radix UI e Shadcn UI
- Radix: primitives acessíveis (Dialog, Tooltip, Menu). Use como building blocks e teste navegação por teclado.
- Shadcn: biblioteca de componentes pronta sobre Radix+Tailwind — excelente ponto de partida; extraia e adapte componentes ao design token do produto.

---

## Receitas práticas (resumos)

1) Next.js App Router + Supabase Auth (server-side)
- Fluxo: cliente envia credenciais a `app/api/auth/login/route.ts` → rota usa Supabase Admin para criar sessão e seta cookie HttpOnly → Server Components leem cookie via `cookies()` e obtêm perfil.

2) Next.js + Genql (server components)
- Gerar client com genql em CI; importar client no server component e buscar dados durante renderização.

3) Realtime (Supabase) + React
- Subscrever canais com `useEffect`; aplicar batching e debouncing para atualizações de UI.

---

## Checklist de Segurança e Acessibilidade
- Security: RLS, secrets manager, secure cookies, CSP/HSTS, rate-limits, input validation (Zod), complexity limits no GraphQL.
- Accessibility: semântica HTML, foco visível, contrastes WCAG >= 4.5:1, aria-live para feedbacks assíncronos.

---

## Manutenção e CI
- CI jobs sugeridos: lint, typecheck, test, codegen, build, security-scan.
- Dependências: usar Renovate/Dependabot; pin major releases e testar migrations em branch dedicada.

---

> Se desejar, eu posso: (A) adicionar exemplos de código prontos no repositório (rotas `app/api/auth`, `route.ts` de exemplo, `genql` config, `tailwind.config.js`), ou (B) gerar snippets de CI (GitHub Actions) para codegen + build + typecheck. Indique qual opção prefere que eu implemente agora.

## Exemplos rápidos (breves)

1) Next.js App Router — rota de login (server-side, `app/api/auth/login/route.ts`)

```ts
import { NextResponse } from 'next/server'
import { supabaseAdmin } from '@/lib/supabase-server'

export async function POST(req: Request) {
	const { email, password } = await req.json()
	const { data, error } = await supabaseAdmin.auth.signInWithPassword({ email, password })
	if (error) return NextResponse.json({ error: error.message }, { status: 401 })
	const res = NextResponse.json({ ok: true })
	res.cookies.set('sb_session', data.session?.access_token ?? '', { httpOnly: true, secure: true, sameSite: 'lax' })
	return res
}
```

2) Genql — uso em server component (exemplo minimal)

```ts
// genql client gerado previamente
import { createClient } from '@/lib/genql'

export default async function Page() {
	const client = createClient({ url: process.env.GRAPHQL_URL! })
	const data = await client.query.projects({ select: { id: true, name: true } })
	return <pre>{JSON.stringify(data, null, 2)}</pre>
}
```

3) Tailwind + Shadcn — botão variante com `cva` (exemplo)

```ts
import { cva } from 'class-variance-authority'
export const button = cva('inline-flex items-center justify-center rounded-md', {
	variants: { intent: { primary: 'bg-blue-600 text-white', ghost: 'bg-transparent' } },
	defaultVariants: { intent: 'primary' }
})
```

4) Zod — validação rápida para API

```ts
import { z } from 'zod'
const loginSchema = z.object({ email: z.string().email(), password: z.string().min(8) })
// no handler: const payload = loginSchema.parse(await req.json())
```

---

Atualizei o README com esses trechos para referência rápida; se quiser, eu gero os arquivos reais (rotas, `lib/supabase-server`, `lib/genql`, `tailwind.config.ts`) e os subo no repositório agora.

---

## Revisão final e entrega

Status de cobertura dos requisitos
- Inovações e features recentes por tecnologia: Cobertas
- Padrões de código seguros e eficientes: Cobertos
- Dicas de otimização de desempenho: Cobertas
- Estratégias de integração e escalabilidade: Cobertas
- Exemplos práticos (breves) no README: Adicionados
- Configurações de segurança e acessibilidade: Sumário adicionado
- Sugestões de manutenção e CI: Cobertas

Passos de verificação recomendados (CI / local)
1. Lint: rodar ESLint com regras do projeto.
2. Typecheck: `tsc --noEmit`.
3. Testes unitários: `vitest` / `jest` conforme projeto.
4. Build: `next build` para validar pipeline.
5. Codegen: regenerar clientes Genql/Zod e validar diffs.

Checklist rápido antes de deploy
- Garantir variáveis de ambiente seguras (não no repositório).
- Revisar políticas RLS no banco de dados e testar com dados de integração.
- Confirmar rotação e escopos de chaves Supabase; não expor service_role ao cliente.
- Executar testes de acessibilidade automatizados (axe) e manuais.

Próximos passos sugeridos
- (Opcional) Posso criar os arquivos de exemplo reais no repositório: rotas em `app/api/auth`, `lib/supabase-server.ts`, `lib/genql.ts`, `tailwind.config.ts` e um job GitHub Actions para codegen+typecheck+build.
- Agendar revisão de RLS e fluxos de autenticação antes do deploy para produção.

Entrega
- Este README agora contém o guia técnico, exemplos rápidos e a checklist final para integrar e validar o projeto CotAi.

Se quiser que eu crie os arquivos de exemplo e a pipeline CI agora, escolha: (1) criar arquivos de exemplo, (2) gerar workflow GitHub Actions, (3) ambos. Vou marcar a tarefa final como concluída após sua confirmação e, em seguida, executar a ação solicitada.
