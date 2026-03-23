export type UserRole = 'user' | 'superadmin'
export type UserStatus = 'active' | 'inactive' | 'banned'

export interface User {
  id: string
  email: string
  first_name: string
  last_name: string
  display_name: string
  bio: string | null
  avatar_url: string | null
  role: UserRole
  status: UserStatus
  is_email_verified: boolean
  is_2fa_enabled: boolean
  last_login_at: string | null
  created_at: string
  updated_at: string
}

export interface Notification {
  id: string
  type: string
  title: string
  body: string | null
  read_at: string | null
  created_at: string
}
