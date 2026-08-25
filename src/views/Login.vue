<template>
  <div class="login-container">
    <el-card class="login-card">
      
      <h2>登录</h2>
      <el-form :model="form" :rules="rules" ref="formRef">
        <el-form-item prop="username">
          <el-input v-model="form.username" placeholder="用户名" prefix-icon="User" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码" prefix-icon="Lock" show-password />
        </el-form-item>
        <el-form-item></el-form-item>
          <el-button type="primary" style="width: 100%" @click="handleLogin" :loading="loading">登录</el-button>
        
        <div class="register-link">
          <router-link to="/reset-password">忘记密码？</router-link>
          <span class="divider">|</span>
          还没有账号？<router-link to="/register">立即注册</router-link>
        </div>
      </el-form>

      <div class="breadcrumb">
      <el-button text @click="$router.back()">
         返回上一级
      </el-button>
     
      <el-button text @click="$router.push('/')">
        返回首页
      </el-button>
    </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import api from '@/utils/api'
import { setToken, setUserInfo } from '@/utils/auth'

const router = useRouter()
const formRef = ref()
const loading = ref(false)

const form = reactive({
  username: '',
  password: ''
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const handleLogin = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  
  loading.value = true
  try {
    const res = await api.login(form)
    setToken(res.token)
    // 登录后立即获取用户信息并存储，确保 AppHeader 能立即展示
    try {
      const userRes = await api.getUserInfo()
      setUserInfo(userRes)
    } catch {
      // 获取用户信息失败不影响登录流程
    }
    ElMessage.success('登录成功')
    router.push('/')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 400px;
  padding: 20px;
}

h2 {
  text-align: center;
  margin-bottom: 30px;
  color: #333;
}

.register-link {
  text-align: center;
  margin-top: 10px;
}

.register-link a {
  color: #409eff;
  text-decoration: none;
}

.register-link .divider {
  margin: 0 10px;
  color: #ccc;
}

.breadcrumb {

  margin-top: 33px;
}

</style>
