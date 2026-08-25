<template>
  <div class="policy-detail-page">
    <AppHeader />
    <div class="container main" v-loading="loading">
      <el-card v-if="notice" class="detail-card">
        <div class="breadcrumb">
          <el-button text @click="$router.push('/rural-info')">
            <el-icon><ArrowLeft /></el-icon> 返回信息中心
          </el-button>
        </div>
        <div class="detail-header">
          <div class="header-tags">
            <el-tag v-if="notice.is_top" type="danger" effect="dark" size="small">置顶</el-tag>
            <el-tag :type="categoryTag" size="small">{{ categoryMap[notice.category]||notice.category }}</el-tag>
          </div>
          <h1>{{ notice.title }}</h1>
          <div class="meta-bar">
            <div class="meta-item">
              <span v-if="notice.publish_dept" class="dept">{{ notice.publish_dept }}</span>
              <span class="date">发布日期：{{ formatDate(notice.publish_date||notice.created_at) }}</span>
              <span class="views"><el-icon><View /></el-icon> {{ notice.views }}</span>
            </div>
            <div class="author" v-if="notice.user"><span>发布者：{{ notice.user.username }}</span></div>
          </div>
        </div>
        <el-divider />
        <div class="detail-content" v-html="formatContent(notice.content)"></div>
        <div v-if="notice.attachment" class="att-bar"><el-link type="primary" :href="notice.attachment" target="_blank"><el-icon><Download /></el-icon>下载附件</el-link></div>

        <!-- 评论区 -->
        <el-divider />
        <div class="comments-section">
          <h4>评论 ({{ totalComments }})</h4>
          <div class="comment-input" v-if="token">
            <el-input v-model="commentContent" type="textarea" :rows="2" placeholder="写下你的评论..." />
            <el-button type="primary" size="small" @click="submitComment" :disabled="!commentContent.trim()" style="margin-top:8px;">发表评论</el-button>
          </div>
          <div v-else class="login-tip"><el-link type="primary" @click="$router.push('/login')">登录</el-link>后参与评论</div>
          <div v-if="comments.length>0" class="comment-list">
            <div v-for="c in comments" :key="c.id" class="comment-item" :class="'level-'+c.level">
              <el-avatar :size="c.level===1?36:28" class="c-avatar">{{ c.user?.username?.[0] }}</el-avatar>
              <div class="c-body">
                <div class="c-header"><span class="c-name">{{ c.user?.username||'匿名' }}</span><span class="c-time">{{ formatDate(c.created_at) }}</span></div>
                <div class="c-content">{{ c.content }}</div>
                <div class="c-actions">
                  <el-button size="small" text @click="toggleReply(c)">{{ replyingTo===c.id?'取消回复':'回复' }}</el-button>
                  <el-button size="small" text @click="likeComment(c)"><el-icon><Star /></el-icon>{{ c.likes||0 }}</el-button>
                </div>
                <div v-if="replyingTo===c.id" class="reply-box">
                  <el-input v-model="replyContent" :placeholder="'回复 '+c.user?.username" size="small" />
                  <el-button size="small" type="primary" @click="submitReply(c)" style="margin-top:4px;">回复</el-button>
                </div>
                <div v-if="c.children?.length" class="child-comments">
                  <div v-for="sub in c.children" :key="sub.id" class="comment-item child">
                    <el-avatar :size="24">{{ sub.user?.username?.[0] }}</el-avatar>
                    <div class="c-body">
                      <div class="c-header"><span class="c-name">{{ sub.user?.username }}</span><span class="c-time">{{ formatDate(sub.created_at) }}</span></div>
                      <div class="c-content">{{ sub.content }}</div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-card>
      <div v-else-if="!loading" class="not-found"><p>公告不存在</p><el-button type="primary" @click="$router.push('/rural-info')">返回信息中心</el-button></div>
    </div>
    <AppFooter />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft,View,Download,Star } from '@element-plus/icons-vue'
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
const notice = ref(null); const loading = ref(true)

const categoryMap={ subsidy:'补贴政策',land:'土地政策',environmental:'环保政策',technology:'科技政策',comprehensive:'综合政策',healthcare:'医疗养老',education:'教育政策',other:'其他政策' }
const tagTypes={ subsidy:'warning',environmental:'success',technology:'primary' }
const categoryTag = computed(()=>tagTypes[notice.value?.category]||'')

// comments
const comments = ref([]); const commentContent = ref('')
const replyingTo = ref(null); const replyContent = ref('')
const totalComments = computed(()=>{ let n=0; comments.value.forEach(c=>{ n++; if(c.children)n+=c.children.length }); return n })

const formatContent=(c)=>{if(!c)return'';let h=marked.parse(c);h=h.replace(/<pre><code class="language-(\w+)">([\s\S]*?)<\/code><\/pre>/g,(m,l,c2)=>{const d=c2.replace(/&lt;/g,'<').replace(/&gt;/g,'>').replace(/&amp;/g,'&');let h2;try{h2=(l&&hljs.getLanguage(l))?hljs.highlight(d,{language:l,ignoreIllegals:true}).value:hljs.highlightAuto(d).value}catch{h2=d};return`<pre class="code-block"><code class="language-${l} hljs">${h2}</code></pre>`});return h}
const formatDate=(t)=>t?new Date(t).toLocaleDateString('zh-CN'):''

const fetchComments=async()=>{try{const r=await api.getCommentRuralInfos({target_type:'policy_notice',target_id:route.params.id});comments.value=r.list||[]}catch{}}
const submitComment=async()=>{if(!commentContent.value.trim())return;try{await api.createCommentRuralInfo({target_type:'policy_notice',target_id:Number(route.params.id),content:commentContent.value,parent_id:0});ElMessage.success('评论成功');commentContent.value='';await fetchComments()}catch{ElMessage.error('评论失败')}}
const toggleReply=(c)=>{replyingTo.value=replyingTo.value===c.id?null:c.id;replyContent.value=''}
const submitReply=async(p)=>{if(!replyContent.value.trim())return;try{await api.createCommentRuralInfo({target_type:'policy_notice',target_id:Number(route.params.id),content:replyContent.value,parent_id:p.id,reply_to_user_id:p.user_id});ElMessage.success('回复成功');replyingTo.value=null;replyContent.value='';await fetchComments()}catch{ElMessage.error('回复失败')}}
const likeComment=async(c)=>{try{const r=await api.likeCommentRuralInfo(c.id);c.likes=r.likes}catch{ElMessage.error('操作失败')}}

onMounted(async()=>{loading.value=true;try{notice.value=await api.getPolicyNotice(route.params.id)}catch{notice.value=null}loading.value=false;fetchComments()})
</script>

<style scoped>
.policy-detail-page{min-height:100vh;background:#f5f7fa}.container{max-width:860px;margin:0 auto;padding:0 20px}.main{padding:30px 20px}
.detail-card{border-radius:12px}.breadcrumb{margin-bottom:15px}
.detail-header{margin-bottom:10px}.header-tags{display:flex;gap:8px;margin-bottom:12px}
.detail-header h1{font-size:26px;margin:0 0 16px;color:#333;line-height:1.4}
.meta-bar{display:flex;flex-direction:column;gap:6px;font-size:13px;color:#999}
.meta-item{display:flex;gap:16px;flex-wrap:wrap}.meta-item span{display:flex;align-items:center;gap:3px}
.dept{color:#e6a23c;font-weight:500}.author{color:#bbb}
.detail-content{line-height:1.9;color:#333;font-size:15px;margin-top:16px}
.detail-content :deep(p){margin-bottom:14px;text-indent:2em}.detail-content :deep(img){max-width:100%;border-radius:8px;margin:10px 0}
.detail-content :deep(ul),.detail-content :deep(ol){padding-left:24px;margin-bottom:14px}.detail-content :deep(li){margin-bottom:6px}
.detail-content :deep(blockquote){border-left:4px solid #e6a23c;background:#fef9e7;padding:12px 16px;margin:14px 0;border-radius:0 6px 6px 0;color:#555}
.detail-content :deep(.code-block){background:#282c34;color:#abb2bf;padding:15px;border-radius:8px;overflow-x:auto;margin:10px 0}
.detail-content :deep(h2){font-size:20px;margin:24px 0 12px}.detail-content :deep(h3){font-size:17px;margin:20px 0 10px}
.detail-content :deep(table){border-collapse:collapse;width:100%;margin:14px 0}
.detail-content :deep(th),.detail-content :deep(td){border:1px solid #e0e0e0;padding:8px 12px;font-size:13px}.detail-content :deep(th){background:#f5f7fa;font-weight:600}
.att-bar{margin-top:20px;padding:12px;background:#f0f9eb;border-radius:8px}.not-found{text-align:center;padding:60px;color:#999}

.comments-section{margin-top:10px}.comments-section h4{margin-bottom:14px;color:#333}
.login-tip{margin:10px 0;color:#999;font-size:13px}.comment-input{margin-bottom:16px}
.comment-list{margin-top:16px}
.comment-item{display:flex;gap:10px;margin-bottom:14px}
.comment-item.child{margin-top:10px;margin-left:38px;padding:8px 10px;background:#fafafa;border-radius:8px}
.c-avatar{flex-shrink:0}.c-body{flex:1;min-width:0}
.c-header{display:flex;gap:8px;margin-bottom:4px}.c-name{font-size:13px;color:#409eff;font-weight:500}.c-time{font-size:12px;color:#bbb}
.c-content{font-size:14px;color:#444;line-height:1.6}.c-actions{margin-top:4px}
.reply-box{margin-top:8px}.child-comments{margin-top:8px}
</style>
