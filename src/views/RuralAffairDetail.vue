<template>
  <div class="affair-detail-page">
    <AppHeader />

    <div class="container main" v-loading="loading">
      <div v-if="fullDetail?.affair" class="content">
        <!-- ============ 状态头部 ============ -->
        <el-card class="status-header-card">
          <div class="status-header">
            <div class="status-left">
              <h1 class="title">{{ fullDetail.affair.title }}</h1>
              <div class="sub-meta">
                <el-tag :type="getStatusType(fullDetail.affair.status)" size="large" effect="dark">
                  {{ getStatusText(fullDetail.affair.status) }}
                </el-tag>
                <el-tag type="info" size="default">{{ getTypeText(fullDetail.affair.type) }}</el-tag>
                <span class="addr"><el-icon><Location /></el-icon>{{ fullDetail.affair.address }}</span>
              </div>
            </div>
            <div class="status-right">
              <span class="time-label">提交时间：{{ formatTime(fullDetail.affair.created_at) }}</span>
              <span v-if="fullDetail.affair.last_modified_at" class="time-label mod">
                最后修改：{{ formatTime(fullDetail.affair.last_modified_at) }}
              </span>
            </div>
          </div>
        </el-card>

        <!-- ============ 完整时间线 ============ -->
        <el-card class="timeline-card">
          <template #header><span>事务处理时间线</span></template>
          <div class="timeline-wrapper">
            <div class="timeline">
              <!-- ① 用户提交 -->
              <div class="tl-node done" :class="{ active: currentStep >= 1 }">
                <div class="tl-dot"><el-icon><Edit /></el-icon></div>
                <div class="tl-line"></div>
                <div class="tl-body">
                  <div class="tl-title">用户提交</div>
                  <div class="tl-time">{{ formatTime(fullDetail.affair.created_at) }}</div>
                  <div class="tl-desc">提交事务申请</div>
                </div>
              </div>

              <!-- ①B 用户修改（条件显示） -->
              <div v-if="fullDetail.modifications?.length" class="tl-node info" :class="{ active: currentStep >= 1 }">
                <div class="tl-dot"><el-icon><Refresh /></el-icon></div>
                <div class="tl-line"></div>
                <div class="tl-body">
                  <div class="tl-title">用户修改 <el-tag size="small" type="info">{{ fullDetail.modifications.length }}次</el-tag></div>
                  <div class="tl-time">最后修改：{{ formatTime(fullDetail.affair.last_modified_at) }}</div>
                </div>
              </div>

              <!-- ② 审核 -->
              <div class="tl-node" :class="getAuditNodeClass()">
                <div class="tl-dot"><el-icon><Checked /></el-icon></div>
                <div class="tl-line"></div>
                <div class="tl-body">
                  <div class="tl-title">{{ fullDetail.affair.status === 3 ? '审核未通过' : '管理员审核' }}</div>
                  <div class="tl-time">{{ formatTime(fullDetail.affair.audit_time) }}</div>
                  <div class="tl-desc" v-if="fullDetail.affair.status === 3">
                    <el-text type="danger">拒绝原因：{{ fullDetail.affair.reject_reason }}</el-text>
                  </div>
                  <div class="tl-desc" v-else-if="fullDetail.affair.audit_name">
                    审核人：{{ fullDetail.affair.audit_name }}
                  </div>
                </div>
              </div>

              <!-- ③ 人员接取 -->
              <div class="tl-node" :class="{ done: currentStep >= 3, active: currentStep === 3 }">
                <div class="tl-dot"><el-icon><UserFilled /></el-icon></div>
                <div class="tl-line"></div>
                <div class="tl-body">
                  <div class="tl-title">事务处理人员接取</div>
                  <div class="tl-time">{{ formatTime(fullDetail.affair.process_time) }}</div>
                  <div class="tl-desc" v-if="fullDetail.affair.handler_name">
                    {{ fullDetail.affair.handler_name }}
                  </div>
                </div>
              </div>

              <!-- ④ 首次处理 -->
              <div class="tl-node" :class="{ done: currentStep >= 4, active: currentStep === 4 }">
                <div class="tl-dot"><el-icon><Setting /></el-icon></div>
                <div class="tl-line"></div>
                <div class="tl-body">
                  <div class="tl-title">处理人员提交处理记录</div>
                  <div class="tl-time">{{ formatTime(fullDetail.affair.first_response_at) }}</div>
                  <div class="tl-desc" v-if="fullDetail.affair.process_content">
                    {{ truncateText(fullDetail.affair.process_content, 60) }}
                  </div>
                </div>
              </div>

              <!-- ⑤ 追问/追答循环 -->
              <div v-for="(fu, fi) in fullDetail.follow_ups" :key="'fu'+fi"
                   class="tl-node" :class="{ info: fu.type === 'answer', warning: fu.type === 'question', done: true }">
                <div class="tl-dot"><el-icon><ChatLineSquare /></el-icon></div>
                <div class="tl-line"></div>
                <div class="tl-body">
                  <div class="tl-title">
                    <el-tag size="small" :type="fu.type === 'question' ? 'warning' : 'info'">
                      {{ fu.type === 'question' ? '追问' : '追答' }}
                    </el-tag>
                    {{ fu.user_name }}
                  </div>
                  <div class="tl-time">{{ formatTime(fu.created_at) }}</div>
                  <div class="tl-desc">{{ truncateText(fu.content, 60) }}</div>
                </div>
              </div>

              <!-- ⑥ 已完成 -->
              <div class="tl-node" :class="{ done: currentStep >= 5, success: currentStep === 5 }">
                <div class="tl-dot"><el-icon><CircleCheck /></el-icon></div>
                <div class="tl-line"></div>
                <div class="tl-body">
                  <div class="tl-title">已完成</div>
                  <div class="tl-time">{{ formatTime(fullDetail.affair.completed_time) }}</div>
                </div>
              </div>

              <!-- ⑦ 申诉 -->
              <div v-if="fullDetail.appeal" class="tl-node" :class="getAppealNodeClass()">
                <div class="tl-dot"><el-icon><Warning /></el-icon></div>
                <div class="tl-line"></div>
                <div class="tl-body">
                  <div class="tl-title">
                    <el-tag size="small" :type="fullDetail.appeal.applicant_type === 'user' ? 'danger' : 'warning'">
                      {{ fullDetail.appeal.applicant_type === 'user' ? '用户申诉' : '处理人员申诉' }}
                    </el-tag>
                    申请人：{{ fullDetail.appeal.applicant_name }}
                  </div>
                  <div class="tl-time">提交：{{ formatTime(fullDetail.appeal.created_at) }}</div>
                  <div class="tl-desc">理由：{{ truncateText(fullDetail.appeal.reason, 60) }}</div>
                  <div class="tl-desc" v-if="fullDetail.appeal.status === 'resolved'">
                    <el-text type="success">结果：{{ fullDetail.appeal.result }}</el-text>
                    · 处理人：{{ fullDetail.appeal.handler_name }}
                    · {{ formatTime(fullDetail.appeal.processed_at) }}
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 审核不通过警告 -->
          <div v-if="fullDetail.affair.status === 3" class="reject-alert">
            <el-alert type="error" :title="`审核未通过：${fullDetail.affair.reject_reason}`" :closable="false" show-icon />
          </div>
        </el-card>

        <!-- ============ 事务描述 ============ -->
        <el-card class="detail-card">
          <template #header><span>事务详情</span></template>
          <div class="detail-content">
            <div class="author-line">
              <el-avatar :src="fullDetail.affair.user?.avatar_url" :size="36">{{ fullDetail.affair.user?.username?.[0] }}</el-avatar>
              <span class="author-name">{{ fullDetail.affair.user?.nickname || fullDetail.affair.user?.username }}</span>
              <span class="author-phone" v-if="fullDetail.affair.user?.phone">{{ fullDetail.affair.user?.phone }}</span>
            </div>
            <div class="description" v-html="(fullDetail.affair.content || '').replace(/\n/g, '<br>')" />
            <!-- 图片 -->
            <div v-if="fullDetail.affair.images" class="media-section">
              <div class="images-grid">
                <el-image v-for="(img, i) in parseImages(fullDetail.affair.images)" :key="i"
                  :src="img" fit="cover" :preview-src-list="parseImages(fullDetail.affair.images)" class="detail-img" />
              </div>
            </div>
            <!-- 视频 -->
            <div v-if="fullDetail.affair.video_url" class="media-section">
              <video :src="fullDetail.affair.video_url" controls class="detail-video" />
            </div>
          </div>
        </el-card>

        <!-- ============ 修改记录详情 ============ -->
        <el-card v-if="fullDetail.modifications?.length" class="mod-card">
          <template #header>
            <span>修改记录 <el-tag size="small" type="info">{{ fullDetail.modifications.length }}次</el-tag></span>
          </template>
          <el-timeline>
            <el-timeline-item v-for="mod in fullDetail.modifications" :key="mod.id"
              :timestamp="formatTime(mod.created_at)" placement="top" type="primary">
              <div class="mod-detail">
                <strong>{{ mod.user_name }}</strong> 修改了事务信息
                <div v-if="mod.new_title !== mod.old_title" class="diff-line">
                  <span class="diff-label">标题：</span>
                  <del>{{ mod.old_title }}</del> → <b>{{ mod.new_title }}</b>
                </div>
                <div v-if="mod.new_content !== mod.old_content" class="diff-line">
                  <span class="diff-label">内容有变更</span>
                </div>
                <div v-if="mod.new_address !== mod.old_address" class="diff-line">
                  <span class="diff-label">地点：</span>
                  <del>{{ mod.old_address }}</del> → <b>{{ mod.new_address }}</b>
                </div>
              </div>
            </el-timeline-item>
          </el-timeline>
        </el-card>

        <!-- ============ 处理结果 ============ -->
        <el-card v-if="fullDetail.affair.process_content" class="process-card">
          <template #header>
            <span>处理记录</span>
          </template>
          <div class="process-info-line">
            <span><strong>处理人员：</strong>{{ fullDetail.affair.handler_name }}</span>
            <span v-if="fullDetail.affair.first_response_at"><strong>首次回应：</strong>{{ formatTime(fullDetail.affair.first_response_at) }}</span>
            <span v-if="fullDetail.affair.completed_time"><strong>完成时间：</strong>{{ formatTime(fullDetail.affair.completed_time) }}</span>
          </div>
          <div class="process-content" v-html="(fullDetail.affair.process_content || '').replace(/\n/g, '<br>')" />
          <div v-if="fullDetail.affair.process_images" class="media-section">
            <div class="images-grid">
              <el-image v-for="(img, i) in parseImages(fullDetail.affair.process_images)" :key="'pi'+i"
                :src="img" fit="cover" :preview-src-list="parseImages(fullDetail.affair.process_images)" class="detail-img" />
            </div>
          </div>
        </el-card>

        <!-- ============ 追问追答详情 ============ -->
        <el-card v-if="fullDetail.follow_ups?.length" class="follow-card">
          <template #header><span>追问/追答对话</span></template>
          <div class="follow-list">
            <div v-for="fu in fullDetail.follow_ups" :key="fu.id"
                 class="follow-item" :class="{ question: fu.type === 'question', answer: fu.type === 'answer' }">
              <div class="follow-header">
                <el-tag :type="fu.type === 'question' ? 'warning' : 'info'" size="small">
                  {{ fu.type === 'question' ? '追问' : '追答' }}
                </el-tag>
                <span class="follow-user">{{ fu.user_name }}</span>
                <span class="follow-time">{{ formatTime(fu.created_at) }}</span>
              </div>
              <div class="follow-body" v-html="(fu.content || '').replace(/\n/g, '<br>')" />
              <div v-if="fu.images" class="media-section" style="margin-top:8px;">
                <div class="images-grid small">
                  <el-image v-for="(img, i) in parseImages(fu.images)" :key="'fui'+i"
                    :src="img" fit="cover" :preview-src-list="parseImages(fu.images)" class="detail-img-sm" />
                </div>
              </div>
            </div>
          </div>
        </el-card>

        <!-- ============ 申诉详情 ============ -->
        <el-card v-if="fullDetail.appeal" class="appeal-card">
          <template #header>
            <span>申诉信息
              <el-tag :type="getAppealStatusType(fullDetail.appeal.status)" size="small">
                {{ getAppealStatusText(fullDetail.appeal.status) }}
              </el-tag>
            </span>
          </template>
          <div class="appeal-info-line">
            <span><strong>申诉人：</strong>{{ fullDetail.appeal.applicant_name }}
              <el-tag size="small" :type="fullDetail.appeal.applicant_type === 'user' ? '' : 'warning'">
                {{ fullDetail.appeal.applicant_type === 'user' ? '用户' : '处理人员' }}
              </el-tag>
            </span>
            <span><strong>提交时间：</strong>{{ formatTime(fullDetail.appeal.created_at) }}</span>
          </div>
          <div class="appeal-reason">
            <strong>申诉理由：</strong>
            <div v-html="(fullDetail.appeal.reason || '').replace(/\n/g, '<br>')" />
          </div>
          <div v-if="fullDetail.appeal.images" class="media-section">
            <div class="images-grid small">
              <el-image v-for="(img, i) in parseImages(fullDetail.appeal.images)" :key="'ai'+i"
                :src="img" fit="cover" :preview-src-list="parseImages(fullDetail.appeal.images)" class="detail-img-sm" />
            </div>
          </div>
          <!-- 处理结果 -->
          <div v-if="fullDetail.appeal.status === 'resolved'" class="appeal-result">
            <el-alert type="success" :closable="false" show-icon>
              <template #title>
                <div><strong>处理人：</strong>{{ fullDetail.appeal.handler_name }}</div>
                <div><strong>处理时间：</strong>{{ formatTime(fullDetail.appeal.processed_at) }}</div>
                <div><strong>处理结果：</strong>{{ fullDetail.appeal.result }}</div>
              </template>
            </el-alert>
          </div>
          <!-- 辅助材料 -->
          <div v-if="fullDetail.appeal_materials?.length" class="appeal-materials">
            <h4>辅助材料（共{{ fullDetail.appeal_materials.length }}份）</h4>
            <div v-for="mat in fullDetail.appeal_materials" :key="mat.id" class="material-item">
              <div class="mat-header">
                <span class="mat-submitter">{{ mat.submitter_name }}
                  <el-tag size="small" :type="mat.submitter_type === 'user' ? '' : 'warning'">
                    {{ mat.submitter_type === 'user' ? '用户' : '处理人员' }}
                  </el-tag>
                </span>
                <span class="mat-time">{{ formatTime(mat.created_at) }}</span>
              </div>
              <div v-if="mat.content" v-html="(mat.content || '').replace(/\n/g, '<br>')" class="mat-content" />
              <div v-if="mat.images" class="media-section" style="margin-top:8px;">
                <div class="images-grid small">
                  <el-image v-for="(img, i) in parseImages(mat.images)" :key="'mi'+i"
                    :src="img" fit="cover" :preview-src-list="parseImages(mat.images)" class="detail-img-sm" />
                </div>
              </div>
            </div>
          </div>
        </el-card>

        <!-- ============ 操作区 ============ -->
        <el-card class="action-card">
          <template #header>
            <div style="display: flex; align-items: center; justify-content: space-between;">
              <span>操作</span>
              <el-tag v-if="isAdmin" type="danger" size="small">管理员</el-tag>
              <el-tag v-else-if="fullDetail.can_respond" type="warning" size="small">处理人员</el-tag>
              <el-tag v-else-if="fullDetail.affair?.user_id === currentUser?.id" type="primary" size="small">事务提交人</el-tag>
            </div>
          </template>
          <div class="action-buttons" v-if="token">
            <!-- 提交者操作 -->
            <template v-if="fullDetail.affair?.user_id === currentUser?.id">
              <el-divider content-position="left">事务操作</el-divider>
            </template>
            <template v-if="fullDetail.can_modify">
              <el-button type="primary" @click="showModifyDialog = true"><el-icon><Edit /></el-icon>修改事务</el-button>
            </template>
            <template v-if="fullDetail.can_follow_up">
              <el-button type="warning" @click="showFollowUpDialog = true"><el-icon><ChatLineSquare /></el-icon>追问处理进度</el-button>
            </template>
            <template v-if="fullDetail.can_confirm">
              <el-button type="success" @click="handleConfirmComplete"><el-icon><CircleCheck /></el-icon>确认事务完成</el-button>
            </template>

            <!-- 处理人员操作 -->
            <template v-if="isProcessor">
              <el-divider content-position="left">处理操作</el-divider>
              <!-- 接取事务：审核通过、尚未有处理人 -->
              <el-button v-if="fullDetail.affair.status === 2 && !fullDetail.affair.handler_id" type="success" @click="handleStartProcess">
                <el-icon><UserFilled /></el-icon>接取事务
              </el-button>
              <!-- 填写处理记录：已接取且状态为处理中 -->
              <el-button v-if="fullDetail.affair.status === 4 && isAffairHandler" type="warning" @click="showProcessDialog = true">
                <el-icon><Setting /></el-icon>填写处理记录
              </el-button>
              <!-- 回复追问：追问中 -->
              <template v-if="fullDetail.can_respond">
                <el-button type="primary" @click="showAnswerDialog = true"><el-icon><ChatDotRound /></el-icon>回复追问</el-button>
              </template>
            </template>

            <!-- 申诉操作 -->
            <template v-if="fullDetail.can_appeal">
              <el-divider content-position="left">申诉</el-divider>
              <el-button type="danger" @click="showAppealDialog = true"><el-icon><Warning /></el-icon>发起申诉</el-button>
            </template>

            <!-- 管理员操作 -->
            <template v-if="isAdmin">
              <el-divider content-position="left">管理员操作</el-divider>
              <template v-if="fullDetail.appeal && fullDetail.appeal.status !== 'resolved'">
                <el-button type="success" @click="showAdminHandleAppealDialog = true"><el-icon><Checked /></el-icon>处理申诉</el-button>
              </template>
            </template>

            <div style="margin-top: 12px; width: 100%;">
              <el-button @click="$router.push('/affairs')">返回事务大厅</el-button>
              <el-button v-if="fullDetail.affair?.user_id === currentUser?.id" @click="$router.push('/user')">返回个人中心</el-button>
            </div>
          </div>
          <div v-else class="login-tip">
            请<el-link type="primary" @click="$router.push('/login')">登录</el-link>后进行操作
          </div>
        </el-card>
      </div>

      <div v-else class="empty-state">
        <el-icon :size="80" color="#dcdfe6"><Document /></el-icon>
        <p>事务不存在或已被删除</p>
        <el-button type="primary" @click="$router.push('/affairs')">返回列表</el-button>
      </div>
    </div>

    <!-- ========== 修改弹窗 ========== -->
    <el-dialog v-model="showModifyDialog" title="修改事务" width="600px">
      <el-form :model="modifyForm" label-width="80px">
        <el-form-item label="标题"><el-input v-model="modifyForm.title" /></el-form-item>
        <el-form-item label="类型">
          <el-select v-model="modifyForm.type">
            <el-option v-for="(v,k) in typeMap" :key="k" :label="v" :value="k" />
          </el-select>
        </el-form-item>
        <el-form-item label="地点"><el-input v-model="modifyForm.address" /></el-form-item>
        <el-form-item label="问题描述"><el-input v-model="modifyForm.content" type="textarea" :rows="4" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showModifyDialog = false">取消</el-button>
        <el-button type="primary" @click="handleModify" :loading="submitting">确认修改</el-button>
      </template>
    </el-dialog>

    <!-- ========== 追问弹窗 ========== -->
    <el-dialog v-model="showFollowUpDialog" title="追问" width="500px">
      <el-input v-model="followUpContent" type="textarea" :rows="3" placeholder="请输入追问内容" />
      <template #footer>
        <el-button @click="showFollowUpDialog = false">取消</el-button>
        <el-button type="primary" @click="handleFollowUp" :loading="submitting">提交追问</el-button>
      </template>
    </el-dialog>

    <!-- ========== 追答弹窗 ========== -->
    <el-dialog v-model="showAnswerDialog" title="追答" width="500px">
      <el-input v-model="followUpContent" type="textarea" :rows="3" placeholder="请输入追答内容" />
      <template #footer>
        <el-button @click="showAnswerDialog = false">取消</el-button>
        <el-button type="primary" @click="handleAnswer" :loading="submitting">提交追答</el-button>
      </template>
    </el-dialog>

    <!-- ========== 申诉弹窗 ========== -->
    <el-dialog v-model="showAppealDialog" title="发起申诉" width="500px">
      <el-form label-width="80px">
        <el-form-item label="申诉类型">
          <el-radio-group v-model="appealType">
            <el-radio label="user" v-if="fullDetail.affair.status === 5">用户申诉（已完成事务）</el-radio>
            <el-radio label="handler" v-if="fullDetail.affair.status === 4 || fullDetail.affair.status === 6">处理人员申诉（处理中事务）</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="申诉理由">
          <el-input v-model="appealReason" type="textarea" :rows="3" placeholder="请详细说明申诉理由" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAppealDialog = false">取消</el-button>
        <el-button type="danger" @click="handleAppeal" :loading="submitting">提交申诉</el-button>
      </template>
    </el-dialog>

    <!-- ========== 管理员处理申诉弹窗 ========== -->
    <el-dialog v-model="showAdminHandleAppealDialog" title="处理申诉" width="500px">
      <el-form label-width="80px">
        <el-form-item label="处理结果">
          <el-input v-model="appealResult" type="textarea" :rows="3" placeholder="请输入处理结果说明" />
        </el-form-item>
        <el-form-item label="是否批准">
          <el-switch v-model="appealApprove" active-text="批准" inactive-text="驳回" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAdminHandleAppealDialog = false">取消</el-button>
        <el-button type="primary" @click="handleAdminProcessAppeal" :loading="submitting">确认处理</el-button>
      </template>
    </el-dialog>

    <!-- ========== 填写处理记录弹窗 ========== -->
    <el-dialog v-model="showProcessDialog" title="填写处理记录" width="600px">
      <el-form :model="processForm" label-width="100px">
        <el-form-item label="处理详情" required>
          <el-input v-model="processForm.content" type="textarea" :rows="6" placeholder="请详细描述处理过程和结果" />
        </el-form-item>
        <el-form-item label="处理图片">
          <el-upload
            action="/api/upload"
            :headers="{ Authorization: tokenValue }"
            list-type="picture-card"
            :on-success="handleProcessImageSuccess"
            :before-upload="beforeProcessUpload"
            :on-remove="handleProcessImageRemove"
            :file-list="processImageList"
          >
            <el-icon><Plus /></el-icon>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showProcessDialog = false">取消</el-button>
        <el-button type="primary" @click="handleProcessRecord" :loading="submitting">提交处理记录</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Document, Location, Edit, Refresh, Checked, UserFilled, Setting, ChatLineSquare, CircleCheck, Warning, ChatDotRound, Plus } from '@element-plus/icons-vue'
import api from '@/utils/api'
import { getToken, getUserInfo } from '@/utils/auth'

const route = useRoute()
const router = useRouter()
const token = getToken()
const currentUser = getUserInfo()

const fullDetail = ref(null)
const loading = ref(false)
const submitting = ref(false)

// 弹窗状态
const showModifyDialog = ref(false)
const showFollowUpDialog = ref(false)
const showAnswerDialog = ref(false)
const showAppealDialog = ref(false)
const showAdminHandleAppealDialog = ref(false)
const showProcessDialog = ref(false)

// 表单数据
const modifyForm = ref({})
const followUpContent = ref('')
const appealType = ref('')
const appealReason = ref('')
const appealResult = ref('')
const appealApprove = ref(false)
const processForm = reactive({ content: '', images: '' })
const processImageList = ref([])

const typeMap = { infrastructure: '基础设施', hardware: '硬件问题', env: '环境问题', safety: '安全隐患', other: '其他' }

const isAdmin = computed(() => {
  const role = currentUser?.role || ''
  return role === 'admin' || role === 'sysadmin'
})

const isProcessor = computed(() => {
  const role = currentUser?.role || ''
  return role === 'processor' || isAdmin.value
})

const isAffairHandler = computed(() => {
  return currentUser?.id && fullDetail.value?.affair?.handler_id === currentUser.id
})

const tokenValue = computed(() => getToken() || '')

const currentStep = computed(() => {
  if (!fullDetail.value?.affair) return 0
  const s = fullDetail.value.affair.status
  if (s === 1) return 1  // 待审核
  if (s === 2) return 2  // 审核通过
  if (s === 3) return 2  // 审核不通过 — 停在审核节点
  if (s === 4 || s === 6) return s === 6 ? 4 : 3 // 处理中/追问中显示处理中
  if (s === 5) return 5  // 已完成
  if (s === 7) return 5  // 申诉中显示在已完成后
  return 0
})

const getStatusText = (s) => {
  const map = { 1: '待审核', 2: '审核通过', 3: '审核不通过', 4: '处理中', 5: '已完成', 6: '追问中', 7: '申诉中' }
  return map[s] || '未知'
}
const getStatusType = (s) => {
  const map = { 1: 'warning', 2: 'primary', 3: 'danger', 4: 'info', 5: 'success', 6: 'warning', 7: 'danger' }
  return map[s] || 'info'
}
const getTypeText = (t) => typeMap[t] || '其他'
const getAuditNodeClass = () => {
  const s = fullDetail.value?.affair?.status
  if (s === 3) return 'error'
  if (s >= 2) return 'done'
  return ''
}
const getAppealNodeClass = () => {
  const a = fullDetail.value?.appeal
  if (!a) return ''
  if (a.status === 'resolved') return 'done success'
  return 'active error'
}
const getAppealStatusText = (s) => {
  const map = { pending: '待处理', processing: '处理中', resolved: '已处理' }
  return map[s] || s
}
const getAppealStatusType = (s) => {
  const map = { pending: 'warning', processing: 'info', resolved: 'success' }
  return map[s] || 'info'
}

const formatTime = (t) => {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}
const truncateText = (t, len) => {
  if (!t) return ''
  return t.length > len ? t.slice(0, len) + '...' : t
}
const parseImages = (s) => {
  if (!s) return []
  try { return JSON.parse(s) } catch { return [] }
}

// 加载数据
const fetchData = async () => {
  loading.value = true
  try {
    const id = route.params.id
    if (token) {
      // 登录用户获取完整详情
      fullDetail.value = await api.getAffairFullDetail(id)
    } else {
      // 未登录获取基础数据
      const affair = await api.getRuralAffair(id)
      fullDetail.value = { affair, modifications: [], follow_ups: [], appeal: null, appeal_materials: [], can_modify: false, can_follow_up: false, can_respond: false, can_confirm: false, can_appeal: false }
    }
  } catch (e) {
    console.error('获取事务详情失败:', e)
  } finally {
    loading.value = false
  }
}

// 修改
const handleModify = async () => {
  if (!modifyForm.value.title?.trim()) { ElMessage.warning('标题不能为空'); return }
  submitting.value = true
  try {
    await api.modifyRuralAffair(route.params.id, modifyForm.value)
    ElMessage.success('修改成功')
    showModifyDialog.value = false
    await fetchData()
  } catch (e) {
    ElMessage.error('修改失败')
  } finally { submitting.value = false }
}

// 追问
const handleFollowUp = async () => {
  if (!followUpContent.value.trim()) { ElMessage.warning('追问内容不能为空'); return }
  submitting.value = true
  try {
    await api.addFollowUpQuestion(route.params.id, { content: followUpContent.value, images: '' })
    ElMessage.success('追问成功')
    showFollowUpDialog.value = false
    followUpContent.value = ''
    await fetchData()
  } catch (e) {
    ElMessage.error('追问失败')
  } finally { submitting.value = false }
}

// 追答
const handleAnswer = async () => {
  if (!followUpContent.value.trim()) { ElMessage.warning('追答内容不能为空'); return }
  submitting.value = true
  try {
    await api.addFollowUpAnswer(route.params.id, { content: followUpContent.value, images: '' })
    ElMessage.success('追答成功')
    showAnswerDialog.value = false
    followUpContent.value = ''
    await fetchData()
  } catch (e) {
    ElMessage.error('追答失败')
  } finally { submitting.value = false }
}

// 确认完成
const handleConfirmComplete = async () => {
  try {
    await ElMessageBox.confirm('确认该事务已处理完成？', '提示', { type: 'info' })
    await api.confirmCompleteAffair(route.params.id)
    ElMessage.success('已确认完成')
    await fetchData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('操作失败')
  }
}

// 处理人员接取事务
const handleStartProcess = async () => {
  try {
    await ElMessageBox.confirm('确认接取该事务？接取后将由您负责处理', '提示', { type: 'info' })
    await api.startProcessRuralAffair(route.params.id)
    ElMessage.success('已接取该事务')
    await fetchData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('操作失败')
  }
}

// 处理图片上传
const beforeProcessUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) ElMessage.error('只能上传图片')
  if (!isLt5M) ElMessage.error('图片大小不能超过5MB')
  return isImage && isLt5M
}

const handleProcessImageSuccess = (res) => {
  const newUrl = res.url
  if (processForm.images) {
    try {
      const arr = JSON.parse(processForm.images)
      processForm.images = JSON.stringify([...arr, newUrl])
    } catch {
      processForm.images = JSON.stringify([newUrl])
    }
  } else {
    processForm.images = JSON.stringify([newUrl])
  }
}

const handleProcessImageRemove = () => {
  processForm.images = ''
}

// 提交处理记录
const handleProcessRecord = async () => {
  if (!processForm.content.trim()) {
    ElMessage.warning('请填写处理详情')
    return
  }
  submitting.value = true
  try {
    await api.processRuralAffair(route.params.id, {
      process_content: processForm.content,
      process_images: processForm.images
    })
    ElMessage.success('处理记录已提交，等待用户确认完成')
    showProcessDialog.value = false
    processForm.content = ''
    processForm.images = ''
    processImageList.value = []
    await fetchData()
  } catch (e) {
    ElMessage.error('提交失败')
  } finally { submitting.value = false }
}

// 申诉
const handleAppeal = async () => {
  if (!appealReason.value.trim()) { ElMessage.warning('申诉理由不能为空'); return }
  if (!appealType.value) { ElMessage.warning('请选择申诉类型'); return }
  submitting.value = true
  try {
    await api.createAppeal(route.params.id, { applicant_type: appealType.value, reason: appealReason.value, images: '' })
    ElMessage.success('申诉已提交')
    showAppealDialog.value = false
    appealReason.value = ''
    await fetchData()
  } catch (e) {
    ElMessage.error('申诉失败')
  } finally { submitting.value = false }
}

// 管理员处理申诉
const handleAdminProcessAppeal = async () => {
  if (!appealResult.value.trim()) { ElMessage.warning('处理结果不能为空'); return }
  submitting.value = true
  try {
    await api.adminProcessAppeal(fullDetail.value.appeal.id, { result: appealResult.value, approve: appealApprove.value })
    ElMessage.success('申诉已处理')
    showAdminHandleAppealDialog.value = false
    await fetchData()
  } catch (e) {
    ElMessage.error('处理失败')
  } finally { submitting.value = false }
}

onMounted(() => {
  fetchData().then(() => {
    // 如果URL中有action=modify参数，自动打开修改对话框
    if (route.query.action === 'modify' && fullDetail.value?.can_modify) {
      showModifyDialog.value = true
      const a = fullDetail.value.affair
      modifyForm.value = { title: a.title, type: a.type, address: a.address, content: a.content }
    }
  })
})

// 弹窗打开时同步数据
const watchModifyOpen = () => {
  if (showModifyDialog.value && fullDetail.value?.affair) {
    const a = fullDetail.value.affair
    modifyForm.value = { title: a.title, type: a.type, address: a.address, content: a.content }
  }
}
const watchAppealOpen = () => {
  if (showAppealDialog.value && fullDetail.value?.affair) {
    if (fullDetail.value.affair.status === 5) appealType.value = 'user'
    else appealType.value = 'handler'
  }
}
</script>

<style scoped>
.affair-detail-page { min-height: 100vh; background: #f0f2f5; }
.container { max-width: 860px; margin: 0 auto; padding: 0 20px; }
.main { padding: 24px 20px; }
.content { display: flex; flex-direction: column; gap: 16px; }

/* ===== 状态头部 ===== */
.status-header-card { background: #fff; }
.status-header { display: flex; justify-content: space-between; align-items: flex-start; flex-wrap: wrap; gap: 12px; }
.title { font-size: 22px; color: #222; margin: 0; }
.sub-meta { display: flex; align-items: center; gap: 10px; margin-top: 8px; flex-wrap: wrap; }
.addr { display: flex; align-items: center; gap: 4px; color: #666; font-size: 13px; }
.status-right { text-align: right; }
.time-label { display: block; font-size: 12px; color: #999; }
.time-label.mod { color: #409eff; margin-top: 4px; }

/* ===== 时间线 ===== */
.timeline-card { background: #fff; }
.timeline-wrapper { overflow-x: auto; }
.timeline { display: flex; align-items: flex-start; padding: 16px 0; min-width: 700px; }
.tl-node { position: relative; display: flex; flex-direction: column; align-items: center; text-align: center; min-width: 90px; flex: 1; opacity: 0.45; transition: opacity 0.3s; }
.tl-node.active, .tl-node.done, .tl-node.error, .tl-node.success, .tl-node.info, .tl-node.warning { opacity: 1; }
.tl-dot { width: 36px; height: 36px; border-radius: 50%; background: #e8e8e8; display: flex; align-items: center; justify-content: center; color: #999; font-size: 16px; position: relative; z-index: 2; }
.tl-node.done .tl-dot { background: #e6f7ff; color: #409eff; border: 2px solid #409eff; }
.tl-node.success .tl-dot { background: #f0fff4; color: #67c23a; border: 2px solid #67c23a; }
.tl-node.error .tl-dot { background: #fef0f0; color: #f56c6c; border: 2px solid #f56c6c; }
.tl-node.active:not(.done) .tl-dot { animation: pulse 1.5s infinite; box-shadow: 0 0 0 4px rgba(64,158,255,0.25); }
.tl-node.warning .tl-dot { background: #fdf6ec; color: #e6a23c; border: 2px solid #e6a23c; }
.tl-node.info .tl-dot { background: #f4f4f5; color: #909399; border: 2px solid #909399; }
.tl-line { position: absolute; top: 18px; left: calc(50% + 20px); width: calc(100% - 40px); height: 2px; background: #e8e8e8; z-index: 1; }
.tl-node:last-child .tl-line { display: none; }
.tl-node.done .tl-line { background: #409eff; }
.tl-node.success .tl-line { background: #67c23a; }
.tl-body { margin-top: 8px; }
.tl-title { font-weight: 600; font-size: 14px; color: #333; white-space: nowrap; }
.tl-time { font-size: 11px; color: #999; margin-top: 2px; }
.tl-desc { font-size: 12px; color: #666; margin-top: 2px; max-width: 130px; }
@keyframes pulse { 0%,100% { box-shadow: 0 0 0 0 rgba(64,158,255,0.4); } 50% { box-shadow: 0 0 0 8px rgba(64,158,255,0); } }

/* ===== 基础卡片 ===== */
.detail-card, .mod-card, .process-card, .follow-card, .appeal-card, .action-card { background: #fff; }
.detail-content, .process-content { color: #555; line-height: 1.8; font-size: 15px; }
.author-line { display: flex; align-items: center; gap: 10px; margin-bottom: 16px; padding-bottom: 12px; border-bottom: 1px solid #f0f0f0; }
.author-name { font-weight: 500; color: #333; }
.author-phone { font-size: 13px; color: #999; margin-left: auto; }
.description { margin-bottom: 8px; }
.media-section { margin-top: 16px; }
.images-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(160px, 1fr)); gap: 10px; }
.images-grid.small { grid-template-columns: repeat(auto-fill, minmax(120px, 1fr)); }
.detail-img { width: 100%; height: 130px; border-radius: 6px; cursor: pointer; }
.detail-img-sm { width: 100%; height: 100px; border-radius: 6px; cursor: pointer; }
.detail-video { width: 100%; max-height: 360px; border-radius: 6px; background: #000; }

/* ===== 修改记录 ===== */
.mod-detail { font-size: 13px; color: #555; }
.diff-line { margin-top: 4px; }
.diff-label { font-weight: 500; color: #666; }

/* ===== 处理记录 ===== */
.process-info-line { display: flex; flex-wrap: wrap; gap: 16px; margin-bottom: 12px; font-size: 13px; color: #666; }

/* ===== 追问追答 ===== */
.follow-list { display: flex; flex-direction: column; gap: 12px; }
.follow-item { padding: 12px; border-radius: 8px; }
.follow-item.question { background: #fef5e7; border-left: 3px solid #e6a23c; }
.follow-item.answer { background: #ecf5ff; border-left: 3px solid #409eff; }
.follow-header { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; }
.follow-user { font-weight: 500; color: #333; }
.follow-time { font-size: 12px; color: #999; margin-left: auto; }
.follow-body { font-size: 14px; color: #555; line-height: 1.7; }

/* ===== 申诉 ===== */
.appeal-info-line { display: flex; flex-wrap: wrap; gap: 16px; margin-bottom: 10px; font-size: 13px; color: #666; }
.appeal-reason { margin-top: 8px; font-size: 14px; color: #555; line-height: 1.7; }
.appeal-result { margin-top: 12px; }
.appeal-materials { margin-top: 16px; border-top: 1px solid #eee; padding-top: 12px; }
.appeal-materials h4 { margin: 0 0 8px; font-size: 14px; color: #333; }
.material-item { padding: 10px; background: #fafafa; border-radius: 6px; margin-bottom: 8px; }
.mat-header { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; }
.mat-submitter { font-weight: 500; color: #333; font-size: 13px; }
.mat-time { font-size: 12px; color: #999; margin-left: auto; }
.mat-content { font-size: 13px; color: #666; line-height: 1.6; }

/* ===== 操作 ===== */
.action-buttons { display: flex; flex-wrap: wrap; gap: 10px; }
.login-tip { text-align: center; color: #999; }

/* ===== 审核不通过 ===== */
.reject-alert { margin-top: 16px; }

/* ===== 空状态 ===== */
.empty-state { background: #fff; border-radius: 12px; padding: 80px; text-align: center; display: flex; flex-direction: column; align-items: center; gap: 20px; color: #909399; }
</style>
