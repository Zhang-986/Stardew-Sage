<template>
  <div class="stardew-profile">
    <!-- 星空背景 -->
    <div class="stardew-stars"></div>
    
    <el-row :gutter="20">
      <el-col :span="8" :xs="24">
        <div class="stardew-card farmer-info-card">
          <div class="card-header">
            <div class="header-icon">
              👤
            </div>
            <h3>老乡信息</h3>
          </div>

          <div class="card-body">
            <div class="avatar-section">
              <userAvatar />
              <div class="farmer-basic">
                <h4 class="farmer-name">🌾 {{ user.userName || '新手农夫' }}</h4>
                <span class="farmer-role">{{ roleGroup || '见习农夫' }}</span>
              </div>
            </div>

            <div class="info-list">
              <div class="info-row">
                <span class="info-icon">📱</span>
                <div class="info-content">
                  <span class="info-label">联系方式</span>
                  <span class="info-value">{{ user.phonenumber || '未设置' }}</span>
                </div>
              </div>
              <div class="info-row">
                <span class="info-icon">📧</span>
                <div class="info-content">
                  <span class="info-label">农场邮箱</span>
                  <span class="info-value">{{ user.email || '未设置' }}</span>
                </div>
              </div>
              <div class="info-row" v-if="user.dept">
                <span class="info-icon">🏢</span>
                <div class="info-content">
                  <span class="info-label">所属农业社</span>
                  <span class="info-value">{{ user.dept.deptName }}</span>
                </div>
              </div>
              <div class="info-row" v-if="postGroup">
                <span class="info-icon">💼</span>
                <div class="info-content">
                  <span class="info-label">农场职务</span>
                  <span class="info-value">{{ postGroup }}</span>
                </div>
              </div>
              <div class="info-row">
                <span class="info-icon">🗓️</span>
                <div class="info-content">
                  <span class="info-label">入驻时间</span>
                  <span class="info-value">{{ user.createTime }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-col>

      <el-col :span="16" :xs="24">
        <div class="stardew-card settings-card">
          <div class="card-header">
            <div class="header-icon">
              ⚙️
            </div>
            <h3>农场设置</h3>
          </div>

          <div class="card-body">
            <div class="stardew-tabs">
              <div
                class="tab-item"
                :class="{ active: selectedTab === 'userinfo' }"
                @click="selectedTab = 'userinfo'"
              >
                <i class="tab-icon">📝</i>
                <span>基本资料</span>
              </div>
              <div
                class="tab-item"
                :class="{ active: selectedTab === 'resetPwd' }"
                @click="selectedTab = 'resetPwd'"
              >
                <i class="tab-icon">🔐</i>
                <span>修改密码</span>
              </div>
            </div>

            <div class="tab-content">
              <userInfo v-if="selectedTab === 'userinfo'" :user="user" />
              <resetPwd v-if="selectedTab === 'resetPwd'" />
            </div>
          </div>
        </div>
      </el-col>
    </el-row>
    
    <!-- 装饰元素 -->
    <div class="stardew-decor soil"></div>
    <div class="stardew-decor grass"></div>
  </div>
</template>

<script>
import userAvatar from "./userAvatar"
import userInfo from "./userInfo"
import resetPwd from "./resetPwd"
import { getUserProfile } from "@/api/system/user"

export default {
  name: "Profile",
  components: { userAvatar, userInfo, resetPwd },
  data() {
    return {
      user: {},
      roleGroup: {},
      postGroup: {},
      selectedTab: "userinfo"
    }
  },
  created() {
    const activeTab = this.$route.params && this.$route.params.activeTab
    if (activeTab) {
      this.selectedTab = activeTab
    }
    this.getUser()
  },
  methods: {
    getUser() {
      getUserProfile().then(response => {
        this.user = response.data
        this.roleGroup = response.roleGroup
        this.postGroup = response.postGroup
      })
    }
  }
}
</script>

<style scoped lang="scss">
.stardew-profile {
  position: relative;
  min-height: calc(100vh - 60px);
  padding: 32px;
  background: var(--stardew-night-gradient);
  font-family: 'Pixel Arial', 'Microsoft YaHei', sans-serif;
  overflow: hidden;
  
  .stardew-stars {
    z-index: 1;
  }
  
  .stardew-card {
    @extend .stardew-card;
    background: var(--stardew-bg-panel);
    border: 4px solid var(--stardew-border-primary);
    border-radius: var(--stardew-radius-large);
    box-shadow: 0 6px 0 var(--stardew-border-shadow), var(--stardew-shadow-heavy);
    margin-bottom: 24px;
    overflow: hidden;
    transition: all 0.3s ease;
    position: relative;
    
    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 8px 0 var(--stardew-border-shadow), var(--stardew-shadow-heavy);
    }
    
    &::before {
      content: '';
      position: absolute;
      top: -4px;
      left: -4px;
      right: -4px;
      bottom: -4px;
      background: linear-gradient(45deg, var(--stardew-border-secondary), var(--stardew-border-primary));
      border-radius: var(--stardew-radius-large);
      z-index: -1;
    }
    
    .card-header {
      background: linear-gradient(135deg, var(--stardew-primary), var(--stardew-orange));
      color: var(--stardew-text-white);
      padding: 20px 24px;
      display: flex;
      align-items: center;
      border-bottom: 3px solid var(--stardew-border-shadow);
      
      .header-icon {
        width: 48px;
        height: 48px;
        border-radius: 50%;
        background: rgba(255, 255, 255, 0.2);
        display: flex;
        align-items: center;
        justify-content: center;
        margin-right: 16px;
        font-size: 24px;
        box-shadow: inset 0 2px 4px rgba(0,0,0,0.2);
      }
      
      h3 {
        margin: 0;
        font-size: 20px;
        font-weight: 600;
        text-shadow: 0 1px 2px rgba(0,0,0,0.3);
      }
    }
    
    .card-body {
      padding: 24px;
    }
  }
  
  .farmer-info-card {
    .avatar-section {
      text-align: center;
      padding-bottom: 24px;
      border-bottom: 2px dashed var(--stardew-border-primary);
      margin-bottom: 24px;
      
      .farmer-basic {
        margin-top: 16px;
        
        .farmer-name {
          margin: 0 0 12px 0;
          color: var(--stardew-text-primary);
          font-size: 18px;
          font-weight: 600;
          text-shadow: 0 1px 2px rgba(0,0,0,0.1);
        }
        
        .farmer-role {
          background: linear-gradient(45deg, var(--stardew-green), var(--stardew-green-light));
          color: var(--stardew-text-white);
          padding: 6px 16px;
          border-radius: 20px;
          font-size: 12px;
          font-weight: 600;
          display: inline-block;
          box-shadow: 0 2px 4px rgba(94, 158, 39, 0.3);
          text-shadow: 0 1px 2px rgba(0,0,0,0.2);
        }
      }
    }
    
    .info-list {
      .info-row {
        display: flex;
        align-items: center;
        padding: 16px 0;
        border-bottom: 1px solid var(--stardew-bg-secondary);
        transition: all 0.3s ease;
        
        &:last-child {
          border-bottom: none;
        }
        
        &:hover {
          background: rgba(139, 69, 19, 0.05);
          margin: 0 -16px;
          padding: 16px 16px;
          border-radius: var(--stardew-radius-small);
        }
        
        .info-icon {
          font-size: 20px;
          margin-right: 16px;
          width: 32px;
          text-align: center;
        }
        
        .info-content {
          flex: 1;
          display: flex;
          justify-content: space-between;
          align-items: center;
          
          .info-label {
            color: var(--stardew-text-secondary);
            font-size: 14px;
            font-weight: 500;
          }
          
          .info-value {
            color: var(--stardew-text-primary);
            font-size: 14px;
            font-weight: 600;
            text-align: right;
            max-width: 60%;
            word-break: break-all;
          }
        }
      }
    }
  }
  
  .settings-card {
    .stardew-tabs {
      display: flex;
      background: var(--stardew-bg-secondary);
      border: 2px solid var(--stardew-border-primary);
      border-radius: var(--stardew-radius-medium);
      padding: 6px;
      margin-bottom: 24px;
      
      .tab-item {
        flex: 1;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 12px 20px;
        border-radius: var(--stardew-radius-small);
        cursor: pointer;
        color: var(--stardew-text-secondary);
        font-weight: 500;
        transition: all 0.3s ease;
        gap: 8px;
        
        .tab-icon {
          font-size: 16px;
        }
        
        &:hover {
          color: var(--stardew-text-primary);
          background: rgba(139, 69, 19, 0.1);
        }
        
        &.active {
          background: linear-gradient(135deg, var(--stardew-primary), var(--stardew-orange));
          color: var(--stardew-text-white);
          box-shadow: 0 3px 0 var(--stardew-border-shadow);
          text-shadow: 0 1px 2px rgba(0,0,0,0.3);
          
          &::before {
            content: "✨";
            margin-right: 4px;
          }
        }
      }
    }
    
    .tab-content {
      min-height: 400px;
      background: rgba(255, 248, 230, 0.5);
      border: 2px dashed var(--stardew-border-primary);
      border-radius: var(--stardew-radius-small);
      padding: 20px;
    }
  }
  
  .stardew-decor {
    z-index: 0;
  }
}

// 响应式设计
@media (max-width: 992px) {
  .stardew-profile {
    padding: 20px;
    
    .el-col:first-child {
      margin-bottom: 20px;
    }
  }
}

@media (max-width: 768px) {
  .stardew-profile {
    padding: 16px;
    
    .card-header {
      padding: 16px 20px !important;
      
      .header-icon {
        width: 40px !important;
        height: 40px !important;
        margin-right: 12px !important;
        font-size: 20px !important;
      }
      
      h3 {
        font-size: 16px !important;
      }
    }
    
    .card-body {
      padding: 20px !important;
    }
    
    .info-list {
      .info-row {
        flex-direction: column;
        align-items: flex-start;
        gap: 8px;
        
        .info-content {
          width: 100%;
          flex-direction: column;
          align-items: flex-start;
          gap: 4px;
          
          .info-value {
            max-width: 100%;
            text-align: left;
          }
        }
      }
    }
    
    .stardew-tabs {
      .tab-item {
        padding: 10px 16px !important;
        font-size: 14px;
        
        .tab-icon {
          font-size: 14px !important;
        }
      }
    }
  }
}

// 覆盖Element UI样式
:deep(.el-card) {
  border: none;
  box-shadow: none;
}

:deep(.el-card__header) {
  padding: 0;
  border-bottom: none;
}

:deep(.el-card__body) {
  padding: 0;
}
</style>

