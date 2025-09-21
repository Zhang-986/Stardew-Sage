import request from '@/utils/request'

// 查询星露谷列表
export function listWiki(query) {
  return request({
    url: '/stardew/wiki/list',
    method: 'get',
    params: query
  })
}

// 查询星露谷详细
export function getWiki(id) {
  return request({
    url: '/stardew/wiki/' + id,
    method: 'get'
  })
}

// 新增星露谷
export function addWiki(data) {
  return request({
    url: '/stardew/wiki',
    method: 'post',
    data: data
  })
}

// 修改星露谷
export function updateWiki(data) {
  return request({
    url: '/stardew/wiki',
    method: 'put',
    data: data
  })
}

// 删除星露谷
export function delWiki(id) {
  return request({
    url: '/stardew/wiki/' + id,
    method: 'delete'
  })
}
