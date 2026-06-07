// 与后端 entity.DownloadTask 字段对齐（Go 默认 PascalCase JSON 序列化）
export type DownloadStatus =
  | 'waiting'
  | 'active'
  | 'error'
  | 'complete'

// 直链解析结果（后端 json 为 snake_case）
export interface DirectLinkResult {
  path: string
  name: string
  size: number
  provider: string
  direct_link: string
  is_openlist_proxy: boolean
}

export interface DownloadTask {
  ID: number
  TaskID: string
  SourcePath: string
  FileName: string
  FileSize: number
  DirectLink: string
  FilePath: string
  Aria2GID: string
  Status: string
  Progress: number
  ErrorMessage: string
  RetryCount: number
  StartedAt: string | null
  FinishedAt: string | null
  CreatedAt: string
  UpdatedAt: string
}
