import request from '@/utils/request'

export function login(data) {
  return request({
    url: '/api/v1/sessions',
    method: 'post',
    data
  })
}

export function register(name, pwd) {
  return request({
    url: '/api/v1/users',
    method: 'post',
    data: {
      name: name,
      pwd: pwd
    }
  })
}

export function logout() {
  return request({
    url: '/api/v1/sessions',
    method: 'delete'
  })
}

// params: from int, size int, searchKeyword string
export function getList(params) {
  return request({
    url: '/api/v1/users',
    method: 'get',
    params: params
  })
}

export function getInfo(token, userID) {
  console.log('try getInfo, token, userID', token, userID)
  return request({
    url: `/api/v1/users/${userID}`,
    method: 'get'
  })
}

export function update(userID, data) {
  return request({
    url: `/api/v1/users/${userID}`,
    method: 'put',
    data
  })
}
