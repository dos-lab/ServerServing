const Mock = require('mockjs')

const List = []
const count = 100

for (let i = 0; i < count; i++) {
  List.push(Mock.mock({
    id: '@increment',
    created_at: +Mock.Random.date('T'),
    updated_at: +Mock.Random.date('T'),
    name: '@first',
    pwd: '@first',
    admin: '@bool'
  }))
}

module.exports = [
  {
    url: '/app/v1/users/([0-9]+)',
    type: 'get',
    response: config => {
      const userID = config.params[0]
      // console.log('getInfo ', config)
      // console.log('userID = ', userID)
      for (const user of List) {
        if (user.id === +userID) {
          return {
            code: 20000,
            message: 'success',
            data: user
          }
        }
      }
      return {
        code: 20000,
        message: 'success',
        data: null
      }
    }
  },
  {
    url: '/app/v1/users/([0-9]+)',
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
    url: '/app/v1/users/',
    type: 'get',
    response: config => {
      const { searchKeyword, from = 0, size = 20 } = config.query

      const numFrom = +from
      const numSize = +size

      const mockList = List.filter(item => {
        return !(searchKeyword && item.name.indexOf(searchKeyword) < 0)
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
