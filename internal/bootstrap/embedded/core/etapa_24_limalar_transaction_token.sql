-- ============================================================================
-- ETAPA 24: Limalar — Token de Acesso Público para Transações
-- ============================================================================
-- Adiciona access_token à tabela limalar.transactions para permitir que o
-- comprador/vendedor acompanhe o progresso do fechamento sem login.
--
-- O token é gerado automaticamente (gen_random_uuid) e pode ser regenerado
-- pelo corretor a qualquer momento. A rota pública /acompanhar/:token retorna
-- apenas os dados permitidos (steps + docs sem dados sensíveis do tenant).
--
-- Depende de: etapa_22 (limalar.transactions)
-- ============================================================================
\echo 'ETAPA 24: Adicionando access_token em limalar.transactions...'

ALTER TABLE limalar.transactions
    ADD COLUMN IF NOT EXISTS access_token UUID UNIQUE DEFAULT gen_random_uuid();

-- Garante que linhas existentes tenham token
UPDATE limalar.transactions SET access_token = gen_random_uuid() WHERE access_token IS NULL;

CREATE INDEX IF NOT EXISTS idx_limalar_transactions_access_token
    ON limalar.transactions (access_token) WHERE access_token IS NOT NULL;

\echo '  access_token adicionado e indexado em limalar.transactions'
\echo 'ETAPA 24 concluída com sucesso!'
