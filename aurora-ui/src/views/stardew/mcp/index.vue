<template>
  <div class="stardew-mcp-container">
    <div class="stardew-card-wrapper">
      <div class="stardew-header clearfix">
        <div class="header-content">
          <i class="stardew-icon">🤖</i>
          <span class="header-title">Stardew Sage - AI百科助手</span>
        </div>
        <el-button 
          style="float: right; padding: 3px 0" 
          type="text"
          @click="clearChat"
          class="refresh-btn"
        >
          清空对话
        </el-button>
      </div>

      <!-- 模式选择区域 -->
      <div class="mode-selector">
        <el-card shadow="never" class="mode-card">
          <div slot="header">
            <span><i class="el-icon-setting"></i> 选择助手模式</span>
          </div>
          <el-radio-group v-model="selectedMode" @change="onModeChange">
            <el-radio-button label="mcp">
              <i class="el-icon-connection"></i>
              MCP工具调用
            </el-radio-button>
            <el-radio-button label="rag">
              <i class="el-icon-document"></i>
              RAG向量数据库
            </el-radio-button>
          </el-radio-group>
          <div class="mode-description">
            <el-alert
              :title="modeDescription"
              :type="selectedMode === 'mcp' ? 'info' : 'success'"
              :closable="false"
              show-icon
            ></el-alert>
          </div>
        </el-card>
      </div>

      <!-- 聊天消息区域 -->
      <div class="chat-container" ref="chatContainer">
        <div class="message-list">
          
          <div 
            v-for="(message, index) in messageList" 
            :key="'msg-' + index + '-' + message.timestamp"
            class="message-item"
            :class="message.type"
          >
            <div class="message-avatar">
              <i v-if="message.type === 'user'" class="el-icon-user-solid"></i>
              <i v-else class="el-icon-chat-dot-round"></i>
            </div>
            <div class="message-content">
              <div class="message-text" v-html="formatMessage(message.content)"></div>
              <div class="message-time">
                {{ message.time }}
                <el-tag v-if="message.mode" size="mini" :type="message.mode === 'mcp' ? 'info' : 'success'">
                  {{ message.mode === 'mcp' ? 'MCP' : 'RAG' }}
                </el-tag>
              </div>
            </div>
          </div>
          
          <!-- 正在输入指示器 -->
          <div v-if="isTyping" class="message-item assistant typing">
            <div class="message-avatar">
              <i class="el-icon-chat-dot-round"></i>
            </div>
            <div class="message-content">
              <div class="typing-indicator">
                <span></span>
                <span></span>
                <span></span>
              </div>
              <div class="typing-mode">
                <el-tag size="mini" :type="selectedMode === 'mcp' ? 'info' : 'success'">
                  {{ selectedMode === 'mcp' ? 'MCP模式处理中...' : 'RAG模式处理中...' }}
                </el-tag>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 输入区域 -->
      <div class="input-container">
        <el-row :gutter="10">
          <el-col :span="20">
            <el-input
              v-model="inputQuestion"
              :placeholder="inputPlaceholder"
              @keyup.enter.native="sendMessage"
              :disabled="isLoading"
              type="textarea"
              :rows="2"
              resize="none"
            ></el-input>
          </el-col>
          <el-col :span="4">
            <el-button 
              type="primary" 
              @click="sendMessage"
              :loading="isLoading"
              :disabled="!inputQuestion.trim()"
              style="height: 60px; width: 100%;"
            >
              <i class="el-icon-s-promotion"></i>
              发送
            </el-button>
          </el-col>
        </el-row>
      </div>
    </div>
  </div>
</template>

<script>
import { getInfoDetailSSE, getInfoRagDetailSSE } from '@/api/ai'
import Vue from 'vue'

export default {
  name: 'StardewMcp',
  data() {
    return {
      messageList: [], // 改用messageList避免与computed冲突
      inputQuestion: '',
      isLoading: false,
      isTyping: false,
      currentEventSource: null,
      selectedMode: 'mcp'
    }
  },
  computed: {
    modeDescription() {
      return this.selectedMode === 'mcp' 
        ? 'MCP工具调用模式：使用工具链接和实时数据库查询，提供准确的星露谷物语信息'
        : 'RAG向量数据库模式：基于预处理的向量数据库，快速检索相关星露谷物语知识'
    },
    inputPlaceholder() {
      return this.selectedMode === 'mcp'
        ? '请输入您的问题（MCP工具调用模式）...'
        : '请输入您的问题（RAG向量数据库模式）...'
    }
  },
  methods: {
    onModeChange(newMode) {
      this.$message.info(`已切换到${newMode === 'mcp' ? 'MCP工具调用' : 'RAG向量数据库'}模式`)
    },
    
    sendMessage() {
      if (!this.inputQuestion.trim() || this.isLoading) return
      
      const question = this.inputQuestion.trim()
      this.inputQuestion = ''
      
      console.log('🚀 开始发送消息:', question)
      
      // 添加用户消息
      const userMsg = {
        type: 'user',
        content: question,
        time: this.formatTime(new Date()),
        mode: this.selectedMode,
        timestamp: Date.now()
      }
      this.messageList.push(userMsg)
      console.log('✅ 用户消息已添加，当前消息数:', this.messageList.length)
      
      // 添加助手消息占位
      const assistantMsg = {
        type: 'assistant',
        content: '',
        time: this.formatTime(new Date()),
        mode: this.selectedMode,
        timestamp: Date.now()
      }
      this.messageList.push(assistantMsg)
      const assistantIndex = this.messageList.length - 1
      console.log('✅ 助手消息占位已添加，索引:', assistantIndex)
      
      // 开始加载状态
      this.isLoading = true
      this.isTyping = true
      
      this.$nextTick(() => {
        this.scrollToBottom()
      })
      
      // 根据模式创建不同的SSE连接
      this.currentEventSource = this.selectedMode === 'mcp' 
        ? getInfoDetailSSE(question)
        : getInfoRagDetailSSE(question)
      
      let assistantMessage = ''
      
      this.currentEventSource.onopen = () => {
        console.log(`✅ SSE连接已建立 - ${this.selectedMode}模式`)
      }
      
      this.currentEventSource.onmessage = (event) => {
        const data = event.data
        console.log('📥 SSE数据:', data.substring(0, 50))
        
        // 处理结束信号
        if (data.includes('[DONE]')) {
          console.log('🏁 流结束，最终长度:', assistantMessage.length)
          this.finishAssistantMessage(assistantMessage, assistantIndex)
          return
        }
        
        // 处理data:前缀
        let actualData = data.startsWith('data: ') ? data.substring(6) : data
        
        // 忽略空数据
        if (!actualData || actualData.trim() === '') return
        
        // 累积消息
        assistantMessage += actualData
        
        // 使用Vue.set强制更新
        Vue.set(this.messageList[assistantIndex], 'content', assistantMessage)
        console.log(`📝 更新消息 [${assistantIndex}]，长度:`, assistantMessage.length)
        
        this.$nextTick(() => {
          this.scrollToBottom()
        })
      }
      
      this.currentEventSource.onerror = (error) => {
        console.error('❌ SSE错误:', error)
        this.finishAssistantMessage(assistantMessage || '抱歉，服务暂时不可用。', assistantIndex)
      }
    },
    
    finishAssistantMessage(finalContent, index) {
      console.log('🔚 完成，索引:', index, '长度:', finalContent.length)
      
      this.isLoading = false
      this.isTyping = false
      
      if (this.currentEventSource) {
        this.currentEventSource.close()
        this.currentEventSource = null
      }
      
      // 最终更新
      if (this.messageList[index]) {
        Vue.set(this.messageList[index], 'content', finalContent || '无响应内容')
      }
      
      this.$nextTick(() => {
        this.scrollToBottom()
      })
    },
    
    formatTime(date) {
      return date.toLocaleTimeString('zh-CN', {
        hour: '2-digit',
        minute: '2-digit'
      })
    },
    
    scrollToBottom() {
      const messageListEl = this.$refs.chatContainer // 改为滚动父容器
      if (messageListEl) {
        messageListEl.scrollTop = messageListEl.scrollHeight
      }
    },
    
    clearChat() {
      this.messageList = []
      if (this.currentEventSource) {
        this.currentEventSource.close()
        this.currentEventSource = null
      }
      this.isLoading = false
      this.isTyping = false
      this.$message.success('对话已清空')
    },
    
    formatMessage(content) {
      if (!content) return ''
      
      let formatted = content
      
      // 移除特殊标记
      formatted = formatted.replace(/<\|begin_of_box\|>/g, '')
      formatted = formatted.replace(/<\|end_of_box\|>/g, '')
      formatted = formatted.replace(/<begin_of_box>/g, '')
      formatted = formatted.replace(/<end_of_box>/g, '')
      
      // 先处理粗体
      formatted = formatted.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
      
      // 处理列表和段落
      const lines = formatted.split('\n')
      const processedLines = []
      let inList = false
      
      for (let i = 0; i < lines.length; i++) {
        let line = lines[i].trim()
        
        // 检测标题
        if (line.startsWith('###')) {
          processedLines.push('</ul>')
          inList = false
          processedLines.push(`<h3>${line.substring(3).trim()}</h3>`)
        } else if (line.startsWith('##')) {
          if (inList) {
            processedLines.push('</ul>')
            inList = false
          }
          processedLines.push(`<h2>${line.substring(2).trim()}</h2>`)
        } else if (line.startsWith('#')) {
          if (inList) {
            processedLines.push('</ul>')
            inList = false
          }
          processedLines.push(`<h1>${line.substring(1).trim()}</h1>`)
        }
        // 检测列表项
        else if (line.match(/^[-•]\s+(.+)$/) || line.match(/^\d+\.\s+(.+)$/)) {
          if (!inList) {
            processedLines.push('<ul class="formatted-list">')
            inList = true
          }
          const listContent = line.replace(/^[-•]\s+/, '').replace(/^\d+\.\s+/, '')
          processedLines.push(`<li>${listContent}</li>`)
        } else {
          // 非列表项
          if (inList) {
            processedLines.push('</ul>')
            inList = false
          }
          if (line) {
            processedLines.push(`<p>${line}</p>`)
          }
        }
      }
      
      // 如果最后还在列表中，需要闭合
      if (inList) {
        processedLines.push('</ul>')
      }
      
      formatted = processedLines.join('')
      
      // 处理行内代码
      formatted = formatted.replace(/`([^`]+)`/g, '<code class="inline-code">$1</code>')
      
      // 清理多余的空段落
      formatted = formatted.replace(/<p>\s*<\/p>/g, '')
      formatted = formatted.replace(/<p><\/p>/g, '')
      
      return formatted
    }
  },
  
  beforeDestroy() {
    // 组件销毁前关闭SSE连接
    if (this.currentEventSource) {
      this.currentEventSource.close()
    }
  }
}
</script>

<style scoped>
.stardew-mcp-container {
  padding: 20px;
  background: linear-gradient(135deg, #8B4513, #A0522D);
  height: calc(100vh - 84px);
  box-sizing: border-box;
}

.stardew-card-wrapper {
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background-color: #fff;
}

.stardew-header {
  background: linear-gradient(90deg, #8B4513, #A0522D);
  color: #fff;
  padding: 10px 20px;
  border-radius: 8px 8px 0 0;
  flex-shrink: 0;
}

.header-content {
  display: flex;
  align-items: center;
}

.stardew-icon {
  font-size: 24px;
  margin-right: 12px;
}

.header-title {
  font-size: 18px;
  font-weight: 600;
  font-family: 'Georgia', serif;
}

.refresh-btn {
  color: #fff !important;
}

.refresh-btn:hover {
  color: #f0e68c !important;
}

.mode-selector {
  padding: 10px 20px;
  flex-shrink: 0;
  background: #fff;
}

.mode-card {
  border: 1px solid #e6a23c;
  border-radius: 8px;
}

.mode-card .el-card__header {
  padding: 10px 20px !important;
}

.mode-card .el-card__body {
  padding: 15px 20px !important;
}

.mode-description {
  margin-top: 10px;
}

.mode-description .el-alert {
  padding: 8px 16px !important;
}

.mode-description .el-alert__content {
  font-size: 12px !important;
  line-height: 1.4 !important;
}

.chat-container {
  flex: 1;
  overflow-y: auto;
  padding: 15px 20px;
  scroll-behavior: smooth;
}

.message-list {
  /* 移除所有布局属性，让父容器控制滚动 */
}

.message-item {
  display: flex;
  margin-bottom: 20px;
  align-items: flex-start;
  animation: fadeIn 0.3s ease-in;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.message-item.user {
  flex-direction: row-reverse;
}

.message-item.user .message-content {
  background: #409eff;
  color: white;
  margin-right: 10px;
}

.message-item.assistant .message-content {
  background: #f5f7fa;
  color: #303133;
  margin-left: 10px;
}

.message-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #e4e7ed;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  color: #606266;
  flex-shrink: 0;
}

.message-item.user .message-avatar {
  background: #409eff;
  color: white;
}

.message-item.assistant .message-avatar {
  background: #67c23a;
  color: white;
}

.message-content {
  max-width: 70%;
  padding: 12px 16px;
  border-radius: 18px;
  word-wrap: break-word;
  overflow-wrap: break-word;
}

.message-text {
  line-height: 1.8;
  word-wrap: break-word;
  overflow-wrap: break-word;
  font-size: 14px;
}

.message-text p {
  margin: 10px 0;
  line-height: 1.8;
}

.message-text h1, .message-text h2, .message-text h3 {
  margin: 15px 0 10px 0;
  font-weight: bold;
}

.message-text h1 { 
  font-size: 1.3em; 
  color: #409eff;
}
.message-text h2 { 
  font-size: 1.2em; 
  color: #67c23a;
}
.message-text h3 { 
  font-size: 1.1em; 
  color: #e6a23c;
}

.message-text strong {
  font-weight: bold;
  color: #e6a23c;
}

.message-text .formatted-list {
  margin: 12px 0;
  padding-left: 24px;
}

.message-text .formatted-list li {
  margin: 8px 0;
  list-style-type: disc;
  line-height: 1.8;
}

.message-text .inline-code {
  background: #f4f4f5;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
  font-size: 0.9em;
  color: #e96900;
}

.message-time {
  font-size: 12px;
  opacity: 0.7;
  margin-top: 5px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.input-container {
  padding: 15px 20px;
  border-top: 1px solid #ebeef5;
  background: #fafafa;
  flex-shrink: 0;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.1);
}

.typing-indicator {
  display: flex;
  align-items: center;
  gap: 4px;
}

.typing-indicator span {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #409eff;
  animation: typing 1.4s infinite ease-in-out;
}

.typing-indicator span:nth-child(1) {
  animation-delay: -0.32s;
}

.typing-indicator span:nth-child(2) {
  animation-delay: -0.16s;
}

@keyframes typing {
  0%, 80%, 100% {
    transform: scale(0);
    opacity: 0.5;
  }
  40% {
    transform: scale(1);
    opacity: 1;
  }
}

.typing-mode {
  margin-top: 8px;
}

.clearfix:before,
.clearfix:after {
  display: table;
  content: "";
}

.clearfix:after {
  clear: both;
}

/* 聊天消息区域内的滚动条样式 */
.chat-container::-webkit-scrollbar {
  width: 12px;
}

.chat-container::-webkit-scrollbar-track {
  background: #ddd;
  border-radius: 6px;
}

.chat-container::-webkit-scrollbar-thumb {
  background: #666;
  border-radius: 6px;
}

.chat-container::-webkit-scrollbar-thumb:hover {
  background: #444;
}

/* Firefox 浏览器滚动条 */
.chat-container {
  scrollbar-width: auto;
  scrollbar-color: #666 #ddd;
}
</style>