<template>
  <div class="stardew-home">
    <div class="sv-header">
      <h1 class="title">🌾 星露谷今日生日人物故事</h1>
      <p class="subtitle">Stardew Sage</p>
    </div>

    <div class="panel">
      <div class="panel-header">
        <span class="panel-title">故事输出</span>
        <span class="status" :class="status">{{ statusText }}</span>
      </div>
      <div class="output" ref="outputBox" v-html="displayHtml"></div>
      <!-- 移除按钮，自动开始 -->
    </div>


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
      eventSource: null,
      status: 'idle', // idle | running | done
      chunks: []
    }
  },
  computed: {
    statusText() {
      if (this.status === 'idle') return '待开始'
      if (this.status === 'running') return '接收中'
      if (this.status === 'done') return '已完成'
      return ''
    },
    displayHtml() {
      return this.chunks.join('') || '<span class="placeholder">等待数据...</span>'
    }
  },
  methods: {
    startSSE() {
      if (this.eventSource) this.eventSource.close()
      this.chunks = []
      this.status = 'running'
      try {
        // 修改为远端 MCPServer 完整地址
        this.eventSource = new EventSource('http://localhost:8099/api/agent/getDressingAdvice')
      } catch (e) {
        this.status = 'idle'
        this.chunks.push('<div class="err">无法建立连接: '+ (e.message||'') +'</div>')
        return
      }
      this.eventSource.onopen = () => {
        // 连接建立，但不显示系统提示
      }
      this.eventSource.onmessage = (e) => {
        const data = (e.data || '').trim()
        if (!data || data === '{}') return
        if (data === '[DONE]') {
          // 完成时不显示系统提示，直接结束
          this.stopSSE(true)
          return
        }
        const clean = data.replace(/\s+/g, ' ')
        // 直接输出数据，不再添加额外的换行
        this.chunks.push('<span>'+ this.escapeHtml(clean) +'</span>')
        this.$nextTick(this.scrollBottom)
      }
      this.eventSource.onerror = () => {
        // 连接异常时不显示错误信息，静默结束
        this.stopSSE(true)
      }
    },
    stopSSE(finished = false) {
      if (this.eventSource) {
        this.eventSource.close()
        this.eventSource = null
      }
      this.status = finished ? 'done' : 'idle'
    },
    clearOutput() {
      this.chunks = []
      this.status = 'idle'
    },
    scrollBottom() {
      const box = this.$refs.outputBox
      if (box) box.scrollTop = box.scrollHeight
    },
    escapeHtml(str) {
      return str
    }
  },
  mounted() {
    // 页面加载完成后自动开始获取数据
    this.startSSE()
  },
  beforeDestroy() {
    if (this.eventSource) this.eventSource.close()
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

  .panel {
    max-width: 980px;
    margin: 0 auto;
    background: #fff8e6;
    border: 3px solid #b88646;
    box-shadow: 0 4px 0 #935d1d, 0 8px 16px rgba(0,0,0,.3);
    border-radius: 12px;
    padding: 20px 22px 28px;
    position: relative;

    .panel-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 14px;
      .panel-title {
        font-size: 18px;
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
      height: 340px;
      overflow-y: auto;
      background: #fcf9f1;
      border: 2px dashed #c6a86a;
      border-radius: 8px;
      padding: 14px 16px;
      font-size: 14px;
      line-height: 1.7;
      color: #4b3a1e;
      box-shadow: inset 0 2px 4px rgba(0,0,0,.08);
      scrollbar-width: thin;
      scrollbar-color: #c6a86a #fcf9f1;
      .placeholder { color: #b9a07a; }
      .sys-line { color: #7d5b25; margin-bottom: 6px; display: block; }
      .err { color: #d9534f; margin: 6px 0; font-weight: 600; }
      .done { color: #3c8a2e; margin-top: 10px; font-weight: 600; }
    }

    .actions {
      margin-top: 18px;
      display: flex;
      gap: 12px;
    }
  }

  .tips {
    max-width: 980px;
    margin: 14px auto 0;
    font-size: 12px;
    color: #e9e2d3;
    text-shadow: 0 1px 2px rgba(0,0,0,.4);
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

@media (max-width: 780px) {
  .stardew-home { padding: 28px 16px 120px; }
  .panel { padding: 18px 16px 26px !important; }
  .panel .output { height: 300px !important; }
  .sv-header .title { font-size: 26px !important; }
}
</style>

