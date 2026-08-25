<template>
  <div class="policy-page">
    <AppHeader />
    <div class="container main">
      <div class="page-header">
        <h1>📋 农村政策公告</h1>
        <p class="subtitle">最新惠农政策、法规解读与补贴申请指南</p>
        <div class="header-actions">
          <el-input v-model="keyword" placeholder="搜索政策..." clearable @keyup.enter="search" style="width: 280px;">
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-select v-model="filterCategory" placeholder="政策分类" clearable @change="fetchData" style="width: 150px; margin-left: 10px;">
            <el-option v-for="(label, value) in categoryMap" :key="value" :label="label" :value="value" />
          </el-select>
          <el-button v-if="isAdmin" type="primary" @click="$router.push('/rich-editor/policy-notice')" style="margin-left: auto;">
            发布公告
          </el-button>
        </div>
      </div>

      <div v-loading="loading">
        <!-- 置顶公告 -->
        <div v-if="topList.length > 0" class="top-section">
          <div class="section-title"><el-icon><Top /></el-icon> 置顶公告</div>
          <div v-for="item in topList" :key="item.id" class="policy-item top-item" @click="$router.push(`/policy/${item.id}`)">
            <div class="item-icon">📌</div>
            <div class="item-body">
              <div class="item-header">
                <el-tag size="small" type="danger" effect="dark">{{ categoryMap[item.category] || item.category }}</el-tag>
                <span class="item-dept" v-if="item.publish_dept">{{ item.publish_dept }}</span>
              </div>
              <h4>{{ item.title }}</h4>
              <div class="item-meta">
                <span><el-icon><View /></el-icon> {{ item.views }}</span>
                <span>{{ formatTime(item.publish_date || item.created_at) }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 普通公告列表 -->
        <div class="policy-list">
          <div class="section-title">全部公告</div>
          <div v-if="list.length === 0" class="empty">暂无政策公告</div>
          <div v-for="item in list" :key="item.id" class="policy-item" @click="$router.push(`/policy/${item.id}`)">
            <div class="item-icon">{{ categoryIcon(item.category) }}</div>
            <div class="item-body">
              <div class="item-header">
                <el-tag size="small" :type="categoryTag(item.category)">{{ categoryMap[item.category] || item.category }}</el-tag>
                <span class="item-dept" v-if="item.publish_dept">{{ item.publish_dept }}</span>
              </div>
              <h4>{{ item.title }}</h4>
              <p class="item-desc">{{ stripMarkdown(item.content).substring(0, 100) }}...</p>
              <div class="item-meta">
                <span><el-icon><View /></el-icon> {{ item.views }}</span>
                <span>{{ formatTime(item.publish_date || item.created_at) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="pagination-wrap" v-if="total > pageSize">
        <el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" layout="prev, pager, next" @current-change="fetchData" />
      </div>
    </div>
    <AppFooter />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Top, View } from '@element-plus/icons-vue'
import AppHeader from '@/components/AppHeader.vue'
import AppFooter from '@/components/AppFooter.vue'
import api from '@/utils/api'
import { getUserInfo } from '@/utils/auth'

const user = computed(() => getUserInfo())
const isAdmin = computed(() => user.value?.role === 'admin' || user.value?.role === 'sysadmin')

const list = ref([])
const topList = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const keyword = ref('')
const filterCategory = ref('')

const categoryMap = {
  subsidy: '补贴政策', land: '土地政策', environmental: '环保政策',
  technology: '科技政策', comprehensive: '综合政策', healthcare: '医疗养老', education: '教育政策', other: '其他政策'
}
const categoryIcon = (c) => ({ subsidy: '💰', land: '🏞️', environmental: '🌿', technology: '🔬', comprehensive: '📊', healthcare: '🏥', education: '📚' }[c] || '📄')
const categoryTag = (c) => ({ subsidy: 'warning', land: '', environmental: 'success', technology: 'primary', other: 'info' }[c] || '')

const stripMarkdown = (t) => t ? t.replace(/[#*`>\[\]()!\-_~]/g, '').replace(/\n/g, ' ') : ''
const formatTime = (t) => t ? new Date(t).toLocaleDateString('zh-CN') : ''
const search = () => { page.value = 1; fetchData() }

const fetchData = async () => {
  loading.value = true
  try {
    const res = await api.getPolicyNotices({ page: page.value, page_size: pageSize.value, category: filterCategory.value, keyword: keyword.value })
    const all = res.list || []
    topList.value = all.filter(item => item.is_top)
    list.value = all.filter(item => !item.is_top)
    total.value = res.total || 0
  } catch { ElMessage.error('加载失败') }
  finally { loading.value = false }
}

onMounted(fetchData)
</script>

<style scoped>
.policy-page { min-height: 100vh; background: #f5f7fa; }
.container { max-width: 900px; margin: 0 auto; padding: 0 20px; }
.main { padding: 30px 20px; }
.page-header { text-align: center; margin-bottom: 30px; }
.page-header h1 { font-size: 28px; color: #d84315; margin-bottom: 8px; }
.subtitle { color: #888; font-size: 14px; margin-bottom: 20px; }
.header-actions { display: flex; align-items: center; justify-content: center; flex-wrap: wrap; gap: 10px; }
.section-title { font-size: 16px; font-weight: 600; color: #555; margin-bottom: 12px; display: flex; align-items: center; gap: 6px; }
.top-section { margin-bottom: 24px; padding: 16px; background: #fff3e0; border-radius: 10px; border: 1px solid #ffe0b2; }
.policy-item { display: flex; gap: 14px; padding: 16px; background: #fff; border-radius: 10px; margin-bottom: 10px; cursor: pointer; transition: box-shadow 0.2s; box-shadow: 0 1px 6px rgba(0,0,0,0.04); }
.policy-item:hover { box-shadow: 0 4px 16px rgba(0,0,0,0.08); }
.top-item { background: #fff8e1; border: 1px solid #ffecb3; }
.item-icon { font-size: 28px; flex-shrink: 0; width: 40px; text-align: center; }
.item-body { flex: 1; min-width: 0; }
.item-header { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; }
.item-dept { font-size: 12px; color: #999; }
.item-body h4 { margin: 0 0 6px; font-size: 16px; color: #333; }
.item-desc { font-size: 13px; color: #888; margin-bottom: 8px; }
.item-meta { display: flex; gap: 16px; font-size: 12px; color: #bbb; }
.item-meta span { display: flex; align-items: center; gap: 3px; }
.empty { text-align: center; padding: 50px; color: #bbb; }
.pagination-wrap { display: flex; justify-content: center; margin-top: 30px; }
.policy-list { margin-top: 10px; }
</style>
