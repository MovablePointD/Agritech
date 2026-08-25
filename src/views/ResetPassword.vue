<template>
  <div class="reset-container">
    <el-card class="reset-card">
      <h2>重置密码</h2>
      <p class="subtitle">输入用户名和新密码即可重置</p>
      <el-form :model="form" :rules="rules" ref="formRef">
        <el-form-item prop="username">
          <el-input v-model="form.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item prop="newPassword">
          <el-input v-model="form.newPassword" type="password" placeholder="请输入新密码（至少6位）" show-password />
        </el-form-item>
        <el-form-item prop="confirmPassword">
          <el-input v-model="form.confirmPassword" type="password" placeholder="请再次输入新密码" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" style="width: 100%" @click="handleReset" :loading="loading">重置密码</el-button>
        </el-form-item>
        <div class="link-group">
          <router-link to="/login">返回登录</router-link>
          <span class="divider">|</span>
          <router-link to="/register">注册新账号</router-link>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import api from '@/utils/api'

const router = useRouter()
const formRef = ref()
const loading = ref(false)

const form = reactive({
  username: '',
  newPassword: '',
  confirmPassword: ''
})

const validateConfirmPassword = (rule, value, callback) => {
  if (value !== form.newPassword) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少为6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

const handleReset = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await api.resetPassword({
      username: form.username,
      new_password: form.newPassword
    })
    ElMessage.success('密码重置成功，请使用新密码登录')
    router.push('/login')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || error?.message || '重置失败，请检查用户名是否正确')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.reset-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.reset-card {
  width: 400px;
  padding: 20px;
}

h2 {
  text-align: center;
  margin-bottom: 5px;
  color: #333;
}

.subtitle {
  text-align: center;
  color: #999;
  font-size: 14px;
  margin-bottom: 20px;
}

.link-group {
  text-align: center;
  margin-top: 10px;
}

.link-group a {
  color: #409eff;
  text-decoration: none;
}

.link-group .divider {
  margin: 0 10px;
  color: #ccc;
}
</style>
