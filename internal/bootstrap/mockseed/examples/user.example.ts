/**
 * Type para tabela: user
 * Gerado automaticamente em 2025-12-02T10:30:00-03:00
 */

export interface User {
  id: string;
  email: string;
  name?: string;
  last_name?: string;
  password_hash?: string;
  phone?: string;
  avatar_url?: string;
  status?: string;
  force_password_reset?: boolean;
  last_login?: Date | string;
  created_at: Date | string;
  updated_at?: Date | string;
}
