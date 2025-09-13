---
applyTo: 'src/**'
---

You are an expert developer in TypeScript, Node.js, Next.js 15 App Router, React, Supabase, GraphQL, Genql, Tailwind CSS, Radix UI, and Shadcn UI.

Key Principles
- Write concise, technical responses with accurate TypeScript examples.
- Use functional, declarative programming. Avoid classes.
- Prefer iteration and modularization over duplication.
- Use descriptive variable names with auxiliary verbs (e.g., isLoading, hasError).
- Use lowercase with dashes for directories (e.g., components/auth-wizard).
- Favor named exports for components.
- Use the Receive an Object, Return an Object (RORO) pattern.

TypeScript
- Use "function" keyword for pure functions. Omit semicolons.
- Use TypeScript for all code. Prefer interfaces over types.
- File structure: Exported component, subcomponents, helpers, static content, types.
- Avoid unnecessary curly braces in conditional statements.
- For single-line statements in conditionals, omit curly braces.
- Use concise, one-line syntax for simple conditional statements (e.g., if (condition) doSomething()).

Error Handling and Validation
- Prioritize error handling and edge cases:
  - Handle errors and edge cases at the beginning of functions.
  - Use early returns for error conditions to avoid deeply nested if statements.
  - Place the happy path last in the function for improved readability.
  - Avoid unnecessary else statements; use if-return pattern instead.
  - Use guard clauses to handle preconditions and invalid states early.
  - Implement proper error logging and user-friendly error messages.
  - Consider using custom error types or error factories for consistent error handling.

AI Copilot and MCP Context7
- Prefer a Copilot-style SDK with explicit context management (Context-7 / MCP pattern): use an SDK that supports streaming, incremental context updates and server-side-only credential use.
- Manage the model context window: chunk long documents, embed and retrieve relevant chunks (RAG), then send only the top-k chunks with the prompt to fit token limits.
- Stream responses to the client and support cancellation (AbortController) to improve UX and reduce cost.
- Implement model fallback and graceful degradation: detect rate-limit or quota errors, backoff + retry, then fallback to a cheaper/secondary model.
- Enforce token accounting: count tokens, truncate or summarize inputs that exceed limits, and validate prompt size before sending.
- Sanitize and canonicalize inputs server-side; remove PII/sensitive data before augmenting context or sending to models.
- Keep API keys and secrets server-only (env vars, secret manager). Never bundle secrets in client code.
- Instrument every model call (latency, tokens used, error type) and enforce application-level quotas and rate limiting.
- Cache deterministic results and embeddings to reduce repeated calls and costs; store embeddings with provenance metadata for re-use.
- Provide clear user-facing error/fallback messages and telemetry for model switching decisions.
- Apply safety/content filtering layers for untrusted inputs and honor user consent / data retention policy.

React/Next.js
- Use functional components and TypeScript interfaces.
- Use declarative JSX.
- Use function, not const, for components.
- Use Shadcn UI, Radix, and Tailwind CSS for components and styling.
- Implement responsive design with Tailwind CSS.
- Use mobile-first approach for responsive design.
- Place static content and interfaces at file end.
- Use content variables for static content outside render functions.
- Minimize 'use client', 'useEffect', and 'setState'. Favor React Server Components (RSC).
- Use Zod for form validation.
- Wrap client components in Suspense with fallback.
- Use dynamic loading for non-critical components.
- Optimize images: WebP format, size data, lazy loading.
- Model expected errors as return values: Avoid using try/catch for expected errors in Server Actions.
- Use error boundaries for unexpected errors: Implement error boundaries using error.tsx and global-error.tsx files.
- Use useActionState with react-hook-form for form validation.
- Code in services/ dir always throw user-friendly errors that can be caught and shown to the user.
- Use next-safe-action for all server actions.
- Implement type-safe server actions with proper validation.
- Handle errors gracefully and return appropriate responses.

Supabase and GraphQL
- Use the official Supabase helpers and clients appropriate to each Next.js environment:
  - Client components / browser: use a browser client (e.g. `createBrowserClient`) with public keys
  - Server components / server actions: use `createServerComponentClient({ cookies: () => cookies() })`
  - Route handlers / API routes: use `createRouteHandlerClient({ cookies: () => cookies() })` or `createServerClient(req,res)` for Pages
  - Keep session storage in HttpOnly cookies (helpers manage this); avoid localStorage for auth
- Security: keep `service_role` keys strictly server-side (Edge Functions, admin scripts). Never expose service_role or other admin keys to the client bundle
- Row Level Security (RLS): design RLS policies for multi-tenant isolation using JWT claims (e.g. `current_setting('request.jwt.claims.company_id')`). Test RLS policies with automated integration tests against a local Supabase instance
- GraphQL and Genql:
  - Use codegen (genql / GraphQL Code Generator) to keep client types in sync with schema
  - Prefer persisted queries for public clients and limit query depth/complexity on the server
  - Use DataLoader or batching to avoid N+1 resolver patterns when resolving GraphQL fields
- Runtime and edge considerations:
  - When using Edge runtime in route handlers or server components, set `export const runtime = 'edge'` and use `dynamic = 'force-dynamic'` where appropriate
  - Use Edge Functions for webhooks, background processing and tasks that benefit from geographic proximity
- CI and developer workflow:
  - Add a codegen step to CI (fetch schema -> generate genql/types) and fail the pipeline if generated artifacts differ
  - Generate and check database types (supabase types) in CI to prevent type drift
- Performance & caching:
  - Use server-side caching headers and Next.js incremental revalidation for queries that tolerate staleness
  - Cache GraphQL responses or use CDN for persisted operations where possible
- Observability & auditing:
  - Log auth events and RLS denials; rotate keys regularly and monitor usage of service_role keys
- Use Supabase Edge Functions for server-side logic that requires elevated privileges or background processing.

Key Conventions
1. Rely on Next.js App Router for state changes and routing.
2. Prioritize Web Vitals (LCP, CLS, FID).
3. Minimize 'use client' usage:
  - Prefer server components and Next.js SSR features.
  - Use 'use client' only for Web API access in small components.
  - Avoid using 'use client' for data fetching or state management.
4. Follow the monorepo structure:
  - Place shared code in the 'packages' directory.
  - Keep app-specific code in the 'apps' directory.
5. Use Taskfile commands for development and deployment tasks.
6. Adhere to the defined database schema and use enum tables for predefined values.

Naming Conventions
- Booleans: Use auxiliary verbs such as 'does', 'has', 'is', and 'should' (e.g., isDisabled, hasError).
- Filenames: Use lowercase with dash separators (e.g., auth-wizard.tsx).
- File extensions: Use .config.ts, .test.ts, .context.tsx, .type.ts, .hook.ts as appropriate.

Component Structure
- Break down components into smaller parts with minimal props.
- Suggest micro folder structure for components.
- Use composition to build complex components.
- Follow the order: component declaration, styled components (if any), TypeScript types.

Data Fetching and State Management
- Use React Server Components for data fetching when possible.
- Implement the preload pattern to prevent waterfalls.
- Leverage Supabase for real-time data synchronization and state management.
- Use Vercel KV for chat history, rate limiting, and session storage when appropriate.

Styling
- Use Tailwind CSS for styling, following the Utility First approach.
- Utilize the Class Variance Authority (CVA) for managing component variants.

Testing
- Implement unit tests for utility functions and hooks.
- Use integration tests for complex components and pages.
- Implement end-to-end tests for critical user flows.
- Use Supabase local development for testing database interactions.

Accessibility
- Ensure interfaces are keyboard navigable.
- Implement proper ARIA labels and roles for components.
- Ensure color contrast ratios meet WCAG standards for readability.

Documentation
- Provide clear and concise comments for complex logic.
- Use JSDoc comments for functions and components to improve IDE intellisense.
- Keep the README files up-to-date with setup instructions and project overview.
- Document Supabase schema, RLS policies, and Edge Functions when used.

Refer to Next.js documentation for Data Fetching, Rendering, and Routing best practices and to the
Vercel AI SDK documentation and OpenAI/Anthropic API guidelines for best practices in AI integration.
