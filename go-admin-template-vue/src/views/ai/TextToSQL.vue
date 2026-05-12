<template>
    <div class="text2sql-container">
        <!-- 页面标题 -->
        <div class="page-header">
            <h2>智能数据查询</h2>
            <p class="subtitle">用自然语言提问，AI 自动生成 SQL 并返回结果</p>
        </div>

        <!-- 输入区 -->
        <el-card class="input-card" shadow="hover">
            <div class="input-area">
                <el-input
                    v-model="question"
                    type="textarea"
                    :rows="3"
                    placeholder="例如：今天新增了多少用户？各角色分别有多少人？"
                    maxlength="500"
                    show-word-limit
                    @keydown.ctrl.enter="handleQuery" />
                <div class="input-actions">
                    <div class="quick-questions">
                        <span class="quick-label">快捷问题：</span>
                        <el-tag
                            v-for="q in quickQuestions"
                            :key="q"
                            class="quick-tag"
                            @click="question = q"
                            effect="plain"
                            style="cursor: pointer;">{{ q }}</el-tag>
                    </div>
                    <el-button
                        type="primary"
                        :loading="loading"
                        @click="handleQuery"
                        :disabled="!question.trim()">
                        <el-icon><Search /></el-icon>
                        查询（Ctrl+Enter）
                    </el-button>
                </div>
            </div>
        </el-card>

        <!-- SQL 展示区 -->
        <el-card v-if="result.sql" class="sql-card" shadow="hover">
            <template #header>
                <span>生成的 SQL</span>
                <el-tag size="small" style="margin-left: 8px;">只读，已通过安全校验</el-tag>
            </template>
            <pre class="sql-block">{{ result.sql }}</pre>
        </el-card>

        <!-- 结果区 -->
        <el-card v-if="result.rows !== null" class="result-card" shadow="hover">
            <template #header>
                <span>查询结果</span>
                <span class="result-meta">共 {{ result.rows.length }} 条记录</span>
            </template>

            <!-- 数字卡片 -->
            <div v-if="result.chartType === 'number'" class="number-result">
                <div class="number-value">{{ result.rows[0][0] }}</div>
                <div class="number-label">{{ result.columns[0] }}</div>
            </div>

            <!-- 表格 -->
            <div v-else-if="result.chartType === 'table'">
                <el-table :data="tableData" stripe border style="width: 100%">
                    <el-table-column
                        v-for="col in result.columns"
                        :key="col"
                        :prop="col"
                        :label="col"
                        min-width="120" />
                </el-table>
            </div>

            <!-- ECharts 图表 -->
            <div v-else class="chart-container">
                <v-chart :option="chartOption" autoresize style="height: 360px;" />
            </div>
        </el-card>

        <!-- 空状态 -->
        <el-empty v-if="!loading && result.rows === null && !result.sql" description="输入问题后点击查询" />
    </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { use } from 'echarts/core'
import { SVGRenderer } from 'echarts/renderers'
import { BarChart, LineChart, PieChart } from 'echarts/charts'
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { text2sql } from '@/api/ai'

use([SVGRenderer, BarChart, LineChart, PieChart, GridComponent, LegendComponent, TooltipComponent])

const question = ref('')
const loading = ref(false)
const result = ref({
    sql: '',
    columns: [],
    rows: null,
    chartType: ''
})

const quickQuestions = [
    '总共有多少用户？',
    '各角色分别有多少用户？',
    '状态正常的用户有几个？',
    '最近注册的10个用户是谁？'
]

// Table data: convert rows array to array of objects keyed by column name
const tableData = computed(() => {
    if (!result.value.rows || !result.value.columns) return []
    return result.value.rows.map(row => {
        const obj = {}
        result.value.columns.forEach((col, i) => {
            obj[col] = row[i]
        })
        return obj
    })
})

// ECharts option, built from result
const chartOption = computed(() => {
    const { chartType, columns, rows } = result.value
    if (!rows || rows.length === 0) return {}

    if (chartType === 'pie') {
        return {
            tooltip: { trigger: 'item' },
            legend: { top: '5%', left: 'center' },
            series: [{
                type: 'pie',
                radius: ['40%', '70%'],
                data: rows.map(r => ({ name: String(r[0]), value: r[1] })),
                label: { formatter: '{b}: {c} ({d}%)' }
            }]
        }
    }

    if (chartType === 'bar') {
        return {
            tooltip: { trigger: 'axis' },
            xAxis: { type: 'category', data: rows.map(r => String(r[0])) },
            yAxis: { type: 'value' },
            series: [{ type: 'bar', data: rows.map(r => r[1]), name: columns[1] }]
        }
    }

    if (chartType === 'line') {
        return {
            tooltip: { trigger: 'axis' },
            xAxis: { type: 'category', data: rows.map(r => String(r[0])) },
            yAxis: { type: 'value' },
            series: [{ type: 'line', data: rows.map(r => r[1]), smooth: true, name: columns[1] }]
        }
    }

    return {}
})

async function handleQuery() {
    const q = question.value.trim()
    if (!q) return

    loading.value = true
    result.value = { sql: '', columns: [], rows: null, chartType: '' }

    try {
        const res = await text2sql({ question: q })
        result.value = {
            sql: res.data.sql,
            columns: res.data.columns || [],
            rows: res.data.rows || [],
            chartType: res.data.chartType
        }
    } catch (err) {
        ElMessage.error(typeof err === 'string' ? err : '查询失败，请重试')
    } finally {
        loading.value = false
    }
}
</script>

<style lang="scss" scoped>
.text2sql-container {
    padding: 20px;
    max-width: 1100px;
    margin: 0 auto;
}

.page-header {
    margin-bottom: 20px;
    h2 { margin: 0 0 4px; font-size: 22px; }
    .subtitle { color: #909399; margin: 0; font-size: 14px; }
}

.input-card { margin-bottom: 16px; }

.input-area {
    display: flex;
    flex-direction: column;
    gap: 12px;
}

.input-actions {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 8px;
}

.quick-questions {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 6px;
}

.quick-label {
    font-size: 13px;
    color: #909399;
}

.quick-tag {
    &:hover { color: var(--el-color-primary); border-color: var(--el-color-primary); }
}

.sql-card {
    margin-bottom: 16px;
    .sql-block {
        background: #1e1e2e;
        color: #cdd6f4;
        padding: 14px 16px;
        border-radius: 6px;
        font-size: 13px;
        line-height: 1.6;
        overflow-x: auto;
        margin: 0;
        white-space: pre-wrap;
        word-break: break-all;
    }
}

.result-card {
    .result-meta {
        float: right;
        font-size: 13px;
        color: #909399;
    }
}

.number-result {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 40px 0;
    .number-value {
        font-size: 64px;
        font-weight: 700;
        color: var(--el-color-primary);
        line-height: 1;
    }
    .number-label {
        font-size: 16px;
        color: #909399;
        margin-top: 12px;
    }
}

.chart-container {
    width: 100%;
}
</style>
