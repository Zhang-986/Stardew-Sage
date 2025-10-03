<template>
  <div class="stardew-show-container">
    <el-card class="stardew-card">
      <div slot="header" class="stardew-header clearfix">
        <div class="header-content">
          <i class="stardew-icon">📊</i>
          <span class="header-title">Stardew Sage - 数据统计展示</span>
        </div>
        <el-button 
          style="float: right; padding: 3px 0" 
          type="text"
          @click="fetchGroupedData"
          class="refresh-btn"
        >
          刷新
        </el-button>
      </div>
      
      <!-- 表格统计查询区域 -->
      <div class="stats-query-section">
        <el-card class="query-card" shadow="never">
          <div slot="header">
            <span><i class="el-icon-search"></i> 表格RAG统计查询</span>
          </div>
          <el-row :gutter="20" type="flex" align="middle">
            <el-col :span="16">
              <el-input
                v-model="queryTableName"
                placeholder="请输入表名（如：stardew_animal）"
                clearable
                @keyup.enter.native="fetchTableStats"
              >
                <template slot="prepend">表名</template>
              </el-input>
            </el-col>
            <el-col :span="8">
              <el-button 
                type="primary" 
                @click="fetchTableStats"
                :loading="statsLoading"
                :disabled="!queryTableName.trim()"
              >
                查询统计
              </el-button>
            </el-col>
          </el-row>
          
          <!-- 统计结果展示 -->
          <div v-if="tableStats" class="stats-result">
            <el-alert
              :title="`RAG向量数据统计 - ${tableStats.tableName}`"
              type="success"
              :closable="false"
              show-icon
              style="margin-top: 15px;"
            >
              <template slot="default">
                <div class="stats-content">
                  <el-row :gutter="20">
                    <el-col :span="8">
                      <div class="stat-item">
                        <span class="stat-label">键数量:</span>
                        <span class="stat-value">{{ tableStats.keyCount }} 条</span>
                      </div>
                    </el-col>
                    <el-col :span="16">
                      <div class="stat-item">
                        <span class="stat-label">匹配模式:</span>
                        <span class="stat-value">{{ tableStats.keysPattern }}</span>
                      </div>
                    </el-col>
                  </el-row>
                  
                  <div v-if="tableStats.sampleData" class="sample-data">
                    <p class="sample-title"><strong>样本数据预览:</strong></p>
                    <el-row :gutter="15">
                      <el-col :span="12" v-if="tableStats.sampleData.name_ch">
                        <div class="sample-item">
                          <span class="sample-label">名称:</span>
                          <el-tag type="primary">{{ tableStats.sampleData.name_ch }}</el-tag>
                        </div>
                      </el-col>
                      <el-col :span="12" v-if="tableStats.sampleData.material">
                        <div class="sample-item">
                          <span class="sample-label">材料:</span>
                          <el-tag type="success">{{ tableStats.sampleData.material }}</el-tag>
                        </div>
                      </el-col>
                      <el-col :span="12" v-if="tableStats.sampleData.price">
                        <div class="sample-item">
                          <span class="sample-label">价格:</span>
                          <el-tag type="warning">{{ tableStats.sampleData.price }}</el-tag>
                        </div>
                      </el-col>
                      <el-col :span="12" v-if="tableStats.sampleData.embedding">
                        <div class="sample-item">
                          <span class="sample-label">向量维度:</span>
                          <el-tag type="info">{{ tableStats.sampleData.embedding.length }}维</el-tag>
                        </div>
                      </el-col>
                    </el-row>
                    
                    <div v-if="tableStats.sampleData.content" class="content-data">
                      <p class="sample-title"><strong>完整内容:</strong></p>
                      <el-input
                        type="textarea"
                        :value="formatContentData(tableStats.sampleData.content)"
                        :rows="3"
                        readonly
                        resize="none"
                      ></el-input>
                    </div>
                  </div>
                  
                  <div v-if="tableStats.error" class="error-section">
                    <el-alert
                      :title="tableStats.error"
                      type="error"
                      :closable="false"
                      show-icon
                    ></el-alert>
                  </div>
                </div>
              </template>
            </el-alert>
          </div>
        </el-card>
      </div>
      
      <div v-loading="loading">
        <el-empty v-if="!loading && groupedData.length === 0" description="暂无统计数据"></el-empty>
        
        <div v-else>
          <el-row :gutter="20">
            <el-col 
              v-for="(item, index) in paginatedData" 
              :key="index"
              :xs="24" 
              :sm="12" 
              :md="8" 
              :lg="6"
              style="margin-bottom: 20px;"
            >
              <el-card class="stat-card" shadow="hover" @click.native="quickQuery(item.key)">
                <div class="stat-content">
                  <div class="stat-title">{{ item.key || '未知分类' }}</div>
                  <div class="stat-count">{{ item.count || 0 }}</div>
                  <div class="stat-label">数据条数</div>
                </div>
              </el-card>
            </el-col>
          </el-row>
          
          <el-pagination
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
            :current-page="currentPage"
            :page-sizes="[8, 16, 24, 32]"
            :page-size="pageSize"
            layout="total, sizes, prev, pager, next, jumper"
            :total="groupedData.length"
            style="margin-top: 20px; text-align: center;"
          >
          </el-pagination>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script>
import { getDefaultKeysGroupedCount, getTableStatsSimple } from '@/api/ai'

export default {
  name: 'StardewShow',
  data() {
    return {
      groupedData: [],
      loading: false,
      currentPage: 1,
      pageSize: 16,
      queryTableName: '',
      tableStats: null,
      statsLoading: false
    }
  },
  computed: {
    paginatedData() {
      const start = (this.currentPage - 1) * this.pageSize
      const end = start + this.pageSize
      return this.groupedData.slice(start, end)
    }
  },
  created() {
    this.fetchGroupedData()
  },
  methods: {
    async fetchGroupedData() {
      this.loading = true
      try {
        const data = await getDefaultKeysGroupedCount()
        // 将对象转换为数组格式
        if (data && typeof data === 'object') {
          this.groupedData = Object.keys(data).map(key => ({
            key: key,
            count: data[key]
          }))
        } else {
          this.groupedData = []
        }
        this.currentPage = 1
        this.$message.success(`统计数据加载成功，共${this.groupedData.length}个分类`)
      } catch (error) {
        this.$message.error('数据加载失败: ' + (error.message || '未知错误'))
        console.error('获取统计数据失败:', error)
      } finally {
        this.loading = false
      }
    },
    handleSizeChange(val) {
      this.pageSize = val
      this.currentPage = 1
    },
    handleCurrentChange(val) {
      this.currentPage = val
    },
    async fetchTableStats() {
      if (!this.queryTableName.trim()) {
        this.$message.warning('请输入表名')
        return
      }
      
      this.statsLoading = true
      try {
        const data = await getTableStatsSimple(this.queryTableName.trim())
        this.tableStats = data
        this.$message.success('RAG统计数据获取成功')
      } catch (error) {
        this.$message.error('获取RAG统计失败: ' + (error.message || '未知错误'))
        console.error('获取RAG统计失败:', error)
        this.tableStats = null
      } finally {
        this.statsLoading = false
      }
    },
    formatSampleData(data) {
      try {
        // 如果是对象，排除embedding字段并格式化
        if (typeof data === 'object') {
          const filteredData = { ...data }
          delete filteredData.embedding // 移除过长的embedding数组
          return JSON.stringify(filteredData, null, 2)
        }
        // 如果是字符串，尝试解析并格式化
        const parsed = JSON.parse(data)
        return JSON.stringify(parsed, null, 2)
      } catch (e) {
        return data
      }
    },
    formatContentData(content) {
      try {
        const parsed = JSON.parse(content)
        return JSON.stringify(parsed, null, 2)
      } catch (e) {
        return content
      }
    },
    quickQuery(tableName) {
      this.queryTableName = tableName
      this.fetchTableStats()
    }
  }
}
</script>

<style scoped>
.stardew-show-container {
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

.refresh-btn {
  color: #fff !important;
}

.refresh-btn:hover {
  color: #f0e68c !important;
}

.stats-query-section {
  margin-bottom: 20px;
}

.query-card {
  border: 1px solid #e6a23c;
}

.stats-result {
  margin-top: 15px;
}

.stats-content .stat-item {
  margin-bottom: 10px;
}

.stat-label {
  font-weight: bold;
  color: #606266;
}

.stat-value {
  color: #409eff;
  margin-left: 8px;
}

.sample-data {
  margin-top: 15px;
}

.sample-title {
  margin-bottom: 8px;
  color: #606266;
}

.error-section {
  margin-top: 15px;
}

.stat-card {
  cursor: pointer;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.clearfix:before,
.clearfix:after {
  display: table;
  content: "";
}

.clearfix:after {
  clear: both;
}

.sample-item {
  margin-bottom: 10px;
  display: flex;
  align-items: center;
}

.sample-label {
  font-weight: bold;
  color: #606266;
  margin-right: 8px;
  min-width: 60px;
}

.content-data {
  margin-top: 15px;
}
</style>
