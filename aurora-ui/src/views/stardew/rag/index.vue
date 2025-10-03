<template>
  <div class="stardew-rag-container">
    <el-card class="stardew-card">
      <div slot="header" class="stardew-header clearfix">
        <div class="header-content">
          <i class="stardew-icon">🧙‍♂️</i>
          <span class="header-title">Stardew Sage - 数据表格管理</span>
        </div>
        <el-button 
          style="float: right; padding: 3px 0" 
          type="text"
          @click="fetchTableData"
          class="refresh-btn"
        >
          刷新
        </el-button>
      </div>
      
      <div v-loading="loading">
        <el-empty v-if="!loading && paginatedTableList.length === 0" description="暂无stardew相关数据"></el-empty>
        
        <div v-else>
          <el-table 
            :data="paginatedTableList" 
            stripe
            style="width: 100%"
            border
          >
            <el-table-column 
              type="index" 
              label="序号" 
              width="80"
              :index="getTableIndex"
            ></el-table-column>
            
            <el-table-column 
              label="表格名称"
            >
              <template slot-scope="scope">
                <el-tag type="success">{{ scope.row }}</el-tag>
              </template>
            </el-table-column>
            
            <el-table-column 
              label="操作" 
              width="150"
              align="center"
            >
              <template slot-scope="scope">
                <el-button 
                  type="primary" 
                  size="mini"
                  @click="viewTableStructure(scope.row)"
                  :loading="structureLoading === scope.row"
                >
                  查看表结构
                </el-button>
              </template>
            </el-table-column>
          </el-table>
          
          <el-pagination
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
            :current-page="currentPage"
            :page-sizes="[5, 10, 20, 50]"
            :page-size="pageSize"
            layout="total, sizes, prev, pager, next, jumper"
            :total="filteredTableList.length"
            style="margin-top: 20px; text-align: center;"
          >
          </el-pagination>
        </div>
      </div>
    </el-card>

    <!-- 表结构信息对话框 -->
    <el-dialog
      title="表结构信息"
      :visible.sync="dialogVisible"
      width="80%"
      :before-close="handleDialogClose"
    >
      <div v-if="tableInfo">
        <el-descriptions :title="selectedTableName" border :column="2">
          <el-descriptions-item label="表名">{{ selectedTableName }}</el-descriptions-item>
          <el-descriptions-item label="样本数据数量">{{ (tableInfo.sample_data || []).length }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">字段信息</el-divider>
        <el-table 
          :data="tableInfo.column_data || []" 
          border 
          size="small"
          @selection-change="handleSelectionChange"
        >
          <el-table-column type="selection" width="55"></el-table-column>
          <el-table-column prop="columnName" label="字段名" min-width="150"></el-table-column>
          <el-table-column prop="columnComment" label="字段说明" min-width="200" show-overflow-tooltip></el-table-column>
        </el-table>
        
        <el-divider content-position="left">样本数据</el-divider>
        <el-table 
          :data="tableInfo.sample_data || []" 
          border 
          size="small" 
          max-height="400"
          style="width: 100%"
        >
          <el-table-column 
            v-for="column in tableInfo.column_data || []" 
            :key="column.columnName"
            :prop="column.columnName" 
            :label="column.columnComment || column.columnName"
            min-width="120"
            show-overflow-tooltip
          >
            <template slot-scope="scope">
              <span v-if="scope.row[column.columnName]">{{ scope.row[column.columnName] }}</span>
              <el-tag v-else type="info" size="mini">空</el-tag>
            </template>
          </el-table-column>
        </el-table>
      </div>
      
      <span slot="footer" class="dialog-footer">
        <el-button 
          type="success" 
          :disabled="selectedColumns.length === 0"
          :loading="uploadLoading"
          @click="uploadToRAG"
        >
          上传到RAG ({{ selectedColumns.length }}个字段)
        </el-button>
        <el-button @click="dialogVisible = false">关闭</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { getTableList, getTableInfo, uploadToRAG } from '@/api/ai'

export default {
  name: 'StardewRag',
  data() {
    return {
      tableList: [],
      loading: false,
      currentPage: 1,
      pageSize: 10,
      dialogVisible: false,
      tableInfo: null,
      selectedTableName: '',
      structureLoading: null,
      selectedColumns: [],
      uploadLoading: false
    }
  },
  computed: {
    // 过滤出stardew开头的表格
    filteredTableList() {
      return this.tableList.filter(table => 
        table.toLowerCase().startsWith('stardew')
      )
    },
    // 分页后的数据
    paginatedTableList() {
      const start = (this.currentPage - 1) * this.pageSize
      const end = start + this.pageSize
      return this.filteredTableList.slice(start, end)
    }
  },
  created() {
    this.fetchTableData()
  },
  methods: {
    async fetchTableData() {
      this.loading = true
      try {
        const data = await getTableList()
        this.tableList = data || []
        this.currentPage = 1 // 重置到第一页
        this.$message.success(`数据加载成功，找到${this.filteredTableList.length}个stardew相关表格`)
      } catch (error) {
        this.$message.error('数据加载失败: ' + (error.message || '未知错误'))
        console.error('获取表格数据失败:', error)
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
    getTableIndex(index) {
      return (this.currentPage - 1) * this.pageSize + index + 1
    },
    async viewTableStructure(tableName) {
      this.structureLoading = tableName
      try {
        const data = await getTableInfo(tableName)
        this.tableInfo = data
        this.selectedTableName = tableName
        this.dialogVisible = true
      } catch (error) {
        this.$message.error('获取表结构失败: ' + (error.message || '未知错误'))
        console.error('获取表结构失败:', error)
      } finally {
        this.structureLoading = null
      }
    },
    handleDialogClose() {
      this.dialogVisible = false
      this.tableInfo = null
      this.selectedTableName = ''
      this.selectedColumns = []
    },
    handleSelectionChange(selection) {
      this.selectedColumns = selection
    },
    async uploadToRAG() {
      if (this.selectedColumns.length === 0) {
        this.$message.warning('请选择要上传的字段')
        return
      }
      
      this.uploadLoading = true
      try {
        const columnSimpleMetas = this.selectedColumns.map(column => ({
          columnName: column.columnName,
          columnComment: column.columnComment
        }))
        
        const result = await uploadToRAG(this.selectedTableName, columnSimpleMetas)
        this.$message.success(`上传成功: ${result}`)
        this.dialogVisible = false
      } catch (error) {
        this.$message.error('上传失败: ' + (error.message || '未知错误'))
        console.error('上传到RAG失败:', error)
      } finally {
        this.uploadLoading = false
      }
    },
  }
}
</script>

<style scoped>
.stardew-rag-container {
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

.clearfix:before,
.clearfix:after {
  display: table;
  content: "";
}

.clearfix:after {
  clear: both;
}

.stats-content p {
  margin: 8px 0;
  font-size: 14px;
}

.sample-data {
  margin-top: 10px;
}

.error-text {
  color: #f56c6c;
  font-weight: bold;
}
</style>