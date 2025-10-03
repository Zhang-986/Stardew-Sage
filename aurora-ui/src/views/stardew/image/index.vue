<template>
  <div class="stardew-image-container">
    <el-card class="stardew-card">
      <div slot="header" class="stardew-header clearfix">
        <div class="header-content">
          <i class="stardew-icon">🖼️</i>
          <span class="header-title">Stardew Sage - 图像分析助手</span>
        </div>
      </div>

      <el-row :gutter="20">
        <!-- 左侧：上传和预览 -->
        <el-col :span="10">
          <el-card shadow="never">
            <div slot="header">
              <span>上传图片进行分析</span>
            </div>
            <el-upload
              class="image-uploader"
              action="#"
              :show-file-list="false"
              :http-request="handleUpload"
              :before-upload="beforeUpload"
              drag
            >
              <div v-if="imageUrl">
                <img :src="imageUrl" class="uploaded-image">
                <div class="re-upload-mask">
                  <i class="el-icon-upload"></i>
                  <span>点击或拖拽重新上传</span>
                </div>
              </div>
              <div v-else>
                <i class="el-icon-upload"></i>
                <div class="el-upload__text">将图片拖到此处，或<em>点击上传</em></div>
                <div class="el-upload__tip" slot="tip">只能上传jpg/png文件，且不超过5MB</div>
              </div>
            </el-upload>
            
            <el-input
              v-model="question"
              type="textarea"
              :rows="3"
              placeholder="请输入您想问的关于图片的问题（选填）"
              style="margin-top: 20px;"
            ></el-input>

            <el-button
              type="primary"
              @click="startAnalysis"
              :loading="loading"
              :disabled="!file"
              style="width: 100%; margin-top: 20px;"
            >
              <i class="el-icon-magic-stick"></i>
              开始分析
            </el-button>
          </el-card>
        </el-col>

        <!-- 右侧：分析结果 -->
        <el-col :span="14">
          <el-card shadow="never" class="result-card">
            <div slot="header">
              <span>分析结果</span>
            </div>
            <div class="result-content">
              <div v-if="loading" class="loading-indicator">
                <i class="el-icon-loading"></i>
                <span>正在分析中，请稍候...</span>
              </div>
              <div v-else-if="!analysisResult && !error" class="empty-state">
                <i class="el-icon-chat-dot-square"></i>
                <p>分析结果将在这里显示</p>
              </div>
              <div v-else-if="error" class="error-state">
                <i class="el-icon-error"></i>
                <p>分析失败</p>
                <p class="error-message">{{ error }}</p>
              </div>
              <div v-else class="analysis-text" v-html="analysisResult"></div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script>
import { analyzeImage } from '@/api/ai'

export default {
  name: 'StardewImage',
  data() {
    return {
      file: null,
      imageUrl: '',
      question: '',
      loading: false,
      analysisResult: '',
      error: ''
    }
  },
  methods: {
    beforeUpload(file) {
      const isImage = file.type.startsWith('image/')
      const isLt5M = file.size / 1024 / 1024 < 5

      if (!isImage) {
        this.$message.error('只能上传图片格式!')
        return false
      }
      if (!isLt5M) {
        this.$message.error('上传图片大小不能超过 5MB!')
        return false
      }
      return true
    },
    handleUpload(options) {
      this.file = options.file
      this.imageUrl = URL.createObjectURL(this.file)
      this.analysisResult = '' // 清空旧结果
      this.error = ''
    },
    async startAnalysis() {
      if (!this.file) {
        this.$message.warning('请先上传图片')
        return
      }

      this.loading = true
      this.analysisResult = ''
      this.error = ''

      try {
        const response = await analyzeImage(this.file, this.question)

        if (!response.ok) {
          throw new Error(`服务错误: ${response.status} ${response.statusText}`)
        }

        const reader = response.body.getReader()
        const decoder = new TextDecoder()

        const processStream = async () => {
          while (true) {
            const { done, value } = await reader.read()
            if (done) {
              break
            }
            const chunk = decoder.decode(value, { stream: true })
            
            if (chunk.includes('[ERROR]')) {
              this.error = chunk.replace('data: [ERROR]', '').trim()
              break
            }
            
            this.analysisResult += chunk
          }
          this.loading = false
        }

        processStream()

      } catch (err) {
        this.error = err.message || '请求失败，请检查网络或联系管理员'
        this.loading = false
        console.error('分析失败:', err)
      }
    }
  }
}
</script>

<style scoped>
.stardew-image-container {
  padding: 20px;
  background: linear-gradient(135deg, #8B4513, #A0522D);
  min-height: calc(100vh - 84px);
}

.stardew-card {
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.stardew-header {
  background: linear-gradient(90deg, #8B4513, #A0522D);
  color: #fff;
  padding: 15px 20px;
  border-radius: 8px 8px 0 0;
  margin: -20px -20px 20px -20px;
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

.image-uploader .el-upload {
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  width: 100%;
  height: 250px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
}
.image-uploader .el-upload:hover {
  border-color: #409EFF;
}
.el-upload-dragger {
  width: 100%;
  height: 100%;
}
.uploaded-image {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
.re-upload-mask {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0,0,0,0.5);
  color: white;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.3s;
}
.image-uploader:hover .re-upload-mask {
  opacity: 1;
}

.result-card {
  height: 100%;
}
.result-content {
  height: 450px;
  overflow-y: auto;
  padding: 10px;
  border: 1px solid #eee;
  border-radius: 4px;
  background: #fdfdfd;
}
.loading-indicator, .empty-state, .error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #909399;
}
.loading-indicator i, .empty-state i, .error-state i {
  font-size: 48px;
  margin-bottom: 20px;
}
.error-state {
  color: #f56c6c;
}
.error-message {
  font-size: 12px;
  color: #999;
}
.analysis-text {
  white-space: pre-wrap;
  line-height: 1.8;
  font-size: 14px;
}
</style>
