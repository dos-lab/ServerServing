import request from '@/utils/request'

// params: from int, size int, searchKeyword string
export function getList(params) {
  return request({
    url: '/app/v1/users/',
    method: 'get',
    params: params
  })
}

export function getInfo(userID) {
  console.log('try getInfo', userID)
  return request({
    url: `/app/v1/users/${userID}`,
    method: 'get'
  })
}

export function update(userID, data) {
  return request({
    url: `/app/v1/users/${userID}`,
    method: 'put',
    data
  })
}
