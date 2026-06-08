import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { DownloadTask, DirectLinkResult } from '@/types/download'
import endpoints from './endpoints'
import { ensureDeviceId } from '@/utils/session'

// 创建下载任务（后端：POST /download/tasks, body: {path, dir}）
export function createTask(data: {
  path: string
  dir?: string
}, options?: {
  timeout?: number
}): Promise<ApiResponse<{ task_id: string }>> {
  return request.post(endpoints.downloadTasks, data, {
    timeout: options?.timeout ?? 60000,
  })
}

// 解析直链（后端：POST /download/resolve, body: {path})
export function resolveDirectLink(path: string, options?: {
  timeout?: number
}): Promise<ApiResponse<DirectLinkResult>> {
  return request.post(endpoints.downloadResolve, { path }, {
    timeout: options?.timeout ?? 60000,
  })
}

export function buildFolderZipUrl(path: string): string {
  const apiBase = import.meta.env.VITE_API_BASE_URL || '/api/v1'
  const base = typeof window === 'undefined' ? 'http://localhost' : window.location.origin
  const url = new URL(`${apiBase}${endpoints.downloadFolderZip}`, base)
  url.searchParams.set('path', path)
  url.searchParams.set('device_id', ensureDeviceId())
  return url.toString()
}

// 获取单个任务详情（后端：GET /download/tasks/:id）
export function getTaskDetail(taskId: string): Promise<ApiResponse<DownloadTask>> {
  return request.get(endpoints.downloadTaskDetail(taskId))
}

// 重试下载任务（后端：POST /download/tasks/:id/retry）
export function retryTask(taskId: string): Promise<ApiResponse<DownloadTask>> {
  return request.post(endpoints.downloadTaskRetry(taskId))
}

// 打开已下载的文件（后端：POST /download/tasks/:id/open）
export function openFile(taskId: string): Promise<ApiResponse<{ file_path: string }>> {
  return request.post(endpoints.downloadTaskOpen(taskId))
}

// 打开文件所在文件夹（后端：POST /download/tasks/:id/open-location）
export function openFileLocation(taskId: string): Promise<ApiResponse<{ folder_path: string }>> {
  return request.post(endpoints.downloadTaskOpenLocation(taskId))
}

// 获取 aria2 状态（后端：GET /download/aria2-status）
export function getAria2Status(): Promise<ApiResponse<{ version: string, downloadSpeed?: string | number, uploadSpeed?: string | number }>> {
  return request.get(endpoints.downloadAria2Status)
}
