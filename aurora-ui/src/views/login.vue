<template>
  <div class="login">
    <div class="login-container">
      <div class="login-brand">
        <h1 class="brand-title">
          <span class="brand-icon">🌙</span>
          Aurora
        </h1>
        <p class="brand-subtitle">极光管理系统</p>
      </div>
      
      <el-form ref="loginForm" :model="loginForm" :rules="loginRules" class="login-form">
        <h3 class="title">{{title}}</h3>
        <el-form-item prop="username">
          <el-input
            v-model="loginForm.username"
            type="text"
            auto-complete="off"
            placeholder="账号"
          >
            <svg-icon slot="prefix" icon-class="user" class="el-input__icon input-icon" />
          </el-input>
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            auto-complete="off"
            placeholder="密码"
            @keyup.enter.native="handleLogin"
          >
            <svg-icon slot="prefix" icon-class="password" class="el-input__icon input-icon" />
          </el-input>
        </el-form-item>
        <el-form-item prop="code" v-if="captchaEnabled">
          <el-input
            v-model="loginForm.code"
            auto-complete="off"
            placeholder="验证码"
            style="width: 63%"
            @keyup.enter.native="handleLogin"
          >
            <svg-icon slot="prefix" icon-class="validCode" class="el-input__icon input-icon" />
          </el-input>
          <div class="login-code">
            <img :src="codeUrl" @click="getCode" class="login-code-img"/>
          </div>
        </el-form-item>
        <el-checkbox v-model="loginForm.rememberMe" style="margin:0px 0px 25px 0px;">记住密码</el-checkbox>
        <el-form-item style="width:100%;">
          <el-button
            :loading="loading"
            size="medium"
            type="primary"
            style="width:100%;"
            @click.native.prevent="handleLogin"
          >
            <span v-if="!loading">登 录</span>
            <span v-else>登 录 中...</span>
          </el-button>
          <div style="float: right;" v-if="register">
            <router-link class="link-type" :to="'/register'">立即注册</router-link>
          </div>
        </el-form-item>
      </el-form>
    </div>
    
    <!-- 装饰元素 -->
    <div class="stars">
      <div class="star"></div>
      <div class="star"></div>
      <div class="star"></div>
      <div class="star"></div>
      <div class="star"></div>
    </div>
  </div>
</template>

<script>
import { getCodeImg } from "@/api/login"
import Cookies from "js-cookie"
import { encrypt, decrypt } from '@/utils/jsencrypt'

export default {
  name: "Login",
  data() {
    return {
      title: process.env.VUE_APP_TITLE,
      codeUrl: "",
      loginForm: {
        username: "admin",
        password: "admin123",
        rememberMe: false,
        code: "",
        uuid: ""
      },
      loginRules: {
        username: [
          { required: true, trigger: "blur", message: "请输入您的账号" }
        ],
        password: [
          { required: true, trigger: "blur", message: "请输入您的密码" }
        ],
        code: [{ required: true, trigger: "change", message: "请输入验证码" }]
      },
      loading: false,
      // 验证码开关
      captchaEnabled: true,
      // 注册开关
      register: true,
      redirect: undefined
    }
  },
  watch: {
    $route: {
      handler: function(route) {
        this.redirect = route.query && route.query.redirect
      },
      immediate: true
    }
  },
  created() {
    this.getCode()
    this.getCookie()
  },
  methods: {
    getCode() {
      getCodeImg().then(res => {
        this.captchaEnabled = res.captchaEnabled === undefined ? true : res.captchaEnabled
        if (this.captchaEnabled) {
          this.codeUrl = "data:image/gif;base64," + res.img
          this.loginForm.uuid = res.uuid
        }
      })
    },
    getCookie() {
      const username = Cookies.get("username")
      const password = Cookies.get("password")
      const rememberMe = Cookies.get('rememberMe')
      this.loginForm = {
        username: username === undefined ? this.loginForm.username : username,
        password: password === undefined ? this.loginForm.password : decrypt(password),
        rememberMe: rememberMe === undefined ? false : Boolean(rememberMe)
      }
    },
    handleLogin() {
      this.$refs.loginForm.validate(valid => {
        if (valid) {
          this.loading = true
          if (this.loginForm.rememberMe) {
            Cookies.set("username", this.loginForm.username, { expires: 30 })
            Cookies.set("password", encrypt(this.loginForm.password), { expires: 30 })
            Cookies.set('rememberMe', this.loginForm.rememberMe, { expires: 30 })
          } else {
            Cookies.remove("username")
            Cookies.remove("password")
            Cookies.remove('rememberMe')
          }
          this.$store.dispatch("Login", this.loginForm).then(() => {
            this.$router.push({ path: this.redirect || "/" }).catch(()=>{})
          }).catch(() => {
            this.loading = false
            if (this.captchaEnabled) {
              this.getCode()
            }
          })
        }
      })
    }
  }
}
</script>

<style rel="stylesheet/scss" lang="scss">
.login {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: linear-gradient(135deg, #0c1445 0%, #1a2b5c 50%, #2d4a8a 100%);
  position: relative;
  overflow: hidden;
  
  // 添加动态背景效果
  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: radial-gradient(circle at 20% 80%, rgba(120, 200, 255, 0.1) 0%, transparent 50%),
                radial-gradient(circle at 80% 20%, rgba(255, 150, 120, 0.1) 0%, transparent 50%),
                radial-gradient(circle at 40% 40%, rgba(150, 255, 150, 0.05) 0%, transparent 50%);
    animation: aurora 20s ease-in-out infinite;
  }
  
  .login-container {
    display: flex;
    align-items: center;
    gap: 60px;
    z-index: 2;
    position: relative;
  }
  
  .login-brand {
    text-align: center;
    color: white;
    
    .brand-title {
      font-size: 48px;
      font-weight: 300;
      margin: 0 0 16px 0;
      background: linear-gradient(45deg, #4fc3f7, #29b6f6, #03a9f4);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      background-clip: text;
      text-shadow: 0 0 30px rgba(79, 195, 247, 0.3);
      
      .brand-icon {
        display: block;
        font-size: 64px;
        margin-bottom: 16px;
        animation: float 3s ease-in-out infinite;
      }
    }
    
    .brand-subtitle {
      font-size: 18px;
      color: rgba(255, 255, 255, 0.7);
      margin: 0;
      letter-spacing: 2px;
    }
  }
  
  .login-form {
    border-radius: 20px;
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(20px);
    width: 400px;
    padding: 40px 35px;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3),
                0 0 100px rgba(79, 195, 247, 0.1);
    border: 1px solid rgba(255, 255, 255, 0.2);
    
    .title {
      margin: 0px auto 40px auto;
      text-align: center;
      color: #2c3e50;
      font-size: 24px;
      font-weight: 500;
    }
    
    .el-form-item {
      margin-bottom: 24px;
      
      .el-input {
        height: 48px;
        
        input {
          height: 48px;
          border-radius: 12px;
          border: 2px solid #e1e8ed;
          background: rgba(255, 255, 255, 0.9);
          transition: all 0.3s ease;
          
          &:focus {
            border-color: #4fc3f7;
            box-shadow: 0 0 20px rgba(79, 195, 247, 0.2);
          }
        }
      }
      
      .input-icon {
        height: 48px;
        width: 16px;
        margin-left: 4px;
        color: #4fc3f7;
      }
    }
    
    .el-checkbox {
      color: #5a6c7d;
      
      .el-checkbox__input.is-checked .el-checkbox__inner {
        background-color: #4fc3f7;
        border-color: #4fc3f7;
      }
    }
    
    .el-button--primary {
      background: linear-gradient(45deg, #4fc3f7, #29b6f6);
      border: none;
      border-radius: 12px;
      height: 48px;
      font-size: 16px;
      font-weight: 500;
      box-shadow: 0 8px 25px rgba(79, 195, 247, 0.3);
      transition: all 0.3s ease;
      
      &:hover {
        background: linear-gradient(45deg, #29b6f6, #03a9f4);
        transform: translateY(-2px);
        box-shadow: 0 12px 35px rgba(79, 195, 247, 0.4);
      }
    }
    
    .link-type {
      color: #4fc3f7;
      text-decoration: none;
      font-size: 14px;
      
      &:hover {
        color: #29b6f6;
      }
    }
  }
  
  .login-code {
    width: 33%;
    height: 48px;
    float: right;
    
    .login-code-img {
      height: 48px;
      border-radius: 8px;
      cursor: pointer;
      transition: all 0.3s ease;
      
      &:hover {
        transform: scale(1.05);
      }
    }
  }
  
  // 星星装饰
  .stars {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
    
    .star {
      position: absolute;
      width: 2px;
      height: 2px;
      background: white;
      border-radius: 50%;
      animation: twinkle 3s infinite;
      
      &:nth-child(1) {
        top: 20%;
        left: 15%;
        animation-delay: 0s;
      }
      
      &:nth-child(2) {
        top: 60%;
        left: 80%;
        animation-delay: 1s;
      }
      
      &:nth-child(3) {
        top: 30%;
        left: 70%;
        animation-delay: 2s;
      }
      
      &:nth-child(4) {
        top: 80%;
        left: 20%;
        animation-delay: 1.5s;
      }
      
      &:nth-child(5) {
        top: 10%;
        left: 50%;
        animation-delay: 0.5s;
      }
    }
  }
}

// 动画效果
@keyframes aurora {
  0%, 100% {
    opacity: 0.5;
    transform: rotate(0deg) scale(1);
  }
  50% {
    opacity: 0.8;
    transform: rotate(180deg) scale(1.1);
  }
}

@keyframes float {
  0%, 100% {
    transform: translateY(0px);
  }
  50% {
    transform: translateY(-10px);
  }
}

@keyframes twinkle {
  0%, 100% {
    opacity: 0.3;
    transform: scale(1);
  }
  50% {
    opacity: 1;
    transform: scale(1.2);
  }
}

// 响应式设计
@media (max-width: 1024px) {
  .login {
    .login-container {
      flex-direction: column;
      gap: 40px;
    }
    
    .login-brand {
      .brand-title {
        font-size: 36px;
        
        .brand-icon {
          font-size: 48px;
        }
      }
    }
  }
}

@media (max-width: 480px) {
  .login {
    padding: 20px;
    
    .login-form {
      width: 100%;
      max-width: 350px;
      padding: 30px 25px;
    }
  }
}

// 保留原有样式兼容
.login-tip {
  font-size: 13px;
  text-align: center;
  color: #bfbfbf;
}

.el-login-footer {
  height: 40px;
  line-height: 40px;
  position: fixed;
  bottom: 0;
  width: 100%;
  text-align: center;
  color: #fff;
  font-family: Arial;
  font-size: 12px;
  letter-spacing: 1px;
}
</style>
