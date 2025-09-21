  <template>
    <div class="home">
      <!-- 系统标题 -->
      <div class="header-section">
        <h1 class="system-title">
          Aurora 极光管理系统
        </h1>
        <p class="system-subtitle">现代化企业级后台管理平台</p>
      </div>

      <!-- 卡片网格 -->
      <el-row :gutter="20" class="card-container">
        <!-- 系统简介 -->
        <el-col :xs="24" :sm="12" :md="8" :lg="8" :xl="8">
          <div class="info-card">
            <div class="card-header">
              <i class="el-icon-info"></i>
              <span>系统简介</span>
            </div>
            <div class="card-content">
              <p>基于Vue.js和Spring Boot构建的现代化企业级后台管理系统</p>
              <ul class="feature-list">
                <li>响应式设计，支持多端适配</li>
                <li>模块化架构，易于扩展</li>
                <li>完善的权限管理体系</li>
                <li>丰富的组件库支持</li>
              </ul>
            </div>
          </div>
        </el-col>

        <!-- 核心功能 -->
        <el-col :xs="24" :sm="12" :md="8" :lg="8" :xl="8">
          <div class="info-card">
            <div class="card-header">
              <i class="el-icon-setting"></i>
              <span>核心功能</span>
            </div>
            <div class="card-content">
              <ul class="function-list">
                <li>用户权限管理</li>
                <li>菜单角色配置</li>
                <li>系统日志监控</li>
                <li>数据统计分析</li>
                <li>系统工具集成</li>
                <li>系统参数配置</li>
              </ul>
            </div>
          </div>
        </el-col>

        <!-- 技术栈 -->
        <el-col :xs="24" :sm="12" :md="8" :lg="8" :xl="8">
          <div class="info-card">
            <div class="card-header">
              <i class="el-icon-cpu"></i>
              <span>技术栈</span>
            </div>
            <div class="card-content">
              <div class="tech-section">
                <h4>前端技术</h4>
                <p>Vue.js 2.x、Element UI、Vue Router、Vuex</p>
              </div>
              <div class="tech-section">
                <h4>后端技术</h4>
                <p>Spring Boot、Spring Security、MyBatis Plus、Redis</p>
              </div>
            </div>
          </div>
        </el-col>


      </el-row>
    </div>
  </template>

  <script>
  export default {
    name: "Index",
    data() {
      return {
        // 版本号
        version: "3.9.0"
      }
    },
    methods: {
      goTarget(href) {
        window.open(href, "_blank")
      }
    }
  }
  </script>

  <style scoped lang="scss">
  .home {
    padding: 20px;
    background: #f8f9fa;
    min-height: calc(100vh - 60px);

    .header-section {
      text-align: center;
      margin-bottom: 40px;
      padding: 40px 0;
      background: linear-gradient(135deg, #000000 0%, #3d1b1b 100%);
      border-radius: 12px;
      color: #fff;
      margin: 0 0 40px 0;

      .system-title {
        font-size: 28px;
        margin-bottom: 8px;
        font-weight: 400;
      }

      .system-subtitle {
        font-size: 16px;
        opacity: 0.9;
        margin: 0;
      }
    }

    .info-card {
      background: #fff;
      border-radius: 8px;
      box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
      margin-bottom: 20px;
      overflow: hidden;
      transition: all 0.3s ease;

      &:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
      }

      .card-header {
        background: #f8f9fa;
        padding: 16px 20px;
        border-bottom: 1px solid #e9ecef;
        display: flex;
        align-items: center;
        color: #2c3e50;
        font-weight: 500;

        i {
          margin-right: 8px;
          font-size: 16px;
          color: #6c757d;
        }
      }

      .card-content {
        padding: 20px;

        p {
          color: #666;
          line-height: 1.6;
          margin-bottom: 16px;
          font-size: 14px;
        }

        .feature-list, .function-list {
          list-style: none;
          padding: 0;
          margin: 0;

          li {
            padding: 8px 0;
            color: #555;
            font-size: 14px;
            border-bottom: 1px solid #f0f2f5;
            position: relative;
            padding-left: 16px;

            &:before {
              content: '•';
              color: #2c3e50;
              font-weight: bold;
              position: absolute;
              left: 0;
            }

            &:last-child {
              border-bottom: none;
            }
          }
        }

        .tech-section {
          margin-bottom: 16px;

          &:last-child {
            margin-bottom: 0;
          }

          h4 {
            color: #2c3e50;
            font-size: 14px;
            margin: 0 0 8px 0;
            font-weight: 600;
          }

          p {
            color: #666;
            font-size: 13px;
            margin: 0;
            line-height: 1.5;
          }
        }

        .feature-grid {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 16px;

          .feature-item {
            text-align: center;

            h5 {
              color: #2c3e50;
              font-size: 14px;
              margin: 0 0 6px 0;
              font-weight: 600;
            }

            p {
              color: #666;
              font-size: 12px;
              margin: 0;
              line-height: 1.4;
            }
          }
        }

        .version-info {
          margin-bottom: 20px;

          .version-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 8px 0;
            border-bottom: 1px solid #f0f2f5;

            &:last-child {
              border-bottom: none;
            }

            .label {
              color: #666;
              font-size: 14px;
            }

            .value {
              color: #2c3e50;
              font-weight: 600;
              font-size: 14px;
            }
          }
        }

        .action-buttons {
          text-align: center;

          .el-button {
            margin: 0 4px;
          }
        }
      }
    }
  }

  // 响应式设计
  @media (max-width: 768px) {
    .home {
      padding: 15px;

      .header-section {
        padding: 30px 20px;

        .system-title {
          font-size: 24px;
        }

        .system-subtitle {
          font-size: 14px;
        }
      }

      .info-card {
        .card-header {
          padding: 14px 16px;
          font-size: 14px;
        }

        .card-content {
          padding: 16px;
        }

        .feature-grid {
          grid-template-columns: 1fr;
          gap: 12px;
        }
      }
    }
  }
  </style>

