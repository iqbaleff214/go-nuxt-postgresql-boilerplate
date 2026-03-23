export interface ApiResponse<T> {
  success: boolean
  message: string
  data: T
  meta?: PaginationMeta
}

export interface PaginationMeta {
  page: number
  total: number
  page_size: number
}

export interface ApiError {
  success: false
  message: string
  errors?: FieldError[]
}

export interface FieldError {
  field: string
  detail: string
}
