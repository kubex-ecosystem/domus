-- ============================================================================
-- ETAPA 7: Triggers
-- ============================================================================
-- Cria função genérica e triggers para auto-update de updated_at
-- ============================================================================
-- User
-- kubex_adm
-- Default DB
-- postgres
-- Default Schema
-- public
-- Garante que estamos no schema correto
SET search_path TO public;
\ echo 'ETAPA 7: Criando triggers...' -- Função genérica para atualizar updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$ BEGIN NEW.updated_at = now();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;
\ echo '  Função update_updated_at_column() criada' -- Aplicar triggers
CREATE TRIGGER update_tenant_updated_at BEFORE
UPDATE ON tenant FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_team_updated_at BEFORE
UPDATE ON team FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_updated_at BEFORE
UPDATE ON "user" FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tenant_membership_updated_at BEFORE
UPDATE ON tenant_membership FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_team_membership_updated_at BEFORE
UPDATE ON team_membership FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_partner_invitation_updated_at BEFORE
UPDATE ON partner_invitation FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_internal_invitation_updated_at BEFORE
UPDATE ON internal_invitation FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_pipeline_updated_at BEFORE
UPDATE ON pipeline FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_pipeline_stage_updated_at BEFORE
UPDATE ON pipeline_stage FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_partner_updated_at BEFORE
UPDATE ON partner FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_lead_updated_at BEFORE
UPDATE ON lead FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_commission_updated_at BEFORE
UPDATE ON commission FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_clawback_updated_at BEFORE
UPDATE ON clawback FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_pending_access_requests_updated_at BEFORE
UPDATE ON pending_access_requests FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
\ echo '  14 triggers criados' -- End log
\ echo 'ETAPA 7 concluída com sucesso!'