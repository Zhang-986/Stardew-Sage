<template>
  <div class="stardew-home">
    <div class="sv-header">
      <h1 class="title">🌾 星露谷农场管理中心</h1>
      <p class="subtitle">Aurora 极光管理系统</p>
    </div>

    <!-- 第一行：生日故事和每日任务 -->
    <el-row :gutter="20" class="main-content">
      <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
        <div class="panel story-panel">
          <div class="panel-header">
            <span class="panel-title">📖 今日生日人物故事</span>
            <span class="status" :class="storyStatus">{{ storyStatusText }}</span>
          </div>
          <div class="output" ref="storyOutputBox" v-html="storyDisplayHtml"></div>
        </div>
      </el-col>

      <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
        <div class="panel mission-panel">
          <div class="panel-header">
            <span class="panel-title">🎯 每日任务</span>
            <span class="status" :class="missionStatus">{{ missionStatusText }}</span>
          </div>
          <div class="output" ref="missionOutputBox" v-html="missionDisplayHtml"></div>
        </div>
      </el-col>
    </el-row>

    <!-- 第二行：快捷操作 -->
    <el-row :gutter="20" class="quick-actions">
      <el-col :xs="24" :sm="12" :md="8" :lg="6" :xl="6">
        <div class="action-card">
          <div class="action-icon">🌱</div>
          <div class="action-title">农场管理</div>
          <div class="action-desc">管理您的数字农场</div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="12" :md="8" :lg="6" :xl="6">
        <div class="action-card">
          <div class="action-icon">👥</div>
          <div class="action-title">老乡管理</div>
          <div class="action-desc">查看社区成员信息</div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="12" :md="8" :lg="6" :xl="6">
        <div class="action-card">
          <div class="action-icon">📊</div>
          <div class="action-title">数据统计</div>
          <div class="action-desc">查看农场运营数据</div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="12" :md="8" :lg="6" :xl="6">
        <div class="action-card">
          <div class="action-icon">⚙️</div>
          <div class="action-title">系统设置</div>
          <div class="action-desc">配置系统参数</div>
        </div>
      </el-col>
    </el-row>

    <div class="decor soil"></div>
    <div class="decor grass"></div>
  </div>
</template>

<script>
export default {
  name: 'Index',
  data() {
    return {
      version: '3.9.0',
      // 生日故事相关
      storyEventSource: null,
      storyStatus: 'idle', // idle | running | done
      storyChunks: [],
      // 每日任务相关
      missionEventSource: null,
      missionStatus: 'idle', // idle | running | done
      missionChunks: []
    }
  },
  computed: {
    storyStatusText() {
      if (this.storyStatus === 'idle') return '待开始'
      if (this.storyStatus === 'running') return '生成中'
      if (this.storyStatus === 'done') return '已完成'
      return ''
    },
    storyDisplayHtml() {
      return this.storyChunks.join('') || '<span class="placeholder">等待生日故事...</span>'
    },
    missionStatusText() {
      if (this.missionStatus === 'idle') return '待开始'
      if (this.missionStatus === 'running') return '获取中'
      if (this.missionStatus === 'done') return '已完成'
      return ''
    },
    missionDisplayHtml() {
      return this.missionChunks.join('') || '<span class="placeholder">等待每日任务...</span>'
    }
  },
  methods: {
    // 启动生日故事SSE
    startStorySSE() {
      if (this.storyEventSource) this.storyEventSource.close()
      this.storyChunks = []
      this.storyStatus = 'running'
      try {
        this.storyEventSource = new EventSource('http://localhost:8099/api/agent/getBirthInfoStream')
      } catch (e) {
        this.storyStatus = 'idle'
        this.storyChunks.push('<div class="err">无法建立连接: '+ (e.message||'') +'</div>')
        return
      }
      this.storyEventSource.onopen = () => {
        // 连接建立，但不显示系统提示
      }
      this.storyEventSource.onmessage = (e) => {
        const data = (e.data || '').trim()
        if (!data || data === '{}') return
        if (data === '[DONE]') {
          this.stopStorySSE(true)
          return
        }
        const clean = data.replace(/\s+/g, ' ')
        this.storyChunks.push('<span>'+ this.escapeHtml(clean) +'</span>')
        this.$nextTick(() => this.scrollBottom('storyOutputBox'))
      }
      this.storyEventSource.onerror = () => {
        this.stopStorySSE(true)
      }
    },
    // 停止生日故事SSE
    stopStorySSE(finished = false) {
      if (this.storyEventSource) {
        this.storyEventSource.close()
        this.storyEventSource = null
      }
      this.storyStatus = finished ? 'done' : 'idle'
    },
    // 启动每日任务SSE
    startMissionSSE() {
      if (this.missionEventSource) this.missionEventSource.close()
      this.missionChunks = []
      this.missionStatus = 'running'
      try {
        this.missionEventSource = new EventSource('http://localhost:8099/api/agent/getMissionInfoStream')
      } catch (e) {
        this.missionStatus = 'idle'
        this.missionChunks.push('<div class="err">无法建立连接: '+ (e.message||'') +'</div>')
        return
      }
      this.missionEventSource.onopen = () => {
        // 连接建立，但不显示系统提示
      }
      this.missionEventSource.onmessage = (e) => {
        const data = (e.data || '').trim()
        if (!data || data === '{}') return
        if (data === '[DONE]') {
          this.stopMissionSSE(true)
          return
        }
        const clean = data.replace(/\s+/g, ' ')
        this.missionChunks.push('<span>'+ this.escapeHtml(clean) +'</span>')
        this.$nextTick(() => this.scrollBottom('missionOutputBox'))
      }
      this.missionEventSource.onerror = () => {
        this.stopMissionSSE(true)
      }
    },
    // 停止每日任务SSE
    stopMissionSSE(finished = false) {
      if (this.missionEventSource) {
        this.missionEventSource.close()
        this.missionEventSource = null
      }
      this.missionStatus = finished ? 'done' : 'idle'
    },
    // 滚动到底部
    scrollBottom(refName) {
      const box = this.$refs[refName]
      if (box) box.scrollTop = box.scrollHeight
    },
    // HTML转义
    escapeHtml(str) {
      return str
    }
  },
  mounted() {
    // 页面加载完成后自动开始获取数据
    this.startStorySSE()
    // 延迟1秒启动任务SSE，避免同时请求
    setTimeout(() => {
      this.startMissionSSE()
    }, 1000)
  },
  beforeDestroy() {
    if (this.storyEventSource) this.storyEventSource.close()
    if (this.missionEventSource) this.missionEventSource.close()
  }
}
</script>

<style scoped lang="scss">
.stardew-home {
  position: relative;
  min-height: calc(100vh - 60px);
  padding: 40px 40px 120px;
  background: linear-gradient(#1e2c44 0%, #314d61 45%, #3f5d6b 60%, #4a6a64 75%, #5c7a4b 100%);
  font-family: 'Pixel Arial', 'Microsoft YaHei', sans-serif;
  overflow: hidden;
  color: #27323a;

  /* 星星 */
  &:before, &:after {
    content: '';
    position: absolute;
    left: 0; top: 0; right: 0; bottom: 0;
    background-image:
      radial-gradient(2px 2px at 10% 20%, rgba(255,255,255,0.9), transparent 60%),
      radial-gradient(2px 2px at 30% 70%, rgba(255,255,255,0.7), transparent 60%),
      radial-gradient(2px 2px at 80% 40%, rgba(255,255,255,0.8), transparent 60%),
      radial-gradient(2px 2px at 55% 15%, rgba(255,255,255,0.6), transparent 60%),
      radial-gradient(2px 2px at 70% 85%, rgba(255,255,255,0.5), transparent 60%);
    animation: twinkle 8s linear infinite;
    pointer-events: none;
  }
  &:after { animation-direction: reverse; opacity: .5; }

  .sv-header {
    text-align: center;
    margin-bottom: 28px;
    .title {
      margin: 0 0 8px;
      font-size: 32px;
      font-weight: 600;
      letter-spacing: 2px;
      color: #f2e5b5;
      text-shadow: 0 2px 0 #523711, 0 3px 4px rgba(0,0,0,.4);
    }
    .subtitle {
      margin: 0;
      color: #f8f4e6;
      font-size: 14px;
      opacity: .85;
      text-shadow: 0 1px 2px rgba(0,0,0,.4);
    }
  }

  .main-content {
    margin-bottom: 32px;
    z-index: 2;
    position: relative;
  }

  .panel {
    background: #fff8e6;
    border: 3px solid #b88646;
    box-shadow: 0 4px 0 #935d1d, 0 8px 16px rgba(0,0,0,.3);
    border-radius: 12px;
    padding: 20px 22px 28px;
    position: relative;
    height: 100%;

    .panel-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 14px;
      .panel-title {
        font-size: 16px;
        font-weight: 600;
        color: #5a3d14;
      }
      .status {
        font-size: 12px;
        padding: 4px 10px;
        border-radius: 20px;
        background: #d1d5db;
        &.running { background: #ffd04d; }
        &.done { background: #67c23a; color: #fff; }
        &.idle { background: #d8c9aa; }
      }
    }

    .output {
      height: 280px;
      overflow-y: auto;
      background: #fcf9f1;
      border: 2px dashed #c6a86a;
      border-radius: 8px;
      padding: 14px 16px;
      font-size: 13px;
      line-height: 1.6;
      color: #4b3a1e;
      box-shadow: inset 0 2px 4px rgba(0,0,0,.08);
      scrollbar-width: thin;
      scrollbar-color: #c6a86a #fcf9f1;
      
      .placeholder { color: #b9a07a; font-style: italic; }
      .err { color: #d9534f; margin: 6px 0; font-weight: 600; }
    }
  }

  .quick-actions {
    z-index: 2;
    position: relative;
    
    .action-card {
      background: #fff8e6;
      border: 3px solid #b88646;
      border-radius: 12px;
      padding: 24px 20px;
      text-align: center;
      cursor: pointer;
      transition: all 0.3s ease;
      box-shadow: 0 4px 0 #935d1d, 0 6px 12px rgba(0,0,0,.2);
      margin-bottom: 20px;
      
      &:hover {
        transform: translateY(-3px);
        box-shadow: 0 7px 0 #935d1d, 0 10px 20px rgba(0,0,0,.3);
      }
      
      .action-icon {
        font-size: 48px;
        margin-bottom: 12px;
        filter: drop-shadow(0 2px 4px rgba(0,0,0,0.2));
      }
      
      .action-title {
        font-size: 16px;
        font-weight: 600;
        color: #5a3d14;
        margin-bottom: 8px;
        text-shadow: 0 1px 2px rgba(0,0,0,0.1);
      }
      
      .action-desc {
        font-size: 12px;
        color: #6b5b3d;
        opacity: 0.8;
      }
    }
  }

  .decor.soil {
    position: absolute; left: 0; right:0; bottom: 60px; height: 40px;
    background: repeating-linear-gradient(45deg,#6e4a28 0 14px,#5c3c1d 14px 28px);
    filter: brightness(.9);
  }
  .decor.grass {
    position: absolute; left:0; right:0; bottom:0; height: 70px;
    background: linear-gradient(#7fbf3a,#5e9e27);
    box-shadow: inset 0 4px 0 #4b811d; border-top: 3px solid #35530f;
  }
}

@keyframes twinkle {
  0%,100% { opacity: .7; }
  50% { opacity: .3; }
}

@media (max-width: 992px) {
  .stardew-home {
    .main-content .el-col:first-child {
      margin-bottom: 20px;
    }
  }
}

@media (max-width: 780px) {
  .stardew-home { 
    padding: 28px 16px 120px; 
    
    .panel { 
      padding: 18px 16px 26px !important; 
      
      .output { 
        height: 200px !important; 
      }
    }
    
    .sv-header .title { 
      font-size: 26px !important; 
    }
    
    .quick-actions {
      .action-card {
        padding: 16px;
        margin-bottom: 16px;
        
        .action-icon {
          font-size: 36px;
          margin-bottom: 8px;
        }
        
        .action-title {
          font-size: 14px;
        }
        
        .action-desc {
          font-size: 11px;
        }
      }
    }
  }
}
</style>

