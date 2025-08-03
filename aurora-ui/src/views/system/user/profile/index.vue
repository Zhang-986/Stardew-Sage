<template>
  <div class="profile-container">
    <el-row :gutter="20">
      <el-col :span="8" :xs="24">
        <div class="profile-card user-card">
          <div class="card-header">
            <div class="header-icon">
              <i class="el-icon-user"></i>
            </div>
            <h3>个人信息</h3>
          </div>
          
          <div class="card-body">
            <div class="avatar-section">
              <userAvatar />
              <div class="user-basic">
                <h4 class="user-name">{{ user.userName || '用户' }}</h4>
                <span class="user-role">{{ roleGroup || '暂无角色' }}</span>
              </div>
            </div>
            
            <div class="info-list">
              <div class="info-row">
                <span class="info-label">手机号码</span>
                <span class="info-value">{{ user.phonenumber || '未设置' }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">用户邮箱</span>
                <span class="info-value">{{ user.email || '未设置' }}</span>
              </div>
              <div class="info-row" v-if="user.dept">
                <span class="info-label">所属部门</span>
                <span class="info-value">{{ user.dept.deptName }}</span>
              </div>
              <div class="info-row" v-if="postGroup">
                <span class="info-label">岗位信息</span>
                <span class="info-value">{{ postGroup }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">创建时间</span>
                <span class="info-value">{{ user.createTime }}</span>
              </div>
            </div>
          </div>
        </div>
      </el-col>
      
      <el-col :span="16" :xs="24">
        <div class="profile-card settings-card">
          <div class="card-header">
            <div class="header-icon">
              <i class="el-icon-setting"></i>
            </div>
            <h3>基本资料</h3>
          </div>
          
          <div class="card-body">
            <div class="tab-navigation">
              <div 
                class="nav-item" 
                :class="{ active: selectedTab === 'userinfo' }"
                @click="selectedTab = 'userinfo'"
              >
                <i class="el-icon-edit"></i>
                <span>基本资料</span>
              </div>
              <div 
                class="nav-item" 
                :class="{ active: selectedTab === 'resetPwd' }"
                @click="selectedTab = 'resetPwd'"
              >
                <i class="el-icon-lock"></i>
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
.profile-container {
  padding: 32px;
  background: linear-gradient(135deg, #ffffff 0%, #513f61 100%);
  min-height: calc(100vh - 60px);
  
  .profile-card {
    background: #fff;
    border-radius: 16px;
    margin-bottom: 24px;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
    transition: all 0.3s cubic-bezier(0.25, 0.8, 0.25, 1);
    overflow: hidden;
    
    &:hover {
      transform: translateY(-4px);
      box-shadow: 0 16px 48px rgba(0, 0, 0, 0.2);
    }
    
    .card-header {
      background: linear-gradient(135deg, #2c3e50 0%, #34495e 100%);
      color: #fff;
      padding: 24px;
      display: flex;
      align-items: center;
      
      .header-icon {
        width: 48px;
        height: 48px;
        border-radius: 50%;
        background: rgba(255, 255, 255, 0.2);
        display: flex;
        align-items: center;
        justify-content: center;
        margin-right: 16px;
        
        i {
          font-size: 20px;
        }
      }
      
      h3 {
        margin: 0;
        font-size: 18px;
        font-weight: 500;
      }
    }
    
    .card-body {
      padding: 24px;
    }
  }
  
  .user-card {
    .avatar-section {
      text-align: center;
      padding-bottom: 24px;
      border-bottom: 1px solid #f0f2f5;
      margin-bottom: 24px;
      
      .user-basic {
        margin-top: 16px;
        
        .user-name {
          margin: 0 0 8px 0;
          color: #2c3e50;
          font-size: 18px;
          font-weight: 600;
        }
        
        .user-role {
          background: linear-gradient(45deg, #667eea, #764ba2);
          color: #fff;
          padding: 4px 16px;
          border-radius: 20px;
          font-size: 12px;
          font-weight: 500;
        }
      }
    }
    
    .info-list {
      .info-row {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 12px 0;
        border-bottom: 1px solid #f8f9fa;
        
        &:last-child {
          border-bottom: none;
        }
        
        &:hover {
          background: #f8f9fa;
          margin: 0 -16px;
          padding: 12px 16px;
          border-radius: 8px;
        }
        
        .info-label {
          color: #666;
          font-size: 14px;
          font-weight: 500;
        }
        
        .info-value {
          color: #2c3e50;
          font-size: 14px;
          font-weight: 600;
          text-align: right;
          max-width: 60%;
          word-break: break-all;
        }
      }
    }
  }
  
  .settings-card {
    .tab-navigation {
      display: flex;
      background: #f8f9fa;
      border-radius: 12px;
      padding: 4px;
      margin-bottom: 24px;
      
      .nav-item {
        flex: 1;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 12px 20px;
        border-radius: 8px;
        cursor: pointer;
        color: #666;
        font-weight: 500;
        transition: all 0.3s ease;
        gap: 8px;
        
        i {
          font-size: 16px;
        }
        
        &:hover {
          color: #2c3e50;
          background: rgba(44, 62, 80, 0.1);
        }
        
        &.active {
          background: #fff;
          color: #2c3e50;
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
        }
      }
    }
    
    .tab-content {
      min-height: 400px;
    }
  }
}

// 响应式设计
@media (max-width: 992px) {
  .profile-container {
    padding: 20px;
    
    .el-col:first-child {
      margin-bottom: 20px;
    }
  }
}

@media (max-width: 768px) {
  .profile-container {
    padding: 16px;
    
    .card-header {
      padding: 20px !important;
      
      .header-icon {
        width: 40px !important;
        height: 40px !important;
        margin-right: 12px !important;
        
        i {
          font-size: 18px !important;
        }
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
        gap: 4px;
        
        .info-value {
          max-width: 100%;
          text-align: left;
        }
      }
    }
    
    .tab-navigation {
      .nav-item {
        padding: 10px 16px !important;
        font-size: 14px;
        
        i {
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

