<template>
    <el-upload :action="action"
               :headers="localHeader"
               :data="data"
               :accept="accept"
               list-type="picture-card"
               :show-file-list="false"
               :on-progress="handleProgress"
               :on-success="uploadSuccess"
               :on-error="handleError">
        <!-- 上传中显示进度条 -->
        <div class="upload-progress" v-if="uploading">
            <el-progress type="circle" :percentage="progress" :width="80" />
        </div>
        <!-- 已上传显示预览 -->
        <div class="el-upload-list--picture-card"
             v-else-if="modelValue">
            <div class="el-upload-list__item">
                <!-- 视频预览 -->
                <video v-if="isVideo"
                       class="el-upload-list__item-thumbnail"
                       :src="modelValue" />
                <!-- 图片预览 -->
                <img v-else
                     class="el-upload-list__item-thumbnail"
                     :src="modelValue">
                <span class="el-upload-list__item-actions">
                    <span class="el-upload-list__item-preview"
                          @click.stop="uploadRemove()">
                        <svg-icon icon-class="delete"
                                  style="font-size: 1.5rem; fill: #ddd;" />
                    </span>
                </span>
            </div>
        </div>
        <!-- 未上传显示加号 -->
        <el-icon v-else>
            <Plus />
        </el-icon>
    </el-upload>
</template>

<script setup>
import { ref, computed } from 'vue'
import useUserStore from '@/store/user'
import { Plus } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const uStore = useUserStore()
const { token } = uStore

const modelValue = defineModel({ type: String })

// 上传状态
const uploading = ref(false)
const progress = ref(0)

// 判断是否为视频
const isVideo = computed(() => {
    if (!modelValue.value) return false
    const videoExts = ['.mp4', '.mov', '.avi', '.webm']
    return videoExts.some(ext => modelValue.value.toLowerCase().endsWith(ext))
})

const localHeader = computed(() => {
    if (Object.keys(props.headers).length == 0) {
        return { 'Authorization': token }
    }
    return props.headers
})

const props = defineProps({
    action: {
        type: String,
        default: import.meta.env.VITE_APP_BASE_API + '/upload/image'
    },
    headers: {
        type: Object,
        default: () => ({})
    },
    data: Object,
    accept: {
        type: String,
        default: 'image/png, image/jpeg, image/gif, video/mp4, video/quicktime, video/x-msvideo, video/webm'
    }
})

const emit = defineEmits(['uploadSuccess', 'uploadRemove'])

// 上传进度
const handleProgress = (event) => {
    uploading.value = true
    progress.value = Math.round(event.percent)
}

// 上传成功
const uploadSuccess = ({ data }) => {
    uploading.value = false
    progress.value = 0
    modelValue.value = data
    emit('uploadSuccess', data)
}

// 上传失败
const handleError = () => {
    uploading.value = false
    progress.value = 0
    ElMessage.error('上传失败')
}

// 删除
const uploadRemove = () => {
    modelValue.value = ''
    emit('uploadRemove')
}
</script>

<style lang="scss" scoped>
.el-upload-list__item {
    margin: 0;
}
.upload-progress {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
}
video.el-upload-list__item-thumbnail {
    object-fit: cover;
}
</style>
