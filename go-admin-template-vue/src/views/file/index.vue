<template>
    <div class="file-manager">
        <el-container style="height: 100%">
            <!-- 左侧菜单 -->
            <el-aside width="280px" class="file-menu">
                <MaterialMenu
                    :file-type="fileType"
                    :sort-by="sortBy"
                    @type-change="handleTypeChange"
                    @sort-change="handleSortChange"
                    @create-folder="handleCreateFolder"
                    @refresh="loadFileList"
                    @search="handleSearch"/>
            </el-aside>

            <!-- 右侧内容 -->
            <el-main class="file-content">
                <MaterialContent
                    :loading="loading"
                    :file-list="fileList"
                    :total="total"
                    :current-folder-id="currentFolderId"
                    :breadcrumb-list="breadcrumbList"
                    :page="page"
                    :page-size="pageSize"
                    @folder-click="handleFolderClick"
                    @file-click="handleFileClick"
                    @folder-back="handleFolderBack"
                    @folder-rename="handleFolderRename"
                    @folder-delete="handleFolderDelete"
                    @file-delete="handleFileDelete"
                    @file-download="handleFileDownload"
                    @page-change="handlePageChange"
                    @upload-success="loadFileList"/>
            </el-main>
        </el-container>
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import MaterialMenu from './MaterialMenu.vue'
import MaterialContent from './MaterialContent.vue'
import { getFileList, createFolder, renameFolder, deleteFolder, deleteFile } from '@/api/file'

// 状态
const loading = ref(false)
const fileList = ref([])
const total = ref(0)
const currentFolderId = ref(null)
const breadcrumbList = ref([{ id: null, name: '根目录' }])
const page = ref(1)
const pageSize = ref(20)
const fileType = ref('')
const sortBy = ref('time_desc')
const keyword = ref('')

// 加载文件列表
const loadFileList = async () => {
    loading.value = true
    try {
        const res = await getFileList({
            parent_folder_id: currentFolderId.value,
            file_type: fileType.value || undefined,
            sort_by: sortBy.value,
            keyword: keyword.value || undefined,
            page: page.value,
            page_size: pageSize.value
        })
        if (res.code === 0) {
            fileList.value = res.data.list
            total.value = res.data.total
        }
    } catch (error) {
        console.error('加载文件列表失败:', error)
    } finally {
        loading.value = false
    }
}

// 文件类型筛选
const handleTypeChange = (type) => {
    fileType.value = type
    page.value = 1
    loadFileList()
}

// 排序方式改变
const handleSortChange = (sort) => {
    sortBy.value = sort
    loadFileList()
}

// 文件夹点击
const handleFolderClick = (folder) => {
    currentFolderId.value = folder.folder_id
    breadcrumbList.value.push({
        id: folder.folder_id,
        name: folder.folder_name
    })
    page.value = 1
    loadFileList()
}

// 文件点击
const handleFileClick = (file) => {
    // 可以打开预览或执行其他操作
    console.log('文件点击:', file)
}

// 文件夹返回
const handleFolderBack = (index) => {
    // 截断面包屑到指定位置
    breadcrumbList.value = breadcrumbList.value.slice(0, index + 1)
    currentFolderId.value = breadcrumbList.value[breadcrumbList.value.length - 1].id
    page.value = 1
    loadFileList()
}

// 创建文件夹
const handleCreateFolder = async (folderName) => {
    try {
        const res = await createFolder({
            folder_name: folderName,
            parent_id: currentFolderId.value
        })
        if (res.code === 0) {
            loadFileList()
        }
    } catch (error) {
        console.error('创建文件夹失败:', error)
    }
}

// 重命名文件夹
const handleFolderRename = async (id, newName) => {
    try {
        const res = await renameFolder(id, { folder_name: newName })
        if (res.code === 0) {
            loadFileList()
        }
    } catch (error) {
        console.error('重命名文件夹失败:', error)
    }
}

// 删除文件夹
const handleFolderDelete = async (id) => {
    try {
        const res = await deleteFolder(id)
        if (res.code === 0) {
            loadFileList()
        }
    } catch (error) {
        console.error('删除文件夹失败:', error)
    }
}

// 删除文件
const handleFileDelete = async (id) => {
    try {
        const res = await deleteFile(id)
        if (res.code === 0) {
            loadFileList()
        }
    } catch (error) {
        console.error('删除文件失败:', error)
    }
}

// 下载文件
const handleFileDownload = (file) => {
    window.open(file.file_url, '_blank')
}

// 分页改变
const handlePageChange = (newPage) => {
    page.value = newPage
    loadFileList()
}

// 搜索
const handleSearch = (searchKeyword) => {
    keyword.value = searchKeyword
    page.value = 1
    loadFileList()
}

onMounted(() => {
    loadFileList()
})
</script>

<style scoped>
.file-manager {
  height: 100vh;
  background-color: #f5f7fa;
}

.file-menu {
  background-color: #fff;
  border-right: 1px solid #e4e7ed;
}

.file-content {
  padding: 0;
  overflow: hidden;
}
</style>
