<template>
  <div class="info-detail-page">
    <AppHeader />
    <div class="container main" v-loading="loading">
      <el-card v-if="info" class="detail-card">
        <div class="breadcrumb">
          <el-button text @click="$router.push('/rural-info')">
            <el-icon><ArrowLeft /></el-icon> 返回信息中心
          </el-button>
        </div>
        <div class="detail-header">
          <el-tag type="success">{{ typeMap[info.type] || info.type }}</el-tag>
          <h1>{{ info.title }}</h1>
          <div class="meta">
            <span v-if="info.address"><el-icon><Location /></el-icon> {{ info.address }}</span>
            <span><el-icon><View /></el-icon> {{ info.views }} 次浏览</span>
            <span>{{ formatTime(info.created_at) }}</span>
          </div>
          <div class="author" v-if="info.user">
            <el-avatar :size="32">{{ info.user?.username?.[0] }}</el-avatar>
            <span>{{ info.user?.username }}</span>
          </div>
        </div>
        <div class="detail-content" v-html="formatContent(info.content)"></div>
        <div class="images-section" v-if="imageList.length > 0">
          <el-image v-for="(img, i) in imageList" :key="i" :src="img" fit="cover"
            :preview-src-list="imageList" :initial-index="i" class="content-image" />
        </div>

        <!-- 关联公告与事务（基于地址自动匹配） -->
        <div v-if="info.address && (associatedPolicies.length > 0 || associatedAffairs.length > 0)" class="associated-section">
          <el-divider>🔗 该地区关联信息</el-divider>
          <p class="associated-hint">系统根据地址"<strong>{{ info.address }}</strong>"自动匹配以下关联内容</p>
          
          <!-- 关联政策公告 -->
          <div v-if="associatedPolicies.length > 0" class="associated-block">
            <h4 class="associated-title">📜 同地区政策公告</h4>
            <div v-for="p in associatedPolicies" :key="'ap'+p.id" class="associated-item" @click="$router.push(`/policy/${p.id}`)">
              <div class="ai-left">
                <span class="ai-icon">{{ policyIcon(p.category) }}</span>
              </div>
              <div class="ai-body">
                <div class="ai-head">
                  <el-tag size="small" :type="policyTagType(p.category)">{{ policyCategoryMap[p.category] || p.category }}</el-tag>
                  <span class="ai-dept" v-if="p.publish_dept">{{ p.publish_dept }}</span>
                  <el-tag v-if="p.is_top" size="small" type="danger" effect="dark">置顶</el-tag>
                </div>
                <h5>{{ p.title }}</h5>
                <div class="ai-meta">
                  <span><el-icon><View /></el-icon>{{ p.views }}</span>
                  <span>{{ formatTime(p.publish_date || p.created_at) }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 关联农村事务 -->
          <div v-if="associatedAffairs.length > 0" class="associated-block">
            <h4 class="associated-title">🔧 同地区农村事务</h4>
            <div v-for="a in associatedAffairs" :key="'aa'+a.id" class="associated-item" @click="$router.push(`/affair/${a.id}`)">
              <div class="ai-left">
                <span class="ai-icon">{{ affairIcon(a.type) }}</span>
              </div>
              <div class="ai-body">
                <div class="ai-head">
                  <el-tag size="small" :type="affairStatusTag(a.status)">{{ affairStatusMap[a.status] || '未知' }}</el-tag>
                  <el-tag size="small" type="info">{{ affairTypeMap[a.type] || a.type }}</el-tag>
                </div>
                <h5>{{ a.title }}</h5>
                <div class="ai-meta">
                  <span>提交人：{{ a.user?.nickname || a.user?.username || '匿名' }}</span>
                  <span>{{ formatTime(a.created_at) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 评论区 -->
        <el-divider />
        <div class="comments-section">
          <h4>评论 ({{ totalComments }})</h4>
          <div class="comment-input" v-if="token">
            <el-input v-model="commentContent" type="textarea" :rows="2" placeholder="写下你的评论..." />
            <el-button type="primary" size="small" @click="submitComment" :disabled="!commentContent.trim()" style="margin-top:8px;">发表评论</el-button>
          </div>
          <div v-else class="login-tip"><el-link type="primary" @click="$router.push('/login')">登录</el-link> 后参与评论</div>
          <div v-if="comments.length > 0" class="comment-list">
            <div v-for="c in comments" :key="c.id" class="comment-item" :class="'level-'+c.level">
              <el-avatar :size="c.level===1?36:28" class="c-avatar">{{ c.user?.username?.[0] }}</el-avatar>
              <div class="c-body">
                <div class="c-header">
                  <span class="c-name">{{ c.user?.username||'匿名' }}</span>
                  <span class="c-time">{{ formatTime(c.created_at) }}</span>
                </div>
                <div class="c-content">{{ c.content }}</div>
                <div class="c-actions">
                  <el-button size="small" text @click="toggleReply(c)">{{ replyingTo===c.id?'取消回复':'回复' }}</el-button>
                  <el-button size="small" text @click="likeComment(c)"><el-icon><Star /></el-icon> {{ c.likes||0 }}</el-button>
                </div>
                <div v-if="replyingTo===c.id" class="reply-box">
                  <el-input v-model="replyContent" :placeholder="'回复 '+c.user?.username" size="small" />
                  <el-button size="small" type="primary" @click="submitReply(c)" style="margin-top:4px;">回复</el-button>
                </div>
                <div v-if="c.children?.length" class="child-comments">
                  <div v-for="sub in c.children" :key="sub.id" class="comment-item child">
                    <el-avatar :size="24">{{ sub.user?.username?.[0] }}</el-avatar>
                    <div class="c-body">
                      <div class="c-header"><span class="c-name">{{ sub.user?.username }}</span><span class="c-time">{{ formatTime(sub.created_at) }}</span></div>
                      <div class="c-content">{{ sub.content }}</div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-card>
      <div v-else-if="!loading" class="not-found"><p>信息不存在</p><el-button type="primary" @click="$router.push('/rural-info')">返回信息中心</el-button></div>
    </div>
    <AppFooter />
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft,Location,View,Star } from '@element-plus/icons-vue'
import AppHeader from '@/components/AppHeader.vue'
import AppFooter from '@/components/AppFooter.vue'
import api from '@/utils/api'
import { getToken } from '@/utils/auth'
import { marked } from 'marked'
import hljs from 'highlight.js'
import 'highlight.js/styles/github.css'

marked.setOptions({ breaks: true, gfm: true })

const route = useRoute()
const token = computed(()=>getToken())
const info = ref(null); const loading = ref(true); const imageList = ref([])
const typeMap = { overview:'综合概况',village_intro:'村情介绍',resource:'农业资源',culture:'乡村文化',transportation:'交通设施',education:'教育医疗',other:'其他信息' }

// 地址关联数据
const associatedPolicies = ref([])
const associatedAffairs = ref([])
const policyCategoryMap = { subsidy:'补贴政策',land:'土地政策',environmental:'环保政策',technology:'科技政策',comprehensive:'综合政策',healthcare:'医疗养老',education:'教育政策',other:'其他政策' }
const affairTypeMap = { infrastructure:'基础设施',hardware:'硬件问题',env:'环境问题',safety:'安全隐患',other:'其他' }
const affairStatusMap = { 1:'待审核',2:'审核通过',3:'审核不通过',4:'处理中',5:'已完成' }
const policyIcon = (c) => ({ subsidy:'💰',land:'🏞️',environmental:'🌿',technology:'🔬',comprehensive:'📊',healthcare:'🏥',education:'📚' }[c]||'📄')
const policyTagType = (c) => ({ subsidy:'warning',environmental:'success',technology:'primary' }[c]||'')
const affairIcon = (t) => ({ infrastructure:'🏗️',hardware:'🔧',env:'🌍',safety:'🛡️',other:'📋' }[t]||'📋')
const affairStatusTag = (s) => ({ 1:'warning',2:'primary',3:'danger',4:'info',5:'success' }[s]||'')

// comments
const comments = ref([])
const commentContent = ref('')
const replyingTo = ref(null)
const replyContent = ref('')
const totalComments = computed(()=>{ let n=0; comments.value.forEach(c=>{ n++; if(c.children)n+=c.children.length }); return n })

const formatContent = (content) => {
  if (!content) return ''
  let html = marked.parse(content)
  html = html.replace(/<pre><code class="language-(\w+)">([\s\S]*?)<\/code><\/pre>/g, (match, lang, code) => {
    const decoded = code.replace(/&lt;/g,'<').replace(/&gt;/g,'>').replace(/&amp;/g,'&').replace(/&quot;/g,'"').replace(/&#39;/g,"'")
    let h; try { h=(lang&&hljs.getLanguage(lang))?hljs.highlight(decoded,{language:lang,ignoreIllegals:true}).value:hljs.highlightAuto(decoded).value } catch { h=decoded }
    return `<pre class="code-block"><code class="language-${lang} hljs">${h}</code></pre>`
  })
  return html
}

const formatTime = (t) => t ? new Date(t).toLocaleString('zh-CN') : ''

const fetchComments = async () => {
  try {
    const r = await api.getCommentRuralInfos({ target_type:'rural_info', target_id: route.params.id })
    comments.value = r.list || []
  } catch {}
}

const submitComment = async () => {
  if (!commentContent.value.trim()) return
  try {
    await api.createCommentRuralInfo({ target_type:'rural_info', target_id: Number(route.params.id), content: commentContent.value, parent_id: 0 })
    ElMessage.success('评论成功'); commentContent.value = ''; await fetchComments()
  } catch { ElMessage.error('评论失败') }
}

const toggleReply = (c) => {
  if (!token.value) { ElMessage.warning('请先登录'); return }
  replyingTo.value = replyingTo.value===c.id ? null : c.id; replyContent.value = ''
}

const submitReply = async (parent) => {
  if (!replyContent.value.trim()) return
  try {
    await api.createCommentRuralInfo({ target_type:'rural_info', target_id: Number(route.params.id), content: replyContent.value, parent_id: parent.id, reply_to_user_id: parent.user_id })
    ElMessage.success('回复成功'); replyingTo.value=null; replyContent.value=''; await fetchComments()
  } catch { ElMessage.error('回复失败') }
}

const likeComment = async (c) => {
  if (!token.value) { ElMessage.warning('请先登录'); return }
  try {
    const r = await api.likeCommentRuralInfo(c.id)
    c.likes = r.likes
  } catch { ElMessage.error('操作失败') }
}

// 获取关联的公告和事务（基于地址匹配，忽略大小写）
const fetchAssociatedItems = async (address) => {
  if (!address) return
  try {
    const res = await api.getAssociatedByAddress(address)
    associatedPolicies.value = res.policies || []
    associatedAffairs.value = res.affairs || []
  } catch {
    associatedPolicies.value = []
    associatedAffairs.value = []
  }
}

onMounted(async ()=>{
  loading.value=true
  try { const r=await api.getRuralInfo(route.params.id); info.value=r; if(r.images) try{imageList.value=JSON.parse(r.images)}catch{imageList.value=[]} } catch { info.value=null }
  loading.value=false
  fetchComments()
  if (info.value?.address) {
    fetchAssociatedItems(info.value.address)
  }
})
</script>

<style scoped>
.info-detail-page{min-height:100vh;background:#f5f7fa}.container{max-width:860px;margin:0 auto;padding:0 20px}.main{padding:30px 20px}
.detail-card{border-radius:12px}.breadcrumb{margin-bottom:15px}
.detail-header{margin-bottom:24px}.detail-header h1{font-size:26px;margin:12px 0;color:#333}
.meta{display:flex;gap:20px;font-size:13px;color:#999;margin:10px 0}.meta span{display:flex;align-items:center;gap:4px}
.author{display:flex;align-items:center;gap:8px;margin-top:12px;padding-top:12px;border-top:1px solid #f0f0f0;font-size:14px;color:#666}
.detail-content{line-height:1.9;color:#333;font-size:15px}
.detail-content :deep(p){margin-bottom:14px}.detail-content :deep(img){max-width:100%;border-radius:8px;margin:10px 0}
.detail-content :deep(ul),.detail-content :deep(ol){padding-left:24px;margin-bottom:14px}
.detail-content :deep(li){margin-bottom:6px}
.detail-content :deep(blockquote){border-left:4px solid #67c23a;background:#f0f9eb;padding:12px 16px;margin:14px 0;border-radius:0 6px 6px 0;color:#555}
.detail-content :deep(.code-block){background:#282c34;color:#abb2bf;padding:15px;border-radius:8px;overflow-x:auto;margin:10px 0}
.detail-content :deep(h2){font-size:20px;margin:20px 0 12px}.detail-content :deep(h3){font-size:17px;margin:16px 0 10px}
.images-section{display:flex;flex-wrap:wrap;gap:10px;margin-top:20px}.content-image{width:200px;height:150px;border-radius:8px;object-fit:cover}
.not-found{text-align:center;padding:60px;color:#999}

.comments-section{margin-top:10px}.comments-section h4{margin-bottom:14px;color:#333}
.login-tip{margin:10px 0;color:#999;font-size:13px}
.comment-input{margin-bottom:16px}
.comment-list{margin-top:16px}
.comment-item{display:flex;gap:10px;margin-bottom:14px}
.comment-item.child{margin-top:10px;margin-left:38px;padding:8px 10px;background:#fafafa;border-radius:8px}
.c-avatar{flex-shrink:0}.c-body{flex:1;min-width:0}
.c-header{display:flex;gap:8px;margin-bottom:4px}.c-name{font-size:13px;color:#409eff;font-weight:500}.c-time{font-size:12px;color:#bbb}
.c-content{font-size:14px;color:#444;line-height:1.6}
.c-actions{margin-top:4px}
.reply-box{margin-top:8px}
.child-comments{margin-top:8px}

/* 关联公告与事务 */
.associated-section { margin-top: 10px; }
.associated-hint { font-size: 13px; color: #909399; margin-bottom: 16px; line-height: 1.5; }
.associated-block { margin-bottom: 20px; }
.associated-title { font-size: 15px; font-weight: 600; color: #333; margin-bottom: 12px; display: flex; align-items: center; gap: 6px; }
.associated-item { display: flex; gap: 10px; padding: 12px 14px; background: #f9fafb; border-radius: 8px; margin-bottom: 8px; cursor: pointer; border: 1px solid #ebeef5; transition: all .2s; }
.associated-item:hover { border-color: #67c23a; background: #f0f9eb; transform: translateX(4px); }
.ai-left { width: 32px; text-align: center; flex-shrink: 0; }
.ai-icon { font-size: 22px; }
.ai-body { flex: 1; min-width: 0; }
.ai-head { display: flex; align-items: center; gap: 6px; margin-bottom: 4px; flex-wrap: wrap; }
.ai-dept { font-size: 12px; color: #999; }
.ai-body h5 { margin: 0 0 4px; font-size: 14px; color: #333; }
.ai-meta { display: flex; gap: 12px; font-size: 12px; color: #bbb; }

</style>
