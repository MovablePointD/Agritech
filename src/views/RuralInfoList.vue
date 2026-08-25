<template>
  <div class="rural-info-page">
    <AppHeader />
    <div class="container main">
      <div class="page-header">
        <h1>🌾 农村信息介绍</h1>
        <p class="subtitle">了解乡村风貌、农业资源与发展动态</p>
        <div class="header-actions">
          <el-input v-model="keyword" placeholder="搜索信息..." clearable @keyup.enter="search" style="width: 280px;">
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-select v-model="filterType" placeholder="信息类型" clearable @change="fetchData" style="width: 150px; margin-left: 10px;">
            <el-option v-for="(label, value) in typeMap" :key="value" :label="label" :value="value" />
          </el-select>
          <el-button v-if="isAdmin" type="primary" @click="$router.push('/rich-editor/rural-info')" style="margin-left: auto;">
            发布信息
          </el-button>
        </div>
      </div>

      <div v-loading="loading" class="info-grid">
        <div v-if="list.length === 0" class="empty">暂无农村信息</div>
        <div v-for="item in list" :key="item.id" class="info-card" @click="$router.push(`/rural-info/${item.id}`)">
          <div class="card-image" v-if="firstImage(item.images)">
            <img :src="firstImage(item.images)" :alt="item.title" />
          </div>
          <div class="card-body">
            <div class="card-tags">
              <el-tag size="small" type="success">{{ typeMap[item.type] || item.type }}</el-tag>
              <span v-if="item.address" class="card-address">
                <el-icon><Location /></el-icon> {{ item.address }}
              </span>
            </div>
            <h3 class="card-title">{{ item.title }}</h3>
            <p class="card-desc">{{ stripMarkdown(item.content).substring(0, 120) }}...</p>
            <div class="card-footer">
              <span><el-icon><View /></el-icon> {{ item.views }}</span>
              <span>{{ formatTime(item.created_at) }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="pagination-wrap" v-if="total > pageSize">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="prev, pager, next"
          @current-change="fetchData"
        />
      </div>
    </div>
    <AppFooter />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Location, View } from '@element-plus/icons-vue'
import AppHeader from '@/components/AppHeader.vue'
import AppFooter from '@/components/AppFooter.vue'
import api from '@/utils/api'
import { getUserInfo } from '@/utils/auth'

const user = computed(() => getUserInfo())
const isAdmin = computed(() => user.value?.role === 'admin' || user.value?.role === 'sysadmin')

const list = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(12)
const total = ref(0)
const keyword = ref('')
const filterType = ref('')

const typeMap = {
  overview: '综合概况', village_intro: '村情介绍', resource: '农业资源',
  culture: '乡村文化', transportation: '交通设施', education: '教育医疗', other: '其他信息'
}

const firstImage = (images) => {
  if (!images) return ''
  try { const arr = JSON.parse(images); return arr[0] || '' } catch { return '' }
}

const stripMarkdown = (text) => {
  if (!text) return ''
  return text.replace(/[#*`>\[\]()!\-_~]/g, '').replace(/\n/g, ' ')
}

const formatTime = (t) => t ? new Date(t).toLocaleDateString('zh-CN') : ''

const search = () => { page.value = 1; fetchData() }

const fetchData = async () => {
  loading.value = true
  try {
    const res = await api.getRuralInfos({ page: page.value, page_size: pageSize.value, type: filterType.value, keyword: keyword.value })
    list.value = res.list || []
    total.value = res.total || 0
  } catch { ElMessage.error('加载失败') }
  finally { loading.value = false }
}

onMounted(fetchData)
</script>

<style scoped>
.rural-info-page { min-height: 100vh; background: #f5f7fa; }
.container { max-width: 1200px; margin: 0 auto; padding: 0 20px; }
.main { padding: 30px 20px; }
.page-header { text-align: center; margin-bottom: 30px; }
.page-header h1 { font-size: 28px; color: #2e7d32; margin-bottom: 8px; }
.subtitle { color: #888; font-size: 14px; margin-bottom: 20px; }
.header-actions { display: flex; align-items: center; justify-content: center; flex-wrap: wrap; gap: 10px; }
.info-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap: 20px; }
.info-card { background: #fff; border-radius: 12px; overflow: hidden; cursor: pointer; box-shadow: 0 2px 12px rgba(0,0,0,0.06); transition: transform 0.2s, box-shadow 0.2s; }
.info-card:hover { transform: translateY(-4px); box-shadow: 0 6px 20px rgba(0,0,0,0.1); }
.card-image { height: 180px; overflow: hidden; background: #e8f5e9; }
.card-image img { width: 100%; height: 100%; object-fit: cover; }
.card-body { padding: 16px; }
.card-tags { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.card-address { font-size: 12px; color: #999; display: flex; align-items: center; gap: 2px; }
.card-title { font-size: 17px; margin: 0 0 8px; color: #333; line-height: 1.4; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.card-desc { font-size: 13px; color: #777; line-height: 1.6; margin-bottom: 12px; }
.card-footer { display: flex; justify-content: space-between; font-size: 12px; color: #bbb; }
.empty { grid-column: 1 / -1; text-align: center; padding: 60px; color: #bbb; font-size: 15px; }
.pagination-wrap { display: flex; justify-content: center; margin-top: 30px; }
</style>
