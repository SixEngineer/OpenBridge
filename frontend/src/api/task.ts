import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { DownloadTask, DirectLinkResult } from '@/types/download'
import endpoints from './endpoints'

// 创建下载任务（后端：POST /download/tasks, body: {path, dir}）
export function createTask(data: {
  path: string
  dir?: string
}): Promise<ApiResponse<{ task_id: string }>> {
  return request.post(endpoints.downloadTasks, data)
}

// 解析直链（后端：POST /download/resolve, body: {path})
export function resolveDirectLink(path: string): Promise<ApiResponse<DirectLinkResult>> {
  return request.post(endpoints.downloadResolve, { path })
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
export function getAria2Status(): Promise<ApiResponse<{ version: string }>> {
  return request.get(endpoints.downloadAria2Status)
}