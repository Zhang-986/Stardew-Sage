<template>
  <div class="recommendation-container">
    <div class="page-header">
      <h1>🎯 智能推荐系统</h1>
      <p class="subtitle">基于 AI 和向量相似度的个性化推荐</p>
    </div>

    <!-- 查询输入区 -->
    <el-card class="query-card" shadow="hover">
      <el-tabs v-model="activeTab" type="card">
        <el-tab-pane label="个性化推荐" name="personalized">
          <div class="input-group">
            <el-input
              v-model="personalizedQuery"
              placeholder="输入你想了解的内容，例如：最赚钱的作物"
              clearable
              @keyup.enter.native="getPersonalizedRecommendations"
            >
              <el-button slot="append" icon="el-icon-search" @click="getPersonalizedRecommendations">
                获取推荐
              </el-button>
            </el-input>
            <div class="tips">
              💡 提示：输入任何游戏相关的问题，AI 会为你推荐相关内容
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="策略建议" name="strategy">
          <div class="input-group">
            <el-input
              v-model="strategyScenario"
              type="textarea"
              :rows="3"
              placeholder="描述你的游戏场景，例如：我是春季第一周的新手，只有500金币，想要快速赚钱"
              clearable
            />
            <el-button
              type="primary"
              icon="el-icon-magic-stick"
              @click="getStrategyRecommendations"
              :loading="strategyLoading"
            >
              获取策略建议
            </el-button>
            <div class="tips">
              🎮 描述你的游戏状态，获取专业的策略分析和建议
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="上下文推荐" name="contextual">
          <div class="input-group">
            <el-input
              v-model="contextualContext"
              placeholder="上下文（可选）"
              clearable
            />
            <el-input
              v-model="contextualQuery"
              placeholder="你的问题"
              clearable
              @keyup.enter.native="getContextualRecommendations"
            >
              <el-button slot="append" icon="el-icon-chat-line-round" @click="getContextualRecommendations">
                智能推荐
              </el-button>
            </el-input>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 推荐结果展示区 -->
    <el-row :gutter="20" v-if="recommendations.length > 0">
      <el-col
        :xs="24"
        :sm="12"
        :md="8"
        :lg="6"
        v-for="(rec, index) in recommendations"
        :key="index"
      >
        <el-card class="recommendation-card" shadow="hover">
          <div class="card-header">
            <span class="rank-badge">{{ index + 1 }}</span>
            <el-tag :type="getTypeColor(rec.itemType)" size="mini">
              {{ rec.itemType }}
            </el-tag>
          </div>
          <div class="card-content">
            <h3>{{ rec.itemName }}</h3>
            <div class="score-bar">
              <el-progress
                :percentage="Math.round(rec.score * 100)"
                :color="getScoreColor(rec.score)"
              />
            </div>
            <p class="reason">{{ rec.reason }}</p>
            <el-tag size="mini" effect="plain">{{ rec.source }}</el-tag>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 流式推荐文本展示区 -->
    <el-card v-if="streamingText" class="streaming-card" shadow="hover">
      <div class="streaming-header">
        <h3>
          <i class="el-icon-loading" v-if="isStreaming"></i>
          <i class="el-icon-success" v-else></i>
          {{ isStreaming ? '正在生成推荐...' : '推荐完成' }}
        </h3>
      </div>
      <div class="streaming-content markdown-body" v-html="formatMarkdown(streamingText)"></div>
    </el-card>

    <!-- 空状态 -->
    <el-empty
      v-if="!loading && recommendations.length === 0 && !streamingText"
      description="暂无推荐内容，请输入查询获取推荐"
    >
      <el-button type="primary" @click="showExampleQueries">查看示例查询</el-button>
    </el-empty>

    <!-- 示例查询对话框 -->
    <el-dialog title="示例查询" :visible.sync="exampleDialogVisible" width="50%">
      <el-card shadow="never" v-for="(example, index) in exampleQueries" :key="index" class="example-card">
        <div class="example-content" @click="useExample(example)">
          <div class="example-label">{{ example.label }}</div>
          <div class="example-query">{{ example.query }}</div>
        </div>
      </el-card>
    </el-dialog>
  </div>
</template>

<script>
import axios from "axios";

export default {
  name: 'RecommendationIndex',
  data() {
    return {
      activeTab: 'personalized',
      personalizedQuery: '',
      strategyScenario: '',
      contextualContext: '',
      contextualQuery: '',
      recommendations: [],
      streamingText: '',
      loading: false,
      strategyLoading: false,
      isStreaming: false,
      exampleDialogVisible: false,
      exampleQueries: [
        { label: '作物推荐', query: '最赚钱的春季作物', type: 'personalized' },
        { label: '人物关系', query: 'Abigail 喜欢什么礼物', type: 'personalized' },
        { label: '策略优化', query: '我是新手，有1000金币，在春季第5天，应该如何规划？', type: 'strategy' },
        { label: '任务攻略', query: '如何快速完成社区中心', type: 'contextual' },
        { label: '经济分析', query: '养鸡和种地哪个更赚钱', type: 'strategy' }
      ]
    }
  },
  methods: {
    async getPersonalizedRecommendations() {
      if (!this.personalizedQuery.trim()) {
        this.$message.warning('请输入查询内容');
        return;
      }

      this.loading = true;
      this.recommendations = [];

      try {
        const response = await axios.get('http://localhost:8099/api/recommendation/personalized', {
          params: {
            query: this.personalizedQuery,
            topK: 8
          }
        });

        this.recommendations = response.data;

        if (this.recommendations.length === 0) {
          this.$message.info('未找到相关推荐，请尝试其他关键词');
        }
      } catch (error) {
        console.error('获取推荐失败:', error);
        this.$message.error('获取推荐失败，请稍后重试');
      } finally {
        this.loading = false;
      }
    },

    async getStrategyRecommendations() {
      if (!this.strategyScenario.trim()) {
        this.$message.warning('请输入场景描述');
        return;
      }

      this.strategyLoading = true;
      this.isStreaming = true;
      this.streamingText = '';

      try {
        const baseUrl =  'http://localhost:8099';
        const url = `${baseUrl}/api/recommendation/strategy?scenario=${encodeURIComponent(this.strategyScenario)}`;

        const eventSource = new EventSource(url);

        eventSource.onmessage = (event) => {
          this.streamingText += event.data;
        };

        eventSource.onerror = (error) => {
          console.error('SSE error:', error);
          eventSource.close();
          this.isStreaming = false;
          this.strategyLoading = false;

          if (!this.streamingText) {
            this.$message.error('获取策略建议失败');
          }
        };

        eventSource.addEventListener('end', () => {
          eventSource.close();
          this.isStreaming = false;
          this.strategyLoading = false;
        });

        // 超时处理
        setTimeout(() => {
          if (this.isStreaming) {
            eventSource.close();
            this.isStreaming = false;
            this.strategyLoading = false;
          }
        }, 60000);

      } catch (error) {
        console.error('获取策略建议失败:', error);
        this.$message.error('获取策略建议失败');
        this.isStreaming = false;
        this.strategyLoading = false;
      }
    },

    async getContextualRecommendations() {
      if (!this.contextualQuery.trim()) {
        this.$message.warning('请输入查询内容');
        return;
      }

      this.isStreaming = true;
      this.streamingText = '';

      try {
        const baseUrl =  'http://localhost:8099';
        const url = `${baseUrl}/api/recommendation/contextual?context=${encodeURIComponent(this.contextualContext)}&query=${encodeURIComponent(this.contextualQuery)}`;

        const eventSource = new EventSource(url);

        eventSource.onmessage = (event) => {
          this.streamingText += event.data;
        };

        eventSource.onerror = (error) => {
          console.error('SSE error:', error);
          eventSource.close();
          this.isStreaming = false;
        };

      } catch (error) {
        console.error('获取上下文推荐失败:', error);
        this.$message.error('获取推荐失败');
        this.isStreaming = false;
      }
    },

    getTypeColor(type) {
      const colorMap = {
        'crop': 'success',
        'character': 'primary',
        'recipe': 'warning',
        'general': 'info'
      };
      return colorMap[type] || 'info';
    },

    getScoreColor(score) {
      if (score >= 0.8) return '#67C23A';
      if (score >= 0.6) return '#E6A23C';
      return '#909399';
    },

    formatMarkdown(text) {
      // 简单的 Markdown 格式化
      return text
        .replace(/\n/g, '<br>')
        .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
        .replace(/\*(.*?)\*/g, '<em>$1</em>')
        .replace(/📌/g, '<span class="emoji">📌</span>');
    },

    showExampleQueries() {
      this.exampleDialogVisible = true;
    },

    useExample(example) {
      this.exampleDialogVisible = false;

      if (example.type === 'personalized') {
        this.activeTab = 'personalized';
        this.personalizedQuery = example.query;
        this.getPersonalizedRecommendations();
      } else if (example.type === 'strategy') {
        this.activeTab = 'strategy';
        this.strategyScenario = example.query;
        this.getStrategyRecommendations();
      } else if (example.type === 'contextual') {
        this.activeTab = 'contextual';
        this.contextualQuery = example.query;
        this.getContextualRecommendations();
      }
    }
  }
}
</script>

<style scoped>
.recommendation-container {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  text-align: center;
  margin-bottom: 30px;
}

.page-header h1 {
  font-size: 32px;
  color: #2c3e50;
  margin-bottom: 10px;
}

.subtitle {
  color: #606266;
  font-size: 14px;
}

.query-card {
  margin-bottom: 30px;
}

.input-group {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.tips {
  font-size: 13px;
  color: #909399;
  padding-left: 5px;
}

.recommendation-card {
  margin-bottom: 20px;
  transition: transform 0.3s;
}

.recommendation-card:hover {
  transform: translateY(-5px);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.rank-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 50%;
  font-weight: bold;
  font-size: 14px;
}

.card-content h3 {
  font-size: 16px;
  color: #303133;
  margin-bottom: 10px;
}

.score-bar {
  margin: 15px 0;
}

.reason {
  font-size: 13px;
  color: #606266;
  margin: 10px 0;
  line-height: 1.6;
}

.streaming-card {
  margin-top: 20px;
}

.streaming-header {
  margin-bottom: 15px;
  padding-bottom: 15px;
  border-bottom: 1px solid #ebeef5;
}

.streaming-header h3 {
  font-size: 18px;
  color: #303133;
}

.streaming-content {
  line-height: 1.8;
  color: #606266;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.markdown-body {
  font-size: 14px;
}

.emoji {
  font-size: 18px;
  margin-right: 5px;
}

.example-card {
  margin-bottom: 10px;
  cursor: pointer;
  transition: all 0.3s;
}

.example-card:hover {
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.example-content {
  padding: 5px;
}

.example-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 5px;
}

.example-query {
  font-size: 14px;
  color: #303133;
}

@media (max-width: 768px) {
  .page-header h1 {
    font-size: 24px;
  }

  .recommendation-container {
    padding: 10px;
  }
}
</style>
