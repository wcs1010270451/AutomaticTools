import axios from 'axios'

export interface AdminAccount {
  id: number
  username: string
  status: string
  createdAt?: number
  lastLoginAt?: number
}

export interface AdminLoginResponse {
  token: string
  admin: AdminAccount
}

export interface PlatformUser {
  id: number
  username: string
  email?: string
  phone?: string
  status: 'active' | 'disabled'
  createdAt?: number
  lastLoginAt?: number
}

export interface UserListParams {
  page: number
  pageSize: number
  search?: string
  status?: '' | 'active' | 'disabled'
}

export interface UserListResponse {
  users: PlatformUser[]
  total: number
  page: number
  pageSize: number
}

export interface CreateAdminPayload {
  username: string
  password: string
  status: 'active' | 'disabled'
}

export interface UpdateAdminPayload {
  username: string
  password?: string
  status: 'active' | 'disabled'
}

export interface AdminTool {
  id: number
  code: string
  name: string
  description: string
  priceCents: number
  currency: string
  lifetime: boolean
  active: boolean
  createdAt: number
}

export interface CreateToolPayload {
  code: string
  name: string
  description: string
  priceCents: number
  currency: string
  lifetime: boolean
  active: boolean
}

export type UpdateToolPayload = Omit<CreateToolPayload, 'code'>

export type LicenseCodeStatus = 'active' | 'redeemed' | 'revoked'

export interface LicenseCodeRecord {
  id: number
  codeHint: string
  toolCode: string
  toolName: string
  batchNo: string
  note: string
  status: LicenseCodeStatus
  createdByAdminUsername?: string
  redeemedByUserId?: number
  redeemedByUsername?: string
  redeemedByEmail?: string
  redeemedAt?: number
  revokedAt?: number
  createdAt: number
}

export interface LicenseCodeListParams {
  page: number
  pageSize: number
  status?: '' | LicenseCodeStatus
  toolCode?: string
  search?: string
}

export interface LicenseCodeListResponse {
  codes: LicenseCodeRecord[]
  total: number
  page: number
  pageSize: number
}

export interface GenerateLicenseCodesPayload {
  toolCode: string
  count: number
  note: string
}

export interface GeneratedLicenseCode {
  code: string
  toolCode: string
  batchNo: string
}

export interface GenerateLicenseCodesResponse {
  batchNo: string
  codes: GeneratedLicenseCode[]
}

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 10_000,
  headers: {
    'Content-Type': 'application/json',
  },
})

api.interceptors.request.use((request) => {
  const token = localStorage.getItem('automatictools.admin.token')
  if (token) {
    request.headers.Authorization = `Bearer ${token}`
  }
  return request
})

export async function loginAdmin(username: string, password: string) {
  const response = await api.post<AdminLoginResponse>('/api/admin/auth/login', {
    username,
    password,
  })
  return response.data
}

export async function listUsers(params: UserListParams) {
  const response = await api.get<UserListResponse>('/api/admin/users', { params })
  return response.data
}

export async function listAdmins() {
  const response = await api.get<{ admins: AdminAccount[] }>('/api/admin/admins')
  return response.data.admins
}

export async function createAdmin(payload: CreateAdminPayload) {
  const response = await api.post<{ admin: AdminAccount }>('/api/admin/admins', payload)
  return response.data.admin
}

export async function updateAdmin(adminId: number, payload: UpdateAdminPayload) {
  const response = await api.put<{ admin: AdminAccount }>(`/api/admin/admins/${adminId}`, payload)
  return response.data.admin
}

export async function deleteAdmin(adminId: number) {
  await api.delete(`/api/admin/admins/${adminId}`)
}

export async function listAdminTools() {
  const response = await api.get<{ tools: AdminTool[] }>('/api/admin/tools')
  return response.data.tools
}

export async function createTool(payload: CreateToolPayload) {
  const response = await api.post<{ tool: AdminTool }>('/api/admin/tools', payload)
  return response.data.tool
}

export async function updateTool(code: string, payload: UpdateToolPayload) {
  const response = await api.put<{ tool: AdminTool }>(
    `/api/admin/tools/${encodeURIComponent(code)}`,
    payload,
  )
  return response.data.tool
}

export async function listLicenseCodes(params: LicenseCodeListParams) {
  const response = await api.get<LicenseCodeListResponse>('/api/admin/license-codes', { params })
  return response.data
}

export async function generateLicenseCodes(payload: GenerateLicenseCodesPayload) {
  const response = await api.post<GenerateLicenseCodesResponse>(
    '/api/admin/license-codes/generate',
    payload,
  )
  return response.data
}

export async function revokeLicenseCode(id: number) {
  const response = await api.post<{ code: LicenseCodeRecord }>(
    `/api/admin/license-codes/${id}/revoke`,
  )
  return response.data.code
}
