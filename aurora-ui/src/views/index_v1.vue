<template>
  <div class="stardew-dashboard">
    <!-- 星空背景 -->
    <div class="stardew-stars"></div>

    <panel-group @handleSetLineChartData="handleSetLineChartData" />

    <el-row style="background:#fff8e6;padding:20px 16px 16px;margin-bottom:24px;border-radius:12px;border:3px solid #b88646;">
      <line-chart :chart-data="lineChartData" />
    </el-row>

    <el-row :gutter="24">
      <el-col :xs="24" :sm="24" :lg="8">
        <div class="stardew-chart-wrapper">
          <raddar-chart />
        </div>
      </el-col>
      <el-col :xs="24" :sm="24" :lg="8">
        <div class="stardew-chart-wrapper">
          <pie-chart />
        </div>
      </el-col>
      <el-col :xs="24" :sm="24" :lg="8">
        <div class="stardew-chart-wrapper">
          <bar-chart />
        </div>
      </el-col>
    </el-row>

    <!-- 装饰元素 -->
    <div class="stardew-decor soil"></div>
    <div class="stardew-decor grass"></div>
  </div>
</template>

<script>
import PanelGroup from './dashboard/PanelGroup'
import LineChart from './dashboard/LineChart'
import RaddarChart from './dashboard/RaddarChart'
import PieChart from './dashboard/PieChart'
import BarChart from './dashboard/BarChart'

const lineChartData = {
  newVisitis: {
    expectedData: [100, 120, 161, 134, 105, 160, 165],
    actualData: [120, 82, 91, 154, 162, 140, 145]
  },
  messages: {
    expectedData: [200, 192, 120, 144, 160, 130, 140],
    actualData: [180, 160, 151, 106, 145, 150, 130]
  },
  purchases: {
    expectedData: [80, 100, 121, 104, 105, 90, 100],
    actualData: [120, 90, 100, 138, 142, 130, 130]
  },
  shoppings: {
    expectedData: [130, 140, 141, 142, 145, 150, 160],
    actualData: [120, 82, 91, 154, 162, 140, 130]
  }
}

export default {
  name: 'Index',
  components: {
    PanelGroup,
    LineChart,
    RaddarChart,
    PieChart,
    BarChart
  },
  data() {
    return {
      lineChartData: lineChartData.newVisitis
    }
  },
  methods: {
    handleSetLineChartData(type) {
      this.lineChartData = lineChartData[type]
    }
  }
}
</script>

<style lang="scss" scoped>
.stardew-dashboard {
  position: relative;
  min-height: calc(100vh - 60px);
  padding: 32px;
  background: var(--stardew-night-gradient);
  overflow: hidden;
  
  .stardew-stars {
    z-index: 1;
  }

  .stardew-chart-wrapper {
    background: var(--stardew-bg-panel);
    border: 3px solid var(--stardew-border-primary);
    box-shadow: 0 4px 0 var(--stardew-border-shadow), var(--stardew-shadow-medium);
    border-radius: var(--stardew-radius-medium);
    padding: 16px 16px 0;
    margin-bottom: 24px;
    position: relative;
    
    &::before {
      content: '';
      position: absolute;
      top: -3px;
      left: -3px;
      right: -3px;
      bottom: -3px;
      background: linear-gradient(45deg, var(--stardew-border-secondary), var(--stardew-border-primary));
      border-radius: var(--stardew-radius-medium);
      z-index: -1;
    }
  }
  
  .stardew-decor {
    z-index: 0;
  }
}

@media (max-width:1024px) {
  .stardew-dashboard {
    padding: 20px;
    
    .stardew-chart-wrapper {
      padding: 12px 12px 0;
      margin-bottom: 20px;
    }
  }
}
</style>
