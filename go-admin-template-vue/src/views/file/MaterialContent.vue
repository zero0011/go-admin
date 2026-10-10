<template>
    <div class="material-content">
        <!-- 工具栏 -->
        <div class="content-header">
            <!-- 搜索结果横幅 -->
            <div v-if="searchMode" class="search-banner">
                <span class="search-banner-text">
                    搜索 “{{ searchKeyword }}”，共 {{ total }} 个结果
                </span>
                <el-button link type="primary" @click="emit('clear-search')">
                    清除搜索
                </el-button>
            </div>

            <!-- 面包屑 -->
            <div class="breadcrumb" v-else>
                <el-breadcrumb separator="/">
                    <el-breadcrumb-item
                        v-for="(item, index) in breadcrumbList"
                        :key="index">
                        <span
                            :class="{ 'breadcrumb-link': index < breadcrumbList.length - 1 }"
                            @click="handleBreadcrumbClick(index)">
                            {{ item.name }}
                        </span>
                    </el-breadcrumb-item>
                </el-breadcrumb>
            </div>

            <!-- 上传按钮 -->
            <div class="header-actions">
                <el-upload
                    ref="uploadRef"
                    :action="uploadUrl"
                    :data="uploadData"
                    :multiple="false"
                    :auto-upload="false"
                    :show-file-list="false"
                    :on-change="handleUploadChange"
                    :on-success="handleUploadSuccess"
                    :on-error="handleUploadError">
                    <el-button type="primary">
                        <el-icon><Upload /></el-icon>
                        上传文件
                    </el-button>
                </el-upload>
            </div>
        </div>

        <!-- 文件列表 -->
        <div class="content-body" v-loading="loading">
            <div class="file-grid" v-if="fileList.length > 0">
                <div
                    v-for="item in fileList"
                    :key="item.type === 'folder' ? item.folder_id : item.asset_id"
                    class="file-item"
                    @click="handleItemClick(item)">
                    <!-- 文件夹 -->
                    <div v-if="item.type === 'folder'" class="file-folder">
                        <div class="file-icon folder-icon">
                            <el-icon :size="48"><Folder /></el-icon>
                        </div>
                        <div class="file-name">{{ item.folder_name }}</div>
                        <div
                            v-if="searchMode && item.path && item.path.length"
                            class="file-path"
                            :title="formatPath(item.path)">
                            {{ formatPath(item.path) }}
                        </div>
                        <div class="file-actions">
                            <el-button
                                type="primary"
                                link
                                @click.stop="handleRename(item)">
                                重命名
                            </el-button>
                            <el-button
                                type="danger"
                                link
                                @click.stop="handleDeleteFolder(item)">
                                删除
                            </el-button>
                        </div>
                    </div>

                    <!-- 文件 -->
                    <div v-else class="file-asset">
                        <div class="file-icon">
                            <el-image
                                v-if="isImage(item.file_ext)"
                                :src="item.file_url"
                                :preview-src-list="[item.file_url]"
                                fit="cover"
                                class="file-preview"/>
                            <el-icon v-else :size="48"><Document /></el-icon>
                        </div>
                        <div class="file-name" :title="item.file_name">{{ item.file_name }}</div>
                        <div
                            v-if="searchMode && item.path && item.path.length"
                            class="file-path"
                            :title="formatPath(item.path)">
                            {{ formatPath(item.path) }}
                        </div>
                        <div class="file-size">{{ formatFileSize(item.file_size) }}</div>
                        <div class="file-actions">
                            <el-button
                                type="primary"
                                link
                                @click.stop="handleDownload(item)">
                                下载
                            </el-button>
                            <el-button
                                type="danger"
                                link
                                @click.stop="handleDeleteFile(item)">
                                删除
                            </el-button>
                        </div>
                    </div>
                </div>
            </div>

            <!-- 空状态 -->
            <el-empty v-else description="暂无文件" />
        </div>

        <!-- 分页 -->
        <div class="content-footer" v-if="total > 0">
            <el-pagination
                v-model:current-page="localPage"
                :page-size="pageSize"
                :total="total"
                layout="total, prev, pager, next"
                @current-change="handlePageChange"/>
        </div>

        <!-- 重命名对话框 -->
        <el-dialog v-model="renameDialogVisible" title="重命名" width="400px">
            <el-input v-model="renameValue" placeholder="请输入新名称" />
            <template #footer>
                <el-button @click="renameDialogVisible = false">取消</el-button>
                <el-button type="primary" @click="confirmRename">确定</el-button>
            </template>
        </el-dialog>

        <!-- 大文件上传对话框（分片上传） -->
        <el-dialog
            v-model="chunkUploadDialogVisible"
            title="大文件上传"
            width="600px"
            :close-on-click-modal="false">
            <div class="chunk-upload">
                <el-progress
                    v-if="chunkProgress > 0"
                    :percentage="chunkProgress"
                    :status="chunkProgress === 100 ? 'success' : undefined"/>
                <div v-if="chunkStatus === 'uploading'" class="chunk-status">
                    正在上传分片...
                </div>
                <div v-if="chunkStatus === 'merging'" class="chunk-status">
                    正在合并文件...
                </div>
                <div v-if="chunkStatus === 'success'" class="chunk-status" style="color: #67c23a">
                    上传成功！
                </div>
            </div>
        </el-dialog>
    </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { Upload, Folder, Document } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { uploadFile, initChunkUpload, uploadChunk, mergeChunk } from '@/api/file'

const props = defineProps({
    loading: {
        type: Boolean,
        default: false
    },
    fileList: {
        type: Array,
        default: () => []
    },
    total: {
        type: Number,
        default: 0
    },
    currentFolderId: {
        type: Number,
        default: null
    },
    breadcrumbList: {
        type: Array,
        default: () => []
    },
    page: {
        type: Number,
        default: 1
    },
    pageSize: {
        type: Number,
        default: 20
    },
    searchMode: {
        type: Boolean,
        default: false
    },
    searchKeyword: {
        type: String,
        default: ''
    }
})

const emit = defineEmits([
    'folder-click',
    'file-click',
    'folder-back',
    'folder-rename',
    'folder-delete',
    'file-delete',
    'file-download',
    'page-change',
    'upload-success',
    'search',
    'clear-search'
])

const uploadUrl = '/api/file/upload'
const uploadRef = ref(null)
const localPage = ref(props.page)

// 分片上传相关
const chunkUploadDialogVisible = ref(false)
const chunkProgress = ref(0)
const chunkStatus = ref('')
const currentUploadFile = ref(null)
const uploadData = ref({})

// 重命名相关
const renameDialogVisible = ref(false)
const renameValue = ref('')
const renameItem = ref(null)

watch(() => props.page, (val) => {
    localPage.value = val
})

// 判断是否为图片
const isImage = (ext) => {
    const imageExts = ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.bmp']
    return imageExts.includes((ext || '').toLowerCase())
}

// 格式化文件大小
const formatFileSize = (size) => {
    if (!size) return '0 B'
    const units = ['B', 'KB', 'MB', 'GB']
    let index = 0
    let s = size
    while (s >= 1024 && index < units.length - 1) {
        s /= 1024
        index++
    }
    return `${s.toFixed(2)} ${units[index]}`
}

// 格式化搜索结果所在路径
const formatPath = (path) => (path || []).map(p => p.name).join(' / ')

// 点击面包屑
const handleBreadcrumbClick = (index) => {
    emit('folder-back', index)
}

// 点击文件/文件夹
const handleItemClick = (item) => {
    if (item.type === 'folder') {
        emit('folder-click', item)
    } else {
        emit('file-click', item)
    }
}

// 上传文件变化
const handleUploadChange = async (file) => {
    const CHUNK_SIZE = 5 * 1024 * 1024 // 5MB

    if (file.size > CHUNK_SIZE) {
    // 大文件分片上传
        await handleChunkUpload(file)
    } else {
    // 小文件直接上传
        await handleNormalUpload(file)
    }
}

// 普通上传
const handleNormalUpload = async (file) => {
    const formData = new FormData()
    formData.append('file', file.raw)
    if (props.currentFolderId) {
        formData.append('parent_folder_id', props.currentFolderId)
    }

    try {
        const res = await uploadFile(formData)
        if (res.code === 0) {
            ElMessage.success('上传成功')
            emit('upload-success')
        }
    } catch {
        ElMessage.error('上传失败')
    }
}

// 分片上传
const handleChunkUpload = async (file) => {
    const CHUNK_SIZE = 5 * 1024 * 1024 // 5MB

    chunkUploadDialogVisible.value = true
    chunkProgress.value = 0
    chunkStatus.value = 'uploading'
    currentUploadFile.value = file

    try {
        // 1. 计算文件hash（简化处理，实际应使用Web Worker）
        const fileHash = await calculateFileHash(file.raw)

        // 2. 初始化分片上传
        const initRes = await initChunkUpload({
            file_name: file.name,
            file_size: file.size,
            chunk_size: CHUNK_SIZE,
            file_hash: fileHash,
            parent_folder_id: props.currentFolderId
        })

        if (initRes.code !== 0) {
            throw new Error(initRes.msg || '初始化上传失败')
        }

        const { upload_id, chunk_count } = initRes.data

        // 3. 读取文件并分片上传
        for (let i = 0; i < chunk_count; i++) {
            const start = i * CHUNK_SIZE
            const end = Math.min(start + CHUNK_SIZE, file.size)
            const blob = file.raw.slice(start, end)

            const chunkFormData = new FormData()
            chunkFormData.append('upload_id', upload_id)
            chunkFormData.append('chunk_index', i)
            chunkFormData.append('chunk_hash', fileHash + '_' + i)
            chunkFormData.append('file', blob)

            await uploadChunk(chunkFormData)

            chunkProgress.value = Math.round(((i + 1) / chunk_count) * 80)
        }

        // 4. 合并分片
        chunkStatus.value = 'merging'
        const mergeRes = await mergeChunk({
            upload_id,
            parent_folder_id: props.currentFolderId
        })

        if (mergeRes.code !== 0) {
            throw new Error(mergeRes.msg || '合并文件失败')
        }

        chunkProgress.value = 100
        chunkStatus.value = 'success'

        setTimeout(() => {
            chunkUploadDialogVisible.value = false
            emit('upload-success')
        }, 1000)
    } catch (error) {
        chunkStatus.value = 'error'
        ElMessage.error(error.message || '上传失败')
    }
}

// 计算文件hash（简化版）
const calculateFileHash = async (file) => {
    return new Promise((resolve) => {
        const reader = new FileReader()
        reader.onload = (e) => {
            const hashBuffer = e.target.result
            const hashArray = Array.from(new Uint8Array(hashBuffer))
            const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('')
            resolve(hashHex.substring(0, 32))
        }
        reader.readAsArrayBuffer(file.slice(0, 1024 * 1024)) // 取前1MB计算hash
    })
}

// 上传成功
const handleUploadSuccess = (response) => {
    if (response.code === 0) {
        ElMessage.success('上传成功')
        emit('upload-success')
    } else {
        ElMessage.error(response.msg || '上传失败')
    }
}

// 上传失败
const handleUploadError = () => {
    ElMessage.error('上传失败')
}

// 重命名
const handleRename = (item) => {
    renameItem.value = item
    renameValue.value = item.type === 'folder' ? item.folder_name : item.file_name
    renameDialogVisible.value = true
}

const confirmRename = () => {
    if (!renameValue.value.trim()) {
        ElMessage.warning('请输入新名称')
        return
    }

    if (renameItem.value.type === 'folder') {
        emit('folder-rename', renameItem.value.folder_id, renameValue.value.trim())
    }

    renameDialogVisible.value = false
}

// 删除文件夹
const handleDeleteFolder = (item) => {
    ElMessageBox.confirm(
        `确定要删除文件夹 "${item.folder_name}" 吗？`,
        '警告',
        {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
        }
    ).then(() => {
        emit('folder-delete', item.folder_id)
    }).catch(() => {})
}

// 删除文件
const handleDeleteFile = (item) => {
    ElMessageBox.confirm(
        `确定要删除文件 "${item.file_name}" 吗？`,
        '警告',
        {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
        }
    ).then(() => {
        emit('file-delete', item.asset_id)
    }).catch(() => {})
}

// 下载文件
const handleDownload = (item) => {
    emit('file-download', item)
}

// 分页变化
const handlePageChange = (page) => {
    emit('page-change', page)
}
</script>

<style scoped>
.material-content {
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: #fff;
}

.content-header {
  padding: 16px 20px;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-banner {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 14px;
  color: #606266;
}

.search-banner-text {
  font-weight: 500;
}

.breadcrumb-link {
  cursor: pointer;
  color: #409eff;
}

.breadcrumb-link:hover {
  text-decoration: underline;
}

.content-body {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.file-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 20px;
}

.file-item {
  cursor: pointer;
  padding: 16px;
  border-radius: 8px;
  transition: background-color 0.3s;
  text-align: center;
}

.file-item:hover {
  background-color: #f5f7fa;
}

.file-icon {
  width: 80px;
  height: 80px;
  margin: 0 auto 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.file-icon img,
.file-preview {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 8px;
}

.folder-icon {
  color: #ffd04b;
}

.file-name {
  font-size: 14px;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 4px;
}

.file-path {
  font-size: 12px;
  color: #909399;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100%;
  margin-bottom: 4px;
}

.file-size {
  font-size: 12px;
  color: #909399;
  margin-bottom: 8px;
}

.file-actions {
  display: none;
  justify-content: center;
  gap: 8px;
}

.file-item:hover .file-actions {
  display: flex;
}

.content-footer {
  padding: 16px 20px;
  border-top: 1px solid #e4e7ed;
  display: flex;
  justify-content: flex-end;
}

.chunk-upload {
  padding: 20px 0;
}

.chunk-status {
  margin-top: 16px;
  text-align: center;
  color: #909399;
}
</style>
