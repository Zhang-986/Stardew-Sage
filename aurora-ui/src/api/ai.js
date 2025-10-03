import axios from 'axios'

// 创建专门用于星露谷物语后端的axios实例
const aiRequest = axios.create({
  baseURL: 'http://localhost:8099',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
aiRequest.interceptors.request.use(
  config => {
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
aiRequest.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    console.error('API请求失败:', error)
    return Promise.reject(error)
  }
)

// 获取所有表格数据
export function getTableList() {
  return aiRequest({
    url: '/api/table',
    method: 'get'
  })
}

// 获取表格结构信息
export function getTableInfo(tableName) {
  return aiRequest({
    url: '/api/table/info',
    method: 'get',
    params: {
      tableName: tableName
    }
  })
}

// 上传数据到RAG
export function uploadToRAG(tableName, columnSimpleMetas) {
  return aiRequest({
    url: '/api/table/uploadToRAG',
    method: 'post',
    data: {
      tableName: tableName,
      columnSimpleMetas: columnSimpleMetas
    }
  })
}

// 获取默认键分组统计
export function getDefaultKeysGroupedCount() {
  return aiRequest({
    url: '/api/table/default/keys/grouped-count',
    method: 'get'
  })
}

// 获取表格RAG向量数据统计
export function getTableStatsSimple(tableName) {
  return aiRequest({
    url: '/api/table/stats/simple',
    method: 'get',
    params: {
      tableName: tableName
    }
  })
}

// AI代理问答 - SSE流式响应
export function getInfoDetailSSE(question) {
  return new EventSource(`http://localhost:8099/api/agent/getInfoDetail?question=${encodeURIComponent(question)}`)
}

// AI代理问答 - RAG向量数据库 - SSE流式响应
export function getInfoRagDetailSSE(question) {
  return new EventSource(`http://localhost:8099/api/agent/getInfoRagDetail?question=${encodeURIComponent(question)}`)
}

// 分析图片 - 流式响应
export function analyzeImage(file, question) {
  const formData = new FormData()
  formData.append('file', file)
  if (question) {
    formData.append('question', question)
  }

  return fetch('http://localhost:8099/api/image/analyze', {
    method: 'POST',
    body: formData
    // fetch API 会自动设置 Content-Type 为 multipart/form-data
  })
}

export default aiRequest
