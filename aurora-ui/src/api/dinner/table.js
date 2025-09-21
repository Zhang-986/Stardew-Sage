import request from '@/utils/request'

// 查询午市就餐-桌台列表
export function listTable(query) {
  return request({
    url: '/dinner/table/list',
    method: 'get',
    params: query
  })
}

// 查询午市就餐-桌台详细
export function getTable(tableId) {
  return request({
    url: '/dinner/table/' + tableId,
    method: 'get'
  })
}

// 新增午市就餐-桌台
export function addTable(data) {
  return request({
    url: '/dinner/table',
    method: 'post',
    data: data
  })
}

// 修改午市就餐-桌台
export function updateTable(data) {
  return request({
    url: '/dinner/table',
    method: 'put',
    data: data
  })
}

// 删除午市就餐-桌台
export function delTable(tableId) {
  return request({
    url: '/dinner/table/' + tableId,
    method: 'delete'
  })
}
