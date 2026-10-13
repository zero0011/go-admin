import request from '@/utils/request'

// ========== 消息（用户侧） ==========

// 获取消息列表
export function getMessageList(params) {
    return request({
        url: '/message/list',
        method: 'get',
        params
    })
}

// 获取未读数
export function getUnreadCount() {
    return request({
        url: '/message/unread-count',
        method: 'get'
    })
}

// 单条已读
export function markRead(data) {
    return request({
        url: '/message/read',
        method: 'post',
        data
    })
}

// 一键已读
export function markAllRead() {
    return request({
        url: '/message/read-all',
        method: 'post'
    })
}

// ========== 公告（运营侧） ==========

// 发布公告
export function publishAnnouncement(data) {
    return request({
        url: '/message/announcement',
        method: 'post',
        data
    })
}

// 已发布公告列表
export function getAnnouncementList(params) {
    return request({
        url: '/message/announcement/list',
        method: 'get',
        params
    })
}
