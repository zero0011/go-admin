<template>
    <div class="announcement-page">
        <el-card class="publish-card">
            <template #header>
                <span>发布公告</span>
            </template>
            <el-form ref="formRef"
                     :model="form"
                     :rules="rules"
                     label-width="80px">
                <el-form-item label="标题"
                              prop="title">
                    <el-input v-model="form.title"
                              maxlength="255"
                              show-word-limit
                              placeholder="请输入公告标题" />
                </el-form-item>
                <el-form-item label="正文"
                              prop="content">
                    <el-input v-model="form.content"
                              type="textarea"
                              :rows="6"
                              maxlength="5000"
                              show-word-limit
                              placeholder="请输入公告内容（纯文本）" />
                </el-form-item>
                <el-form-item>
                    <el-button type="primary"
                               :loading="publishing"
                               @click="handlePublish">发布</el-button>
                </el-form-item>
            </el-form>
        </el-card>

        <el-card>
            <template #header>
                <span>已发布列表</span>
            </template>
            <el-table v-loading="loading"
                      :data="list">
                <el-table-column prop="id"
                                 label="ID"
                                 width="80" />
                <el-table-column prop="title"
                                 label="标题"
                                 min-width="180"
                                 show-overflow-tooltip />
                <el-table-column prop="content"
                                 label="内容"
                                 min-width="260"
                                 show-overflow-tooltip />
                <el-table-column prop="publisher_name"
                                 label="发布人"
                                 width="120" />
                <el-table-column prop="publish_time"
                                 label="发布时间"
                                 width="180" />
            </el-table>
            <div class="pagination">
                <el-pagination background
                               layout="total, prev, pager, next"
                               :total="total"
                               :page-size="pageSize"
                               :current-page="page"
                               @current-change="handlePageChange" />
            </div>
        </el-card>
    </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { publishAnnouncement, getAnnouncementList } from '@/api/message'

// refs
const formRef = ref()

// data
const publishing = ref(false)
const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const form = reactive({
    title: '',
    content: ''
})
const rules = {
    title: [{ required: true, message: '请输入公告标题', trigger: 'blur' }],
    content: [{ required: true, message: '请输入公告内容', trigger: 'blur' }]
}

// 发布公告
const handlePublish = () => {
    formRef.value.validate(async valid => {
        if (valid) {
            publishing.value = true
            try {
                await publishAnnouncement({ title: form.title, content: form.content })
                ElMessage.success('发布成功')
                form.title = ''
                form.content = ''
                formRef.value.resetFields()
                reload()
            } finally {
                publishing.value = false
            }
        }
    })
}

// 加载已发布列表
const loadList = async () => {
    loading.value = true
    try {
        const res = await getAnnouncementList({ page: page.value, page_size: pageSize.value })
        list.value = res.data.list || []
        total.value = res.data.total
    } finally {
        loading.value = false
    }
}

const reload = () => {
    page.value = 1
    loadList()
}

const handlePageChange = p => {
    page.value = p
    loadList()
}

onMounted(() => {
    loadList()
})
</script>

<style lang="scss" scoped>
.announcement-page {
    padding: 20px;

    .publish-card {
        margin-bottom: 20px;
    }

    .pagination {
        display: flex;
        justify-content: flex-end;
        margin-top: 16px;
    }
}
</style>
