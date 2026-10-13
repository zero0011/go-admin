<template>
    <div class="message-bell">
        <el-popover placement="bottom-end"
                    :width="380"
                    trigger="click"
                    @show="handleShow">
            <template #reference>
                <el-badge :value="unreadCount"
                          :hidden="unreadCount === 0"
                          :max="99"
                          class="bell-badge">
                    <svg-icon icon-class="bell"
                              class="bell-icon" />
                </el-badge>
            </template>
            <div class="message-panel">
                <div class="panel-header">
                    <span class="panel-title">消息通知</span>
                    <el-button link
                               type="primary"
                               :disabled="messageStore.unreadCount === 0"
                               @click="handleReadAll">全部已读</el-button>
                </div>
                <el-tabs v-model="activeTab"
                         @tab-change="reload">
                    <el-tab-pane label="全部"
                                 name="all" />
                    <el-tab-pane :label="unreadTabLabel"
                                 name="unread" />
                </el-tabs>
                <div v-loading="loading"
                     class="message-list">
                    <div v-for="item in list"
                         :key="item.id"
                         class="message-item"
                         :class="{ 'is-unread': !item.is_read }"
                         @click="handleItemClick(item)">
                        <div class="item-title">
                            <span v-if="!item.is_read"
                                  class="unread-dot" />{{ item.title }}
                        </div>
                        <div class="item-content">{{ item.content }}</div>
                        <div class="item-time">{{ item.publish_time }}</div>
                    </div>
                    <el-empty v-if="list.length === 0"
                              description="暂无消息"
                              :image-size="60" />
                </div>
                <div class="panel-footer">
                    <el-pagination small
                                   background
                                   layout="prev, pager, next"
                                   :total="total"
                                   :page-size="pageSize"
                                   :current-page="page"
                                   @current-change="handlePageChange" />
                </div>
            </div>
        </el-popover>
    </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { getMessageList } from '@/api/message'
import useMessageStore from '@/store/message'

const messageStore = useMessageStore()
// 未读数徽标（store 响应式引用，SSE/轮询更新后红点即时变化）
const { unreadCount } = storeToRefs(messageStore)

// data
const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const activeTab = ref('all')

const unreadTabLabel = computed(() => {
    return messageStore.unreadCount > 0 ? `未读(${messageStore.unreadCount})` : '未读'
})

// 加载消息列表
const loadList = async () => {
    loading.value = true
    try {
        const params = { page: page.value, page_size: pageSize.value }
        if (activeTab.value === 'unread') {
            params.is_read = 0
        }
        const res = await getMessageList(params)
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

// 每次打开面板时刷新
const handleShow = () => {
    reload()
}

const handlePageChange = p => {
    page.value = p
    loadList()
}

// 点击消息：未读则标记已读后刷新
const handleItemClick = async item => {
    if (item.is_read) {
        return
    }
    await messageStore.markRead(item.id)
    loadList()
}

// 全部已读
const handleReadAll = async () => {
    await messageStore.markAllRead()
    reload()
}
</script>

<style lang="scss" scoped>
.message-bell {
    display: flex;
    align-items: center;
    height: 100%;
    padding: 0 8px;
    cursor: pointer;

    &:hover {
        background: rgba(0, 0, 0, 0.025);
    }

    .bell-badge {
        display: flex;
        align-items: center;
    }

    .bell-icon {
        width: 18px;
        height: 18px;
        color: #5a5e66;
    }
}

.message-panel {
    .panel-header {
        display: flex;
        justify-content: space-between;
        align-items: center;

        .panel-title {
            font-size: 15px;
            font-weight: 600;
            color: #303133;
        }
    }

    .message-list {
        min-height: 120px;
        max-height: 360px;
        overflow-y: auto;

        .message-item {
            padding: 10px 4px;
            border-bottom: 1px solid #ebeef5;
            cursor: pointer;

            &:hover {
                background: #f5f7fa;
            }

            .item-title {
                display: flex;
                align-items: center;
                font-size: 14px;
                font-weight: 600;
                color: #303133;

                .unread-dot {
                    width: 8px;
                    height: 8px;
                    margin-right: 6px;
                    border-radius: 50%;
                    background: #f56c6c;
                    flex-shrink: 0;
                }
            }

            .item-content {
                margin-top: 4px;
                font-size: 13px;
                color: #606266;
                display: -webkit-box;
                -webkit-line-clamp: 2;
                -webkit-box-orient: vertical;
                overflow: hidden;
                word-break: break-all;
            }

            .item-time {
                margin-top: 4px;
                font-size: 12px;
                color: #909399;
            }
        }
    }

    .panel-footer {
        display: flex;
        justify-content: center;
        padding-top: 8px;
    }
}
</style>
