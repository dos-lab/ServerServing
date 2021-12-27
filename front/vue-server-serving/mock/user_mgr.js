const Mock = require('mockjs')

const List = []
const count = 100

List.push({
  id: 0,
  created_at: +Mock.Random.date('T'),
  updated_at: +Mock.Random.date('T'),
  name: 'admin',
  pwd: 111111,
  admin: true
})

List.push({
  id: 1123123,
  created_at: +Mock.Random.date('T'),
  updated_at: +Mock.Random.date('T'),
  name: 'normal_user',
  pwd: 111111,
  admin: false
})

const mockUser = function() {
  return Mock.mock({
    id: '@increment',
    created_at: +Mock.Random.date('T'),
    updated_at: +Mock.Random.date('T'),
    name: '@first',
    pwd: '@first',
    admin: '@bool'
  })
}

for (let i = 0; i < count; i++) {
  List.push(mockUser())
}

// const tokens = {
//   admin: {
//     token: 'admin-token'
//   },
//   editor: {
//     token: 'editor-token'
//   }
// }

const getUserByName = function(name) {
  for (const u of List) {
    if (u.name === name) {
      return u
    }
  }
  return null
}

const getUserByID = function(id) {
  for (const u of List) {
    if (u.id === id) {
      return u
    }
  }
  return null
}

module.exports = [
  // user login
  {
    url: '/api/v1/sessions/',
    type: 'post',
    response: config => {
      const { name, pwd } = config.body
      console.log('user login', name, pwd)
      const u = getUserByName(name)
      if (u !== null) {
        return {
          code: 20000,
          message: '',
          data: {
            token: `${u.id}`
          }
        }
      }
      return {
        code: 40004,
        message: '用户不存在！',
        data: {}
      }
    }
  },
  {
    url: '/api/v1/sessions/',
    type: 'delete',
    response: _ => {
      return {
        code: 20000,
        message: 'success',
        data: {}
      }
    }
  },
  {
    url: '/api/v1/users/([0-9]+)',
    type: 'get',
    response: config => {
      const userID = +config.params[0]
      const { token } = config.query
      const headerToken = +config.headers['x-token']
      console.log('get user info token, headerToken', token, headerToken)
      console.log('get user info token, userID', userID)
      const u = getUserByID(headerToken)
      if (u !== null) {
        return {
          code: 20000,
          message: 'success',
          data: u
        }
      }
      return {
        code: 40004,
        message: '用户不存在！',
        data: {}
      }
    }
  },
  {
    url: '/api/v1/users/',
    type: 'post',
    response: config => {
      const { name, pwd } = config.body
      for (const user of List) {
        if (user.name === name) {
          return {
            code: 40004,
            message: '用户名已存在！',
            data: {}
          }
        }
      }
      const u = mockUser()
      u.name = name
      u.pwd = pwd
      List.push(u)
      return {
        code: 20000,
        message: 'success',
        data: {
          token: `${u.id}`
        }
      }
    }
  },
  {
    url: '/api/v1/users/([0-9]+)',
    type: 'put',
    response: config => {
      const userID = config.params[0]
      console.log('put user', config.body)
      const { name, pwd, admin } = config.body
      // const name = config.body.name
      // const pwd = config.body.pwd
      // const admin = config.body.admin
      for (const user of List) {
        if (user.id === +userID) {
          console.log('found one user to be update', user)
          user.name = name || user.name
          user.pwd = pwd || user.pwd
          if (typeof admin !== 'undefined') {
            user.admin = admin
          }
          return {
            code: 20000,
            message: 'success',
            data: null
          }
        }
      }
      console.log('cannot find user to be update, userID', userID)
      return {
        code: 40000,
        message: 'failed',
        data: null
      }
    }
  },
  {
    url: '/api/v1/users/',
    type: 'get',
    response: config => {
      const { search_keyword, from = 0, size = 20 } = config.query

      const numFrom = +from
      const numSize = +size

      const mockList = List.filter(item => {
        return !(search_keyword && item.name.indexOf(search_keyword) < 0)
      })

      console.log('get users, mock, from, size, from + size', numFrom, numSize, numFrom + numSize)

      const pageList = mockList.filter((item, index) => index >= numFrom && index < (numFrom + numSize))

      return {
        code: 20000,
        message: 'success',
        data: {
          total_count: mockList.length,
          infos: pageList
        }
      }
    }
  }
]
