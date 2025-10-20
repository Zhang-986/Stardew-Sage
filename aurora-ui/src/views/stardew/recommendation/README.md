# 智能推荐系统

这是 Stardew Sage 的智能推荐模块。

## 📁 文件说明

- `index.vue` - 主界面组件（400+ 行）

## 🎯 功能特性

### 1. 个性化推荐
- 基于向量相似度的内容匹配
- 支持自然语言查询
- 卡片式展示推荐结果

### 2. 策略建议
- AI 驱动的策略分析
- 实时 SSE 流式响应
- 场景化的游戏建议

### 3. 上下文推荐
- 结合历史对话的智能推荐
- 支持上下文输入
- 个性化内容过滤

## 🔌 API 接口

### 个性化推荐
```javascript
GET /api/recommendation/personalized?query=查询内容&topK=5
```

### 策略建议（SSE）
```javascript
GET /api/recommendation/strategy?scenario=场景描述
```

### 上下文推荐（SSE）
```javascript
GET /api/recommendation/contextual?context=上下文&query=问题
```

## 💻 使用方式

### 在路由中注册

```javascript
{
  path: 'recommendation',
  component: () => import('@/views/stardew/recommendation/index'),
  name: 'Recommendation',
  meta: { title: '智能推荐', icon: 'guide' }
}
```

### 直接访问

```
http://localhost:80/stardew/recommendation
```

## 🎨 UI 特性

- 📱 响应式设计，支持移动端
- 🎨 渐变色排名徽章
- 📊 进度条显示推荐分数
- ⚡ 实时流式响应动画
- 💡 示例查询快速开始

## 🔧 自定义配置

### 修改推荐数量

在 `index.vue` 的 `getPersonalizedRecommendations` 方法中：

```javascript
const response = await this.$http.get('/api/recommendation/personalized', {
  params: {
    query: this.personalizedQuery,
    topK: 8  // 修改这里
  }
});
```

### 修改流式响应超时

在 `index.vue` 的 `getStrategyRecommendations` 方法中：

```javascript
setTimeout(() => {
  if (this.isStreaming) {
    eventSource.close();
    this.isStreaming = false;
    this.strategyLoading = false;
  }
}, 60000);  // 修改这里（毫秒）
```

### 添加新的示例查询

在 `data()` 中的 `exampleQueries` 数组添加：

```javascript
exampleQueries: [
  { label: '你的标签', query: '你的查询', type: 'personalized' },
  // ... 其他示例
]
```

## 🐛 调试技巧

### 启用调试日志

在浏览器控制台查看网络请求：

```javascript
// 在 methods 中添加
console.log('Query:', this.personalizedQuery);
console.log('Response:', response);
```

### 检查 SSE 连接

```javascript
eventSource.onerror = (error) => {
  console.error('SSE error:', error);
  console.log('ReadyState:', eventSource.readyState);
};
```

## 📈 性能优化

### 防抖处理

如果需要防止频繁请求，可以添加防抖：

```javascript
import { debounce } from 'lodash';

methods: {
  getPersonalizedRecommendations: debounce(function() {
    // 原有逻辑
  }, 500)
}
```

### 缓存结果

```javascript
data() {
  return {
    cache: new Map()
  }
},
methods: {
  async getPersonalizedRecommendations() {
    const cacheKey = this.personalizedQuery;
    if (this.cache.has(cacheKey)) {
      this.recommendations = this.cache.get(cacheKey);
      return;
    }
    // 获取数据
    this.cache.set(cacheKey, response.data);
  }
}
```

## 🎓 学习资源

- Vue.js 文档: https://vuejs.org/
- Element UI 文档: https://element.eleme.io/
- SSE 规范: https://html.spec.whatwg.org/multipage/server-sent-events.html

## 📝 更新日志

### v1.0.0 (2024-10-20)
- ✨ 初始版本
- 🎯 支持三种推荐模式
- 🎨 美观的 UI 设计
- ⚡ 实时流式响应

---

**需要帮助？** 查看项目根目录的完整文档：
- `快速启动指南.md`
- `RECOMMENDATION_SYSTEM_GUIDE.md`
