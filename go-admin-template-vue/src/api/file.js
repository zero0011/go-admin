import request from '@/utils/request'

// ========== 文件夹管理 API ==========

// 获取文件夹列表
export function getFolderList(params) {
    return request({
        url: '/file/folder/list',
        method: 'get',
        params
    })
}

// 创建文件夹
export function createFolder(data) {
    return request({
        url: '/file/folder',
        method: 'post',
        data
    })
}

// 重命名文件夹
export function renameFolder(id, data) {
    return request({
        url: `/file/folder/${id}`,
        method: 'put',
        data
    })
}

// 删除文件夹
export function deleteFolder(id) {
    return request({
        url: `/file/folder/${id}`,
        method: 'delete'
    })
}

// ========== 文件管理 API ==========

// 获取文件列表
export function getFileList(params) {
    return request({
        url: '/file/list',
        method: 'get',
        params
    })
}

// 单文件上传
export function uploadFile(formData) {
    return request({
        url: '/file/upload',
        method: 'post',
        data: formData,
        headers: {
            'Content-Type': 'multipart/form-data'
        }
    })
}

// 删除文件
export function deleteFile(id) {
    return request({
        url: `/file/${id}`,
        method: 'delete'
    })
}

// 批量删除文件
export function batchDeleteFiles(data) {
    return request({
        url: '/file/batch-delete',
        method: 'post',
        data
    })
}

// ========== 分片上传 API ==========

// 初始化分片上传
export function initChunkUpload(data) {
    return request({
        url: '/file/chunk/init',
        method: 'post',
        data
    })
}

// 上传分片
export function uploadChunk(formData) {
    return request({
        url: '/file/chunk/upload',
        method: 'post',
        data: formData,
        headers: {
            'Content-Type': 'multipart/form-data'
        }
    })
}

// 合并分片
export function mergeChunk(data) {
    return request({
        url: '/file/chunk/merge',
        method: 'post',
        data
    })
}

// 获取分片上传状态
export function getChunkStatus(params) {
    return request({
        url: '/file/chunk/status',
        method: 'get',
        params
    })
}
