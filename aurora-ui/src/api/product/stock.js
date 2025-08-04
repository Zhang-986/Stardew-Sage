import request from '@/utils/request'

// 查询商品库存管理列表
export function listStock(query) {
  return request({
    url: '/product/stock/list',
    method: 'get',
    params: query
  })
}

// 查询商品库存管理详细
export function getStock(id) {
  return request({
    url: '/product/stock/' + id,
    method: 'get'
  })
}

// 新增商品库存管理
export function addStock(data) {
  return request({
    url: '/product/stock',
    method: 'post',
    data: data
  })
}

// 修改商品库存管理
export function updateStock(data) {
  return request({
    url: '/product/stock',
    method: 'put',
    data: data
  })
}

// 删除商品库存管理
export function delStock(id) {
  return request({
    url: '/product/stock/' + id,
    method: 'delete'
  })
}
