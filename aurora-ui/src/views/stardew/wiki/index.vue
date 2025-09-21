<template>
  <div class="app-container">
    <el-form :model="queryParams" ref="queryForm" size="small" :inline="true" v-show="showSearch" label-width="68px">
      <el-form-item label="主题" prop="title">
        <el-input
          v-model="queryParams.title"
          placeholder="请输入主题"
          clearable
          @keyup.enter.native="handleQuery"
        />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" icon="el-icon-search" size="mini" @click="handleQuery">搜索</el-button>
        <el-button icon="el-icon-refresh" size="mini" @click="resetQuery">重置</el-button>
      </el-form-item>
    </el-form>

    <el-row :gutter="10" class="mb8">
      <el-col :span="1.5">
        <el-button
          type="primary"
          plain
          icon="el-icon-plus"
          size="mini"
          @click="handleAdd"
          v-hasPermi="['stardew:wiki:add']"
        >新增</el-button>
      </el-col>
      <el-col :span="1.5">
        <el-button
          type="success"
          plain
          icon="el-icon-edit"
          size="mini"
          :disabled="single"
          @click="handleUpdate"
          v-hasPermi="['stardew:wiki:edit']"
        >修改</el-button>
      </el-col>
      <el-col :span="1.5">
        <el-button
          type="danger"
          plain
          icon="el-icon-delete"
          size="mini"
          :disabled="multiple"
          @click="handleDelete"
          v-hasPermi="['stardew:wiki:remove']"
        >删除</el-button>
      </el-col>
      <el-col :span="1.5">
        <el-button
          type="warning"
          plain
          icon="el-icon-download"
          size="mini"
          @click="handleExport"
          v-hasPermi="['stardew:wiki:export']"
        >导出</el-button>
      </el-col>
      <right-toolbar :showSearch.sync="showSearch" @queryTable="getList"></right-toolbar>
    </el-row>

    <el-table v-loading="loading" :data="wikiList" @selection-change="handleSelectionChange" size="small">
      <el-table-column type="selection" width="50" align="center" />
      <el-table-column label="主题" align="left" prop="title" width="200" show-overflow-tooltip />
      <el-table-column label="分类" align="center" prop="categories" width="180" show-overflow-tooltip>
        <template slot-scope="scope">
          <div class="categories-tags">
            <el-tag
              v-for="(category, index) in parseCategoriesArray(scope.row.categories)"
              :key="index"
              :type="getCategoryTagType(index)"
              size="mini"
              style="margin: 2px;"
            >
              {{ category }}
            </el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="内容" align="left" prop="content" min-width="300" show-overflow-tooltip>
        <template slot-scope="scope">
          <div class="content-preview" @click="showContentDialog(scope.row)" style="cursor: pointer;">
            {{ stripHtmlTags(scope.row.content) }}
          </div>
        </template>
      </el-table-column>
      <el-table-column label="来源" align="center" prop="source" width="100" show-overflow-tooltip />
      <el-table-column label="操作" align="center" width="120" fixed="right">
        <template slot-scope="scope">
          <el-button
            size="mini"
            type="text"
            icon="el-icon-edit"
            @click="handleUpdate(scope.row)"
            v-hasPermi="['stardew:wiki:edit']"
            style="margin-right: 8px;"
          >修改</el-button>
          <el-button
            size="mini"
            type="text"
            icon="el-icon-delete"
            @click="handleDelete(scope.row)"
            v-hasPermi="['stardew:wiki:remove']"
            style="color: #f56c6c;"
          >删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <pagination
      v-show="total>0"
      :total="total"
      :page.sync="queryParams.pageNum"
      :limit.sync="queryParams.pageSize"
      @pagination="getList"
    />

    <!-- 添加或修改维基对话框 -->
    <el-dialog :title="title" :visible.sync="open" width="600px" append-to-body>
      <el-form ref="form" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="主题" prop="title">
          <el-input v-model="form.title" placeholder="请输入主题" />
        </el-form-item>
        <el-form-item label="分类" prop="categories">
          <el-input v-model="form.categories" type="textarea" placeholder="请输入内容" />
        </el-form-item>
        <el-form-item label="内容">
          <editor v-model="form.content" :min-height="192"/>
        </el-form-item>
        <el-form-item label="来源" prop="source">
          <el-input v-model="form.source" type="textarea" placeholder="请输入内容" />
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button type="primary" @click="submitForm">确 定</el-button>
        <el-button @click="cancel">取 消</el-button>
      </div>
    </el-dialog>

    <!-- 内容详情对话框 -->
    <el-dialog :title="contentDialogTitle" :visible.sync="contentDialogVisible" width="800px" append-to-body>
      <div class="content-detail" v-html="processContentHtml(contentDialogContent)"></div>
      <div slot="footer" class="dialog-footer">
        <el-button @click="contentDialogVisible = false">关 闭</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { listWiki, getWiki, delWiki, addWiki, updateWiki } from "@/api/stardew/wiki"

export default {
  name: "Wiki",
  data() {
    return {
      // 遮罩层
      loading: true,
      // 选中数组
      ids: [],
      // 非单个禁用
      single: true,
      // 非多个禁用
      multiple: true,
      // 显示搜索条件
      showSearch: true,
      // 总条数
      total: 0,
      // 维基表格数据
      wikiList: [],
      // 弹出层标题
      title: "",
      // 是否显示弹出层
      open: false,
      // 查询参数
      queryParams: {
        pageNum: 1,
        pageSize: 10,
        title: null,
        content: null,
      },
      // 表单参数
      form: {},
      // 表单校验
      rules: {
        title: [
          { required: true, message: "主题不能为空", trigger: "blur" }
        ],
      },
      // 内容详情对话框
      contentDialogVisible: false,
      contentDialogTitle: "",
      contentDialogContent: ""
    }
  },
  created() {
    this.getList()
  },
  methods: {
    /** 查询维基列表 */
    getList() {
      this.loading = true
      listWiki(this.queryParams).then(response => {
        this.wikiList = response.rows
        this.total = response.total
        this.loading = false
      })
    },
    // 取消按钮
    cancel() {
      this.open = false
      this.reset()
    },
    // 表单重置
    reset() {
      this.form = {
        id: null,
        title: null,
        categories: null,
        content: null,
        source: null
      }
      this.resetForm("form")
    },
    /** 搜索按钮操作 */
    handleQuery() {
      this.queryParams.pageNum = 1
      this.getList()
    },
    /** 重置按钮操作 */
    resetQuery() {
      this.resetForm("queryForm")
      this.handleQuery()
    },
    // 多选框选中数据
    handleSelectionChange(selection) {
      this.ids = selection.map(item => item.id)
      this.single = selection.length!==1
      this.multiple = !selection.length
    },
    /** 新增按钮操作 */
    handleAdd() {
      this.reset()
      this.open = true
      this.title = "添加维基"
    },
    /** 修改按钮操作 */
    handleUpdate(row) {
      this.reset()
      const id = row.id || this.ids
      getWiki(id).then(response => {
        this.form = response.data
        this.open = true
        this.title = "修改维基"
      })
    },
    /** 提交按钮 */
    submitForm() {
      this.$refs["form"].validate(valid => {
        if (valid) {
          if (this.form.id != null) {
            updateWiki(this.form).then(response => {
              this.$modal.msgSuccess("修改成功")
              this.open = false
              this.getList()
            })
          } else {
            addWiki(this.form).then(response => {
              this.$modal.msgSuccess("新增成功")
              this.open = false
              this.getList()
            })
          }
        }
      })
    },
    /** 删除按钮操作 */
    handleDelete(row) {
      const ids = row.id || this.ids
      this.$modal.confirm('是否确认删除维基编号为"' + ids + '"的数据项？').then(function() {
        return delWiki(ids)
      }).then(() => {
        this.getList()
        this.$modal.msgSuccess("删除成功")
      }).catch(() => {})
    },
    /** 导出按钮操作 */
    handleExport() {
      this.download('stardew/wiki/export', {
        ...this.queryParams
      }, `wiki_${new Date().getTime()}.xlsx`)
    },
    /** 解析分类JSON字符串为数组 */
    parseCategoriesArray(categories) {
      if (!categories) return []

      try {
        // 如果是JSON数组字符串
        if (categories.startsWith('[') && categories.endsWith(']')) {
          return JSON.parse(categories)
        }
        // 如果是逗号分隔的字符串
        else if (categories.includes(',')) {
          return categories.split(',').map(item => item.trim()).filter(item => item)
        }
        // 单个分类
        else {
          return [categories.trim()]
        }
      } catch (error) {
        // 解析失败时按逗号分隔处理
        return categories.split(',').map(item => item.trim()).filter(item => item)
      }
    },
    /** 获取标签类型 */
    getCategoryTagType(index) {
      const types = ['', 'success', 'info', 'warning', 'danger']
      return types[index % types.length]
    },
    /** 显示内容详情弹窗 */
    showContentDialog(row) {
      this.contentDialogTitle = row.title || '内容详情'
      this.contentDialogContent = row.content || '暂无内容'
      this.contentDialogVisible = true
    },
    /** 去除HTML标签，只保留纯文本 */
    stripHtmlTags(htmlString) {
      if (!htmlString) return ''

      // 创建一个临时div元素来解析HTML
      const tempDiv = document.createElement('div')
      tempDiv.innerHTML = htmlString

      // 获取纯文本内容
      return tempDiv.textContent || tempDiv.innerText || ''
    },
    /** 处理内容HTML，将金钱标签转换为图标显示 */
    processContentHtml(htmlString) {
      if (!htmlString) return ''
      
      let processedHtml = htmlString
      
      // 清理CSS样式相关的标签和内容
      processedHtml = processedHtml.replace(
        /\.mw-parser-output[^}]*}/g, 
        ''
      )
      
      // 清理样式标签
      processedHtml = processedHtml.replace(
        /<style[^>]*>[\s\S]*?<\/style>/gi, 
        ''
      )
      
      // 清理class属性复杂的span标签，保留内容
      processedHtml = processedHtml.replace(
        /<span[^>]*class="[^"]*mw-parser-output[^"]*"[^>]*>(.*?)<\/span>/gi, 
        '$1'
      )
      
      // 处理金钱标签 data-sort-value="数字">
      processedHtml = processedHtml.replace(
        /data-sort-value="(\d+)">([^<]*)/g, 
        '<span class="money-tag"><i class="el-icon-coin" style="color: #f39c12; margin-right: 2px;"></i><strong style="color: #f39c12;">$1金</strong></span>'
      )
      
      // 处理星星币标签
      processedHtml = processedHtml.replace(
        /(\d+)\s*星星币/g,
        '<span class="star-coin-tag"><i class="el-icon-star-on" style="color: #ffd700; margin-right: 2px;"></i><strong style="color: #ffd700;">$1星星币</strong></span>'
      )
      
      // 处理卡利科三花蛋
      processedHtml = processedHtml.replace(
        /(\d+)\s*卡利科三花蛋/g,
        '<span class="egg-tag"><i class="el-icon-present" style="color: #e67e22; margin-right: 2px;"></i><strong style="color: #e67e22;">$1卡利科三花蛋</strong></span>'
      )
      
      // 清理多余的空白字符和换行
      processedHtml = processedHtml.replace(/\s+/g, ' ').trim()
      
      return processedHtml
    }
  }
}
</script>

<style scoped>
.content-preview {
  max-height: 60px;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  transition: all 0.3s ease;
}

.content-preview:hover {
  background-color: #f5f7fa;
  border-radius: 4px;
  padding: 4px;
}

.content-detail {
  max-height: 500px;
  overflow-y: auto;
  line-height: 1.6;
  font-size: 14px;
  padding: 10px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  background-color: #fafafa;
}

.app-container {
  padding: 20px;
}

.mb8 {
  margin-bottom: 8px;
}

.dialog-footer {
  text-align: center;
}

.el-table .el-table__cell {
  padding: 8px 0;
}

.categories-tags {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 2px;
}

/* 金钱相关标签样式 */
.content-detail .money-tag,
.content-detail .star-coin-tag,
.content-detail .egg-tag {
  display: inline-block;
  background: rgba(243, 156, 18, 0.1);
  border: 1px solid #f39c12;
  border-radius: 12px;
  padding: 2px 8px;
  margin: 0 2px;
  font-weight: bold;
}

.content-detail .star-coin-tag {
  background: rgba(255, 215, 0, 0.1);
  border-color: #ffd700;
}

.content-detail .egg-tag {
  background: rgba(230, 126, 34, 0.1);
  border-color: #e67e22;
}
</style>
