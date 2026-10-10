<template>
    <div class="material-menu">
        <!-- 新建文件夹按钮 -->
        <div class="menu-header">
            <el-button type="primary" @click="handleCreateFolder" style="width: 100%">
                <el-icon><FolderAdd /></el-icon>
                新建文件夹
            </el-button>
        </div>

        <!-- 文件类型筛选 -->
        <div class="menu-section">
            <div class="section-title">文件类型</div>
            <el-radio-group v-model="localFileType" @change="handleTypeChange" class="type-group">
                <el-radio-button label="">全部</el-radio-button>
                <el-radio-button label="image">图片</el-radio-button>
                <el-radio-button label="video">视频</el-radio-button>
                <el-radio-button label="audio">音频</el-radio-button>
                <el-radio-button label="document">文档</el-radio-button>
                <el-radio-button label="other">其他</el-radio-button>
            </el-radio-group>
        </div>

        <!-- 排序方式 -->
        <div class="menu-section">
            <div class="section-title">排序方式</div>
            <el-select v-model="localSortBy" @change="handleSortChange" style="width: 100%">
                <el-option label="名称升序" value="name_asc" />
                <el-option label="名称降序" value="name_desc" />
                <el-option label="时间升序" value="time_asc" />
                <el-option label="时间降序" value="time_desc" />
                <el-option label="大小升序" value="size_asc" />
                <el-option label="大小降序" value="size_desc" />
            </el-select>
        </div>

        <!-- 搜索框 -->
        <div class="menu-section">
            <div class="section-title">搜索</div>
            <el-input
                v-model="searchKeyword"
                placeholder="搜索全部文件/文件夹"
                clearable
                @keyup.enter="handleSearch"
                @clear="handleSearch">
                <template #prefix>
                    <el-icon><Search /></el-icon>
                </template>
            </el-input>
        </div>

        <!-- 新建文件夹对话框 -->
        <el-dialog
            v-model="dialogVisible"
            title="新建文件夹"
            width="400px">
            <el-input
                v-model="newFolderName"
                placeholder="请输入文件夹名称"
                @keyup.enter="confirmCreateFolder"/>
            <template #footer>
                <el-button @click="dialogVisible = false">取消</el-button>
                <el-button type="primary" @click="confirmCreateFolder">确定</el-button>
            </template>
        </el-dialog>
    </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { FolderAdd, Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const props = defineProps({
    fileType: {
        type: String,
        default: ''
    },
    sortBy: {
        type: String,
        default: 'time_desc'
    },
    keyword: {
        type: String,
        default: ''
    }
})

const emit = defineEmits(['type-change', 'sort-change', 'create-folder', 'refresh', 'search'])

const localFileType = ref(props.fileType)
const localSortBy = ref(props.sortBy)
const searchKeyword = ref('')
const dialogVisible = ref(false)
const newFolderName = ref('')

watch(() => props.fileType, (val) => {
    localFileType.value = val
})

watch(() => props.sortBy, (val) => {
    localSortBy.value = val
})

watch(() => props.keyword, (val) => {
    searchKeyword.value = val
})

const handleTypeChange = (type) => {
    emit('type-change', type)
}

const handleSortChange = (sort) => {
    emit('sort-change', sort)
}

const handleCreateFolder = () => {
    newFolderName.value = ''
    dialogVisible.value = true
}

const confirmCreateFolder = () => {
    if (!newFolderName.value.trim()) {
        ElMessage.warning('请输入文件夹名称')
        return
    }
    emit('create-folder', newFolderName.value.trim())
    dialogVisible.value = false
}

const handleSearch = () => {
    emit('search', searchKeyword.value)
}
</script>

<style scoped>
.material-menu {
  padding: 16px;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.menu-header {
  margin-bottom: 20px;
}

.menu-section {
  margin-bottom: 20px;
}

.section-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 12px;
}

.type-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.type-group :deep(.el-radio-button) {
  width: 100%;
}

.type-group :deep(.el-radio-button__inner) {
  width: 100%;
}
</style>
