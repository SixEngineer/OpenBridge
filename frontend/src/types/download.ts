// 与后端 entity.DownloadTask 字段对齐（Go 默认 PascalCase JSON 序列化）
export type DownloadStatus =
  | 'pending'
  | 'resolving'
  | 'resolved'
  | 'submitted'
  | 'downloading'
  | 'completed'
  | 'failed'
  | 'cancelled'

// 直链解析结果
export interface DirectLinkResult {
  Path: string
  Name: string
  Size: number
  Provider: string
  DirectLink: string
  IsOpenListProxy: boolean
}

export interface DownloadTask {
  ID: number
  TaskID: string
  SourcePath: string
  FileName: string
  FileSize: number
  DirectLink: string
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
