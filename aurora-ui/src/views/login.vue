<template>
  <div class="stardew-login">
    <!-- 星空背景 -->
    <div class="stardew-stars"></div>
    
    <div class="login-container">
      <!-- 品牌区域 -->
      <div class="login-brand">
        <div class="brand-icon">🌙</div>
        <h1 class="brand-title">Stardew Sage</h1>
        <div class="brand-description">
          <p>欢迎来到老乡之家</p>
          <p>这里可以提出来你对星露谷物语的一切疑惑</p>
        </div>
      </div>
      
      <!-- 登录表单 -->
      <div class="login-form-wrapper">
        <div class="form-header">
          <h3 class="form-title">{{title}}</h3>
          <p class="form-subtitle">请输入您的账户信息</p>
        </div>
        
        <el-form ref="loginForm" :model="loginForm" :rules="loginRules" class="stardew-form">
          <el-form-item prop="username">
            <div class="input-wrapper">
              <svg-icon icon-class="user" class="input-icon" />
              <el-input
                v-model="loginForm.username"
                type="text"
                auto-complete="off"
                placeholder="输入您的账号"
                class="stardew-input"
              />
            </div>
          </el-form-item>
          
          <el-form-item prop="password">
            <div class="input-wrapper">
              <svg-icon icon-class="password" class="input-icon" />
              <el-input
                v-model="loginForm.password"
                type="password"
                auto-complete="off"
                placeholder="输入您的密码"
                class="stardew-input"
                @keyup.enter.native="handleLogin"
              />
            </div>
          </el-form-item>
          
          <el-form-item prop="code" v-if="captchaEnabled">
            <div class="captcha-wrapper">
              <div class="input-wrapper captcha-input">
                <svg-icon icon-class="validCode" class="input-icon" />
                <el-input
                  v-model="loginForm.code"
                  auto-complete="off"
                  placeholder="验证码"
                  class="stardew-input"
                  @keyup.enter.native="handleLogin"
                />
              </div>
              <div class="captcha-image" @click="getCode">
                <img :src="codeUrl" class="captcha-img"/>
                <div class="captcha-refresh">点击刷新</div>
              </div>
            </div>
          </el-form-item>
          
          <el-form-item class="form-options">
            <el-checkbox v-model="loginForm.rememberMe" class="stardew-checkbox">
              记住我的登录状态
            </el-checkbox>
          </el-form-item>
          
          <el-form-item class="form-actions">
            <el-button
              :loading="loading"
              type="primary"
              class="login-btn stardew-btn"
              @click="handleLogin"
            >
              <span v-if="!loading">🚀 开始农场之旅</span>
              <span v-else>正在进入农场...</span>
            </el-button>
          </el-form-item>
          
          <div class="form-footer" v-if="register">
            <p>还没有账号？ 
              <router-link class="register-link" :to="'/register'">立即注册</router-link>
            </p>
          </div>
        </el-form>
      </div>
    </div>

    <!-- 装饰元素 -->
    <div class="stardew-decor soil"></div>
    <div class="stardew-decor grass"></div>
    
    <!-- 浮动元素 -->
    <div class="floating-elements">
      <div class="floating-crop">🌽</div>
      <div class="floating-crop">🥕</div>
      <div class="floating-crop">🍅</div>
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
.stardew-login {
  position: relative;
  min-height: 100vh;
  background: var(--stardew-night-gradient);
  display: flex;
  justify-content: center;
  align-items: center;
  overflow: hidden;
  font-family: 'Pixel Arial', 'Microsoft YaHei', sans-serif;
  
  .stardew-stars {
    z-index: 1;
  }
  
  .login-container {
    display: flex;
    align-items: center;
    gap: 80px;
    z-index: 2;
    position: relative;
  }
  
  .login-brand {
    text-align: center;
    color: var(--stardew-text-white);
    
    .brand-icon {
      font-size: 72px;
      margin-bottom: 20px;
      animation: stardew-float 4s ease-in-out infinite;
    }
    
    .brand-title {
      font-size: 48px;
      font-weight: 600;
      margin: 0 0 16px 0;
      color: #FFD04D;
      text-shadow: 0 3px 0 #B8860B, 0 6px 8px rgba(0,0,0,.4);
      letter-spacing: 3px;
    }
    
    .brand-subtitle {
      font-size: 18px;
      margin: 0 0 24px 0;
      opacity: 0.9;
      text-shadow: 0 2px 4px rgba(0,0,0,.4);
    }
    
    .brand-description {
      p {
        margin: 8px 0;
        font-size: 14px;
        opacity: 0.8;
        text-shadow: 0 1px 2px rgba(0,0,0,.4);
      }
    }
  }
  
  .login-form-wrapper {
    background: var(--stardew-bg-panel);
    border: 4px solid var(--stardew-border-primary);
    border-radius: var(--stardew-radius-large);
    padding: 40px 35px;
    width: 420px;
    box-shadow: 
      0 6px 0 var(--stardew-border-shadow),
      0 12px 24px rgba(0,0,0,.3);
    position: relative;
    
    &::before {
      content: '';
      position: absolute;
      top: -4px;
      left: -4px;
      right: -4px;
      bottom: -4px;
      background: linear-gradient(45deg, #D4AF37, var(--stardew-border-primary));
      border-radius: var(--stardew-radius-large);
      z-index: -1;
    }
    
    .form-header {
      text-align: center;
      margin-bottom: 32px;
      
      .form-title {
        color: var(--stardew-text-primary);
        font-size: 24px;
        font-weight: 600;
        margin: 0 0 8px 0;
        text-shadow: 0 1px 2px rgba(0,0,0,.1);
      }
      
      .form-subtitle {
        color: var(--stardew-text-secondary);
        font-size: 14px;
        margin: 0;
      }
    }
    
    .stardew-form {
      .el-form-item {
        margin-bottom: 24px;
        
        .input-wrapper {
          position: relative;
          
          .input-icon {
            position: absolute;
            left: 12px;
            top: 50%;
            transform: translateY(-50%);
            color: var(--stardew-text-secondary);
            z-index: 2;
          }
          
          .el-input__inner {
            padding-left: 40px;
            height: 48px;
            border: 3px solid var(--stardew-border-primary);
            border-radius: var(--stardew-radius-small);
            background: #FFFEF7;
            color: var(--stardew-text-primary);
            font-size: 14px;
            transition: all 0.3s ease;
            
            &:focus {
              border-color: var(--stardew-orange);
              box-shadow: 0 0 12px rgba(255, 140, 0, 0.3);
            }
            
            &::placeholder {
              color: var(--stardew-text-light);
            }
          }
        }
        
        .captcha-wrapper {
          display: flex;
          gap: 12px;
          
          .captcha-input {
            flex: 1;
          }
          
          .captcha-image {
            width: 120px;
            height: 48px;
            border: 3px solid var(--stardew-border-primary);
            border-radius: var(--stardew-radius-small);
            overflow: hidden;
            cursor: pointer;
            position: relative;
            background: #FFF;
            transition: all 0.3s ease;
            
            &:hover {
              transform: scale(1.05);
              border-color: var(--stardew-orange);
            }
            
            .captcha-img {
              width: 100%;
              height: 100%;
              object-fit: cover;
            }
            
            .captcha-refresh {
              position: absolute;
              bottom: 0;
              left: 0;
              right: 0;
              background: rgba(0,0,0,0.7);
              color: white;
              font-size: 10px;
              text-align: center;
              padding: 2px;
              transform: translateY(100%);
              transition: all 0.3s ease;
            }
            
            &:hover .captcha-refresh {
              transform: translateY(0);
            }
          }
        }
      }
      
      .form-options {
        .stardew-checkbox {
          .el-checkbox__input.is-checked .el-checkbox__inner {
            background-color: var(--stardew-primary);
            border-color: var(--stardew-primary);
          }
          
          .el-checkbox__label {
            color: var(--stardew-text-secondary);
            font-size: 14px;
          }
        }
      }
      
      .form-actions {
        .login-btn {
          width: 100%;
          height: 52px;
          font-size: 16px;
          font-weight: 600;
          background: linear-gradient(135deg, var(--stardew-primary), var(--stardew-orange));
          border: 3px solid var(--stardew-border-primary);
          color: var(--stardew-text-white);
          border-radius: var(--stardew-radius-small);
          box-shadow: 0 4px 0 var(--stardew-border-shadow);
          transition: all 0.3s ease;
          
          &:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 0 var(--stardew-border-shadow);
            background: linear-gradient(135deg, var(--stardew-orange), var(--stardew-primary));
          }
          
          &:active {
            transform: translateY(1px);
            box-shadow: 0 3px 0 var(--stardew-border-shadow);
          }
        }
      }
      
      .form-footer {
        text-align: center;
        margin-top: 20px;
        
        p {
          color: var(--stardew-text-secondary);
          font-size: 14px;
          margin: 0;
        }
        
        .register-link {
          color: var(--stardew-orange);
          text-decoration: none;
          font-weight: 600;
          
          &:hover {
            color: var(--stardew-primary);
          }
        }
      }
    }
  }
  
  .stardew-decor {
    z-index: 0;
  }
  
  .floating-elements {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    pointer-events: none;
    z-index: 1;
    
    .floating-crop {
      position: absolute;
      font-size: 24px;
      animation: stardew-float 6s ease-in-out infinite;
      opacity: 0.6;
      
      &:nth-child(1) {
        top: 15%;
        left: 10%;
        animation-delay: 0s;
      }
      
      &:nth-child(2) {
        top: 60%;
        right: 8%;
        animation-delay: 2s;
        animation-duration: 8s;
      }
      
      &:nth-child(3) {
        bottom: 25%;
        left: 15%;
        animation-delay: 4s;
        animation-duration: 7s;
      }
    }
  }
}

// 响应式设计
@media (max-width: 1200px) {
  .stardew-login .login-container {
    flex-direction: column;
    gap: 40px;
  }
}

@media (max-width: 768px) {
  .stardew-login {
    padding: 20px;
    
    .login-form-wrapper {
      width: 100%;
      max-width: 400px;
      padding: 30px 25px;
    }
    
    .login-brand {
      .brand-title {
        font-size: 36px;
      }
      
      .brand-icon {
        font-size: 56px;
      }
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
