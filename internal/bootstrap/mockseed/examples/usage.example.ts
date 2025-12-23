/**
 * Exemplo de uso dos types e mocks gerados pelo Meta-Seeder
 * Copie este exemplo para seu projeto frontend
 */

// ==========================================
// 1. IMPORTAR TYPES
// ==========================================

import { User, Org, Tenant, Role, Permission } from './types'
import { orgMocks, userMocks } from './types/mocks'
import { getMocks, findById, filterBy } from './types/mocks/helpers'

// ==========================================
// 2. USAR TYPES EM FUNÇÕES
// ==========================================

// Função type-safe para criar usuário
function createUser(data: Partial<User>): User {
  return {
    id: data.id || crypto.randomUUID(),
    email: data.email!,
    name: data.name,
    created_at: new Date().toISOString(),
    ...data
  }
}

// Função type-safe para validar org
function isValidOrg(org: Org): boolean {
  return org.id.length > 0 && org.name.length > 0
}

// ==========================================
// 3. USAR MOCKS EM DESENVOLVIMENTO
// ==========================================

// Usar mocks diretamente
const allOrgs = orgMocks
console.log('Orgs disponíveis:', allOrgs)

// Buscar por ID com type-safety
const org = findById<Org>('org', '10000000-0000-0000-0000-000000000001')
if (org) {
  console.log('Org encontrada:', org.name)
}

// Filtrar usuários ativos
const activeUsers = filterBy<User>('user', 'status', 'active')

// ==========================================
// 4. INTEGRAÇÃO COM API
// ==========================================

class UserService {
  async getUser(id: string): Promise<User> {
    // Em desenvolvimento, retornar mock
    if (process.env.NODE_ENV === 'development') {
      const mockUser = findById<User>('user', id)
      if (mockUser) return mockUser
    }

    // Em produção, chamar API real
    const response = await fetch(`/api/users/${id}`)
    return response.json()
  }

  async listOrgs(): Promise<Org[]> {
    // Em desenvolvimento, usar mocks
    if (process.env.NODE_ENV === 'development') {
      return orgMocks
    }

    // Em produção, chamar API
    const response = await fetch('/api/orgs')
    return response.json()
  }
}

// ==========================================
// 5. VALIDAÇÃO COM TYPES
// ==========================================

// Type guard
function isUser(obj: any): obj is User {
  return (
    typeof obj.id === 'string' &&
    typeof obj.email === 'string' &&
    typeof obj.created_at === 'string'
  )
}

// Validar dados da API
async function fetchAndValidateUser(id: string): Promise<User | null> {
  const data = await fetch(`/api/users/${id}`).then(r => r.json())

  if (isUser(data)) {
    return data
  }

  console.error('Dados inválidos recebidos da API:', data)
  return null
}

// ==========================================
// 6. TESTES UNITÁRIOS
// ==========================================

describe('UserService', () => {
  it('should return user from mock', async () => {
    const service = new UserService()
    const user = await service.getUser('test-id')

    // Type-safe assertions
    expect(user.email).toBeDefined()
    expect(user.created_at).toBeDefined()
  })

  it('should list all orgs', async () => {
    const service = new UserService()
    const orgs = await service.listOrgs()

    // Usar mocks para testes
    expect(orgs.length).toBeGreaterThan(0)
    expect(orgs[0]).toHaveProperty('name')
  })
})

// ==========================================
// 7. REACT/NEXT.JS EXAMPLE
// ==========================================

// React component com types
interface UserCardProps {
  user: User
}

function UserCard({ user }: UserCardProps) {
  return (
    <div className="user-card">
      <h3>{user.name || user.email}</h3>
      <p>Email: {user.email}</p>
      <p>Criado em: {new Date(user.created_at).toLocaleDateString()}</p>
    </div>
  )
}

// Next.js page com Server Component
async function UsersPage() {
  // Em desenvolvimento, usar mocks
  const users = process.env.NODE_ENV === 'development'
    ? getMocks<User>('user')
    : await fetch('/api/users').then(r => r.json())

  return (
    <div>
      <h1>Usuários</h1>
      {users.map(user => (
        <UserCard key={user.id} user={user} />
      ))}
    </div>
  )
}

// ==========================================
// 8. VANTAGENS DO META-SEEDER
// ==========================================

/*
✅ Type-safety completa em todo o frontend
✅ Autocomplete no VSCode/IDE
✅ Mocks reais do banco de dados
✅ Zero configuração manual
✅ Sincronização automática PG → TS
✅ Testes mais confiáveis
✅ Desenvolvimento offline
✅ Documentação viva do schema

Para atualizar types quando o schema mudar:

1. cd internal/bootstrap/mockseed
2. ./run_bootstrap.sh full-pipeline
3. cp -r types /path/to/frontend/src/types/database

Pronto! Frontend atualizado automaticamente.
*/
