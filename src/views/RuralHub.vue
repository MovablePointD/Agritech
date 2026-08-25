<template>
  <div class="rural-hub">
    <AppHeader />
    <div class="hub-container">
      <div class="hub-header">
        <h1>🌾 农村信息中心</h1>
        <p>汇聚农村基础信息、政策公告与事务处理的一站式平台</p>
      </div>

      <el-tabs v-model="activeTab" class="hub-tabs">
        <!-- ======== Tab 1: 农村信息介绍 ======== -->
        <el-tab-pane label="农村信息" name="info">
          <div class="tab-toolbar">
            <el-input v-model="infoKeyword" placeholder="搜索农村信息..." clearable @keyup.enter="fetchInfoList" style="width: 260px;">
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
            <el-select v-model="infoType" placeholder="信息类型" clearable @change="fetchInfoList" style="width: 140px; margin-left: 10px;">
              <el-option v-for="(label, value) in infoTypeMap" :key="value" :label="label" :value="value" />
            </el-select>
            <el-button v-if="canManage" type="primary" @click="$router.push('/rich-editor/rural-info')" style="margin-left: auto;">
              发布信息
            </el-button>
          </div>
          <div v-loading="infoLoading" class="info-grid">
            <div v-if="infoList.length === 0" class="empty">暂无农村信息</div>
            <div v-for="item in infoList" :key="'i'+item.id" class="info-card" @click="$router.push(`/rural-info/${item.id}`)">
              <div class="card-img" v-if="firstImg(item.images)">
                <img :src="firstImg(item.images)" :alt="item.title" />
              </div>
              <div class="card-body">
                <div class="card-tags">
                  <el-tag size="small" type="success">{{ infoTypeMap[item.type] || item.type }}</el-tag>
                  <span v-if="item.address" class="card-addr"><el-icon><Location /></el-icon>{{ item.address }}</span>
                </div>
                <h3>{{ item.title }}</h3>
                <p class="card-desc">{{ stripMd(item.content).substring(0, 100) }}...</p>
                <div class="card-foot">
                  <span><el-icon><View /></el-icon>{{ item.views }}</span>
                  <span>{{ formatTime(item.created_at) }}</span>
                </div>
              </div>
            </div>
          </div>
          <el-pagination v-if="infoTotal > infoPageSize" class="pag" v-model:current-page="infoPage" :page-size="infoPageSize" :total="infoTotal" layout="prev,pager,next" @current-change="fetchInfoList" small />
        </el-tab-pane>

        <!-- ======== Tab 2: 政策公告 ======== -->
        <el-tab-pane label="政策公告" name="policy">
          <div class="tab-toolbar">
            <el-input v-model="policyKeyword" placeholder="搜索政策..." clearable @keyup.enter="fetchPolicyList" style="width: 260px;">
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
            <el-select v-model="policyCategory" placeholder="政策分类" clearable @change="fetchPolicyList" style="width: 140px; margin-left: 10px;">
              <el-option v-for="(label, value) in policyCategoryMap" :key="value" :label="label" :value="value" />
            </el-select>
            <el-button v-if="canManage" type="primary" @click="$router.push('/rich-editor/policy-notice')" style="margin-left: auto;">
              发布公告
            </el-button>
          </div>
          <div v-loading="policyLoading">
            <div v-if="topPolicies.length > 0" class="top-sec">
              <div class="sec-title">📌 置顶公告</div>
              <div v-for="item in topPolicies" :key="'t'+item.id" class="policy-item top" @click="$router.push(`/policy/${item.id}`)">
                <span class="p-icon">{{ catIcon(item.category) }}</span>
                <div class="p-body">
                  <div class="p-head"><el-tag size="small" type="danger" effect="dark">{{ policyCategoryMap[item.category]||item.category }}</el-tag><span class="p-dept" v-if="item.publish_dept">{{item.publish_dept}}</span></div>
                  <h4>{{item.title}}</h4>
                  <div class="p-meta"><span><el-icon><View /></el-icon>{{item.views}}</span><span>{{formatTime(item.publish_date||item.created_at)}}</span></div>
                </div>
              </div>
            </div>
            <div v-if="policyList.length === 0 && topPolicies.length === 0" class="empty">暂无政策公告</div>
            <div v-for="item in policyList" :key="'p'+item.id" class="policy-item" @click="$router.push(`/policy/${item.id}`)">
              <span class="p-icon">{{ catIcon(item.category) }}</span>
              <div class="p-body">
                <div class="p-head"><el-tag size="small" :type="catTag(item.category)">{{ policyCategoryMap[item.category]||item.category }}</el-tag><span class="p-dept" v-if="item.publish_dept">{{item.publish_dept}}</span></div>
                <h4>{{item.title}}</h4>
                <p class="p-desc">{{ stripMd(item.content).substring(0, 100) }}...</p>
                <div class="p-meta"><span><el-icon><View /></el-icon>{{item.views}}</span><span>{{formatTime(item.publish_date||item.created_at)}}</span></div>
              </div>
            </div>
          </div>
          <el-pagination v-if="policyTotal > policyPageSize" class="pag" v-model:current-page="policyPage" :page-size="policyPageSize" :total="policyTotal" layout="prev,pager,next" @current-change="fetchPolicyList" small />
        </el-tab-pane>

        <!-- ======== Tab 3: 农村事务 ======== -->
        <el-tab-pane label="农村事务" name="affair">
          <div class="tab-toolbar">
            <el-input v-model="affairKeyword" placeholder="搜索事务..." clearable @keyup.enter="fetchAffairList" style="width: 260px;">
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
            <el-select v-model="affairType" placeholder="事务类型" clearable @change="fetchAffairList" style="width: 140px; margin-left: 10px;">
              <el-option v-for="(l,v) in affairTypeMap" :key="v" :label="l" :value="v" />
            </el-select>
            <el-button type="success" @click="$router.push('/affair/submit')" style="margin-left: auto;">发起事务</el-button>
          </div>
          <div v-loading="affairLoading">
            <div v-if="affairList.length === 0" class="empty">暂无农村事务</div>
            <div v-for="item in affairList" :key="'a'+item.id" class="affair-item" @click="$router.push(`/affair/${item.id}`)">
              <div class="a-left"><span class="a-type-icon">{{ affairTypeIcon(item.type) }}</span></div>
              <div class="a-body">
                <div class="a-head">
                  <el-tag size="small" :type="affairStatusTag(item.status)">{{ affairStatusMap[item.status] || '未知' }}</el-tag>
                  <el-tag size="small" type="info">{{ affairTypeMap[item.type]||item.type }}</el-tag>
                  <span class="a-addr" v-if="item.address"><el-icon><Location /></el-icon>{{item.address}}</span>
                </div>
                <h4>{{item.title}}</h4>
                <p class="a-desc">{{ stripMd(item.content).substring(0, 100) }}...</p>
                <div class="a-meta">
                  <span>提交人：{{ item.user?.nickname||item.user?.username||'匿名' }}</span>
                  <span>{{ formatTime(item.created_at) }}</span>
                </div>
              </div>
            </div>
          </div>
          <el-pagination v-if="affairTotal > affairPageSize" class="pag" v-model:current-page="affairPage" :page-size="affairPageSize" :total="affairTotal" layout="prev,pager,next" @current-change="fetchAffairList" small />
        </el-tab-pane>
      </el-tabs>
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
const canManage = computed(() => {
  const r = user.value?.role
  return r === 'admin' || r === 'sysadmin' || r === 'processor'
})

const activeTab = ref('info')

// ===== 农村信息 =====
const infoList = ref([]); const infoLoading = ref(false)
const infoPage = ref(1); const infoPageSize = ref(12); const infoTotal = ref(0)
const infoKeyword = ref(''); const infoType = ref('')
const infoTypeMap = { overview:'综合概况',village_intro:'村情介绍',resource:'农业资源',culture:'乡村文化',transportation:'交通设施',education:'教育医疗',other:'其他信息' }

// ===== 政策公告 =====
const policyList = ref([]); const topPolicies = ref([]); const policyLoading = ref(false)
const policyPage = ref(1); const policyPageSize = ref(10); const policyTotal = ref(0)
const policyKeyword = ref(''); const policyCategory = ref('')
const policyCategoryMap = { subsidy:'补贴政策',land:'土地政策',environmental:'环保政策',technology:'科技政策',comprehensive:'综合政策',healthcare:'医疗养老',education:'教育政策',other:'其他政策' }

// ===== 农村事务 =====
const affairList = ref([]); const affairLoading = ref(false)
const affairPage = ref(1); const affairPageSize = ref(10); const affairTotal = ref(0)
const affairKeyword = ref(''); const affairType = ref('')
const affairTypeMap = { infrastructure:'基础设施',hardware:'硬件问题',env:'环境问题',safety:'安全隐患',other:'其他' }
const affairStatusMap = { 1:'待审核',2:'审核通过',3:'审核不通过',4:'处理中',5:'已完成' }

// helpers
const firstImg = (s) => { try { const a=JSON.parse(s);return a[0]||'' } catch { return'' } }
const stripMd = (t) => t ? t.replace(/[#*`>\[\]()!\-_~]/g,'').replace(/\n/g,' ') : ''
const formatTime = (t) => t ? new Date(t).toLocaleDateString('zh-CN') : ''
const catIcon = (c) => ({ subsidy:'💰',land:'🏞️',environmental:'🌿',technology:'🔬',comprehensive:'📊',healthcare:'🏥',education:'📚' }[c]||'📄')
const catTag = (c) => ({ subsidy:'warning',environmental:'success',technology:'primary' }[c]||'')
const affairTypeIcon = (t) => ({ infrastructure:'🏗️',hardware:'🔧',env:'🌍',safety:'🛡️',other:'📋' }[t]||'📋')
const affairStatusTag = (s) => ({ 1:'warning',2:'primary',3:'danger',4:'info',5:'success' }[s]||'')

// fetch APIs
const fetchInfoList = async () => {
  infoLoading.value = true
  try {
    const r = await api.getRuralInfos({ page:infoPage.value,page_size:infoPageSize.value,type:infoType.value,keyword:infoKeyword.value })
    infoList.value=r.list||[];infoTotal.value=r.total||0
  } catch { ElMessage.error('加载失败') }
  finally { infoLoading.value = false }
}
const fetchPolicyList = async () => {
  policyLoading.value = true
  try {
    const r = await api.getPolicyNotices({ page:policyPage.value,page_size:policyPageSize.value,category:policyCategory.value,keyword:policyKeyword.value })
    const all=r.list||[]
    topPolicies.value=all.filter(i=>i.is_top);policyList.value=all.filter(i=>!i.is_top);policyTotal.value=r.total||0
  } catch { ElMessage.error('加载失败') }
  finally { policyLoading.value = false }
}
const fetchAffairList = async () => {
  affairLoading.value = true
  try {
    const r = await api.getRuralAffairs({ page:affairPage.value,page_size:affairPageSize.value,type:affairType.value,keyword:affairKeyword.value,status:'2' })
    affairList.value=r.list||[];affairTotal.value=r.total||0
  } catch { ElMessage.error('加载失败') }
  finally { affairLoading.value = false }
}

onMounted(() => { fetchInfoList(); fetchPolicyList(); fetchAffairList() })
</script>

<style scoped>
.rural-hub { min-height:100vh;background:#f5f7fa; }
.hub-container { max-width:1200px;margin:0 auto;padding:30px 20px; }
.hub-header { text-align:center;margin-bottom:24px; }
.hub-header h1 { font-size:28px;color:#2e7d32;margin-bottom:6px; }
.hub-header p { color:#888;font-size:14px; }
.hub-tabs { margin-top:10px; }
.tab-toolbar { display:flex;align-items:center;gap:8px;margin-bottom:16px;flex-wrap:wrap; }

/* info grid */
.info-grid { display:grid;grid-template-columns:repeat(auto-fill,minmax(320px,1fr));gap:18px; }
.info-card { background:#fff;border-radius:10px;overflow:hidden;cursor:pointer;box-shadow:0 2px 10px rgba(0,0,0,0.05);transition:all .2s; }
.info-card:hover { transform:translateY(-3px);box-shadow:0 5px 18px rgba(0,0,0,0.1); }
.card-img { height:170px;overflow:hidden;background:#e8f5e9; }
.card-img img { width:100%;height:100%;object-fit:cover; }
.card-body { padding:14px; }
.card-tags { display:flex;align-items:center;gap:8px;margin-bottom:6px; }
.card-addr { font-size:12px;color:#999;display:flex;align-items:center;gap:2px; }
.card-body h3 { font-size:16px;margin:0 0 6px;color:#333; }
.card-desc { font-size:13px;color:#777;line-height:1.5;margin-bottom:10px; }
.card-foot { display:flex;justify-content:space-between;font-size:12px;color:#bbb; }

/* policy */
.top-sec { background:#fff3e0;padding:14px;border-radius:10px;margin-bottom:18px;border:1px solid #ffe0b2; }
.policy-item { display:flex;gap:12px;padding:14px;background:#fff;border-radius:10px;margin-bottom:10px;cursor:pointer;box-shadow:0 1px 6px rgba(0,0,0,0.03);transition:box-shadow .2s; }
.policy-item:hover { box-shadow:0 3px 12px rgba(0,0,0,0.08); }
.policy-item.top { background:#fff8e1;border:1px solid #ffecb3; }
.p-icon { font-size:26px;width:36px;text-align:center;flex-shrink:0; }
.p-body { flex:1;min-width:0; }
.p-head { display:flex;align-items:center;gap:8px;margin-bottom:4px; }
.p-dept { font-size:12px;color:#999; }
.p-body h4 { margin:0 0 4px;font-size:15px;color:#333; }
.p-desc { font-size:13px;color:#888;margin-bottom:6px; }
.p-meta { display:flex;gap:14px;font-size:12px;color:#bbb; }

/* affair */
.affair-item { display:flex;gap:12px;padding:14px;background:#fff;border-radius:10px;margin-bottom:10px;cursor:pointer;box-shadow:0 1px 6px rgba(0,0,0,0.03);transition:box-shadow .2s; }
.affair-item:hover { box-shadow:0 3px 12px rgba(0,0,0,0.08); }
.a-left { width:36px;text-align:center;flex-shrink:0; }
.a-type-icon { font-size:26px; }
.a-body { flex:1;min-width:0; }
.a-head { display:flex;align-items:center;gap:8px;margin-bottom:4px; }
.a-addr { font-size:12px;color:#999;display:flex;align-items:center;gap:2px; }
.a-body h4 { margin:0 0 4px;font-size:15px;color:#333; }
.a-desc { font-size:13px;color:#888;margin-bottom:6px; }
.a-meta { display:flex;gap:14px;font-size:12px;color:#bbb; }

/* common */
.sec-title { font-size:15px;font-weight:600;color:#555;margin-bottom:10px; }
.empty { text-align:center;padding:50px;color:#bbb; }
.pag { display:flex;justify-content:center;margin-top:20px; }
</style>
