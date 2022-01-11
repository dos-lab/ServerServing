const Mock = require('mockjs')

const List = []
const count = 100

const newMockAccount = function(host, port) {
  const nameReg = /[a-zA-Z][_a-zA-Z0-9]{5,14}/
  const pwdReg = /[\w]{6,12}/
  const v = Mock.mock({
    name: nameReg,
    pwd: pwdReg,
    pwdExists: '@bool'
  })
  if (!v.pwdExists) {
    v.pwd = null
  }
  return Mock.mock({
    created_at: +Mock.Random.date('T'),
    updated_at: +Mock.Random.date('T'),
    deleted_at: null,
    name: v.name,
    pwd: v.pwd,
    host: host,
    port: port,
    uid: '@integer(1000, 1100)',
    gid: '@integer(1000, 1100)',
    not_exists_in_server: '@bool',
    backup_dir_info: {
      output: 'some output for backup dir info',
      failed_info: null,
      backup_dir: `/backup/${v.name}`,
      path_exists: '@bool',
      dir_exists: '@bool'
    }
  })
}

const mockFailedInfo = function(prob) {
  let failedInfo = null
  if (Math.floor(Math.random() * 100) < prob) {
    failedInfo = {
      cause_description: '模拟的偶发错误'
    }
  }
  return failedInfo
}

const newMockServer = function() {
  const access_failed_info = mockFailedInfo(10)
  const account_failed_info = mockFailedInfo(10)
  const cpu_hardware_info_failed = mockFailedInfo(10)
  const gpu_hardware_info_failed = mockFailedInfo(10)
  const cpu_mem_processes_usage_info_failed = mockFailedInfo(10)
  const remote_accessing_usage_info_failed = mockFailedInfo(10)
  const server_gpu_usage_info_failed = mockFailedInfo(10)
  const mockServer = Mock.mock({
    basic: {
      created_at: +Mock.Random.date('T'),
      updated_at: +Mock.Random.date('T'),
      deleted_at: +Mock.Random.date('T'),
      name: '@first',
      description: '@string(10, 50)',
      host: '@ip',
      port: '@integer(22, 50)',
      admin_account_name: '@first',
      admin_account_pwd: '@string(6, 10)',
      os_type: 'os_type_linux'
    },
    access_failed_info: access_failed_info,
    account_infos: {
      output: 'some server original output for accunt infos',
      failed_info: account_failed_info,
      accounts: []
    },
    hardware_info: {
      cpu_hardware_info: {
        output: 'some server original output for cpu mem processes usage info,some server \noriginal output for cpu mem processes usage infosome server original output for cpu mem processes usage infosome server original output for cpu mem processes usage infosome server original\n output for cpu mem processes usage infosome server original output for cpu mem processes usage infosome server original output for cpu mem processes usage infosome server original output for cp\nu mem processes usage infosome server original output for cpu mem processes usage infosome server original output for cpu mem processes usage \ninfosome server original output for cpu mem processes usage infosome server original output for cpu mem processes usage infosome server original output for cpu mem processes usage infosome server original output for cpu mem processes usage infosome server original o\nutput for cpu mem processes usage info\',',
        failed_info: cpu_hardware_info_failed,
        info: {
          architecture: 'x86_64',
          model_name: 'Intel(R) Xeon(R) CPU E5-2682 v4 @ 2.50GHz',
          cores: 20,
          threads_per_core: 1
        }
      },
      gpu_hardware_infos: {
        output: '00:02.0 VGA compatible controller: Cirrus Logic GD 5446\n',
        failed_info: gpu_hardware_info_failed,
        infos: [
          {
            product: 'RTX 3090'
          },
          {
            product: 'RTX 3080ti'
          }
        ]
      }
    },
    cpu_mem_processes_usage_info: {
      output: 'some server original output for cpu mem processes usage info',
      failed_info: cpu_mem_processes_usage_info_failed,
      cpu_mem_usage: {
        user_cpu_usage: '@float(0, 70)',
        mem_usage: '@float(0, 70)',
        mem_total: '8GB'
      },
      process_infos: [
        {
          pid: 1,
          command: 'systemd',
          owner_account_name: 'root',
          cpu_usage: 12.1,
          mem_usage: 0.4
        },
        {
          pid: 2,
          command: 'kthreadd',
          owner_account_name: 'root',
          cpu_usage: 8.4,
          mem_usage: 33.1
        },
        {
          pid: 3,
          command: 'systemd',
          owner_account_name: 'root',
          cpu_usage: 5.4,
          mem_usage: 2.3
        },
        {
          pid: 4,
          command: 'systemd',
          owner_account_name: 'root',
          cpu_usage: 1.4,
          mem_usage: 0.4
        }
      ]
    },
    remote_accessing_usage_info: {
      output: 'some server original output for remote accessing usage info',
      failed_info: remote_accessing_usage_info_failed,
      infos: [
        {
          account_name: 'someuser',
          what: 'w -s -h -u'
        }
      ]
    },
    server_gpu_usage_info: {
      output: 'some server original output for gpu usages',
      failed_info: server_gpu_usage_info_failed
    }
  })
  const accounts = []
  for (let i = 0; i < 25; i++) {
    accounts.push(newMockAccount(mockServer.basic.host, mockServer.basic.port))
  }
  mockServer.account_infos.accounts = accounts
  const processes = []
  const randomAccountName = () => {
    const i = Math.floor((Math.random() * accounts.length))
    return accounts[i].name
  }
  for (let i = 0; i < 22; i++) {
    processes.push(Mock.mock({
      pid: i,
      command: '@string(3, 7)',
      owner_account_name: randomAccountName(),
      cpu_usage: '@float(0, 50)',
      mem_usage: '@float(0, 50)'
      // gpu_usage: null
    }))
  }
  mockServer.cpu_mem_processes_usage_info.process_infos = processes
  return mockServer
}

for (let i = 0; i < count; i++) {
  const mockServer = newMockServer()
  List.push(mockServer)
}

const getServer = function(host, port) {
  for (let i = 0; i < List.length; i++) {
    const server = List[i]
    // console.log(`server.basic.host=[${server.basic.host}], server.basic.port=[${server.basic.port}], host=[${host}], port=[${port}]`)
    if (server.basic.host === host && server.basic.port === port) {
      return server
    }
  }
  return null
}

const getAccount = function(server, account_name) {
  for (const acc of server.account_infos.accounts) {
    if (acc.name === account_name) {
      return acc
    }
  }
  return null
}

function filterWithOptions(copied, query) {
  // console.log('filter Option', query)
  copied.account_infos = query.with_accounts === 'true' ? copied.account_infos : null
  copied.cpu_mem_processes_usage_info = query.with_cmp_usages === 'true' ? copied.cpu_mem_processes_usage_info : null
  copied.server_gpu_usage_info = query.with_gpu_usages === 'true' ? copied.server_gpu_usage_info : null
  copied.hardware_info = query.with_hardware_info === 'true' ? copied.hardware_info : null
  copied.remote_accessing_usage_info = query.with_remote_access_usages === 'true' ? copied.remote_accessing_usage_info : null
  return copied
}

module.exports = [
  {
    url: '/api/v1/servers/accounts/backupDir',
    type: 'get',
    response: config => {
      const { host, port, account_name } = config.query
      const numberPort = +port
      console.log('get account backupDir, request param, host, port, account_name', host, numberPort, account_name)
      const server = getServer(host, numberPort)
      if (server === null) {
        return {
          code: 40004,
          message: '服务器不存在！',
          data: {}
        }
      }
      const acc = getAccount(server, account_name)
      if (acc === null) {
        return {
          code: 40004,
          message: '服务器不存在！',
          data: {}
        }
      }
      return {
        code: 20000,
        message: 'success',
        data: acc.backup_dir_info
      }
    }
  },
  {
    url: '/api/v1/servers/accounts',
    type: 'post',
    response: config => {
      const { host, port, account_name, account_pwd } = config.body
      console.log('create accounts, request body, host, port, account_name, account_pwd', host, port, account_name, account_pwd)
      for (let i = 0; i < List.length; i++) {
        const server = List[i]
        if (server.basic.host === host && server.basic.port === port) {
          console.log('server accounts', server.account_infos.accounts)
          for (const acc of server.account_infos.accounts) {
            if (acc.name === account_name) {
              return {
                code: 40004,
                message: '该账户名已存在！',
                data: {}
              }
            }
          }
          const mockAccount = newMockAccount(host, port)
          mockAccount.name = account_name
          mockAccount.pwd = account_pwd
          server.account_infos.accounts.push(mockAccount)
          return {
            code: 20000,
            message: 'success',
            data: {}
          }
        }
      }
      return {
        code: 40004,
        message: '服务器不存在！',
        data: {}
      }
    }
  },
  {
    url: '/api/v1/servers/accounts',
    type: 'delete',
    response: config => {
      const { host, port, account_name, backup } = config.body
      console.log('delete account, request body, host, port, account_name, backup', host, port, account_name, backup)
      for (let i = 0; i < List.length; i++) {
        const server = List[i]
        if (server.basic.host === host && server.basic.port === port) {
          console.log('server accounts', server.account_infos.accounts)
          for (const acc of server.account_infos.accounts) {
            if (acc.name === account_name) {
              acc.not_exists_in_server = true
              if (backup) {
                acc.backup_dir_info.dir_exists = true
                acc.backup_dir_info.path_exists = true
                acc.backup_dir_info.backup_dir = '/backup/' + account_name
                acc.backup_dir_info.output = 'some output for backup dir info'
              }
              return {
                code: 20000,
                message: 'success',
                data: {
                  backup_dir: acc.backup_dir_info.backup_dir
                }
              }
            }
          }
          return {
            code: 40004,
            message: '该账户不存在！',
            data: {
              backup_dir: null
            }
          }
        }
      }
      return {
        code: 40004,
        message: '服务器不存在！',
        data: {}
      }
    }
  },
  {
    url: '/api/v1/servers/accounts',
    type: 'put',
    response: config => {
      const { host, port, account_name, account_pwd, recover, recover_backup } = config.body
      console.log('update account, request body, host, port, account_name, account_pwd, recover, recover_backup', host, port, account_name, account_pwd, recover, recover_backup)
      for (let i = 0; i < List.length; i++) {
        const server = List[i]
        if (server.basic.host === host && server.basic.port === port) {
          console.log('server accounts', server.account_infos.accounts)
          for (const acc of server.account_infos.accounts) {
            if (acc.name === account_name) {
              acc.not_exists_in_server = false
              acc.pwd = account_pwd
              if (recover_backup) {
                acc.backup_dir_info.dir_exists = false
                acc.backup_dir_info.path_exists = false
              }
              return {
                code: 20000,
                message: 'success',
                data: {}
              }
            }
          }
          return {
            code: 40004,
            message: '该账户不存在！',
            data: {}
          }
        }
      }
      return {
        code: 40004,
        message: '服务器不存在！',
        data: {}
      }
    }
  },
  {
    url: '/api/v1/servers/connections/([0-9.]+)/([0-9]+)',
    type: 'get',
    response: config => {
      // console.log('get servers, config', config)
      const host = config.params[0]
      const port = +config.params[1]

      console.log('connection test servers, mock, host, port, params', host, port, config.query)
      return {
        code: 20000,
        message: 'success',
        data: {
          connected: true,
          cause: '不知道为啥连不上！'
        }
      }
    }
  },
  {
    url: '/api/v1/servers',
    type: 'delete',
    response: config => {
      const { host, port } = config.body
      console.log('delete server, mock, host, port, body', host, port, config.body)
      for (let i = 0; i < List.length; i++) {
        const server = List[i]
        if (server.basic.host === host && server.basic.port === port) {
          List.splice(i, 1)
          return {
            code: 20000,
            message: 'success',
            data: {}
          }
        }
      }
      return {
        code: 40004,
        message: '该服务器不存在！',
        data: {}
      }
    }
  },
  {
    // 单独获取一个server的info
    url: '/api/v1/servers/([0-9.]+)/([0-9]+)',
    type: 'get',
    response: config => {
      const host = config.params[0]
      const port = +config.params[1]
      for (const server of List) {
        if (server.basic.host === host && server.basic.port === port) {
          let copied = Object.assign({}, server)
          copied = filterWithOptions(copied, config.query)
          return {
            code: 20000,
            message: 'success',
            data: copied
          }
        }
      }
      return {
        code: 20001,
        message: '服务器不存在',
        data: {}
      }
    }
  },
  {
    url: '/api/v1/servers/([0-9.]+)/([0-9]+)',
    type: 'put',
    response: config => {
      const host = config.params[0]
      const port = +config.params[1]
      const { name, description, admin_account_name, admin_account_pwd } = config.body
      console.log('update server, host, port, request body, name, description, admin_account_name, admin_account_pwd',
        host, port, name, description, admin_account_name, admin_account_pwd)
      for (const server of List) {
        // console.log(`server.basic.host=[${server.basic.host}], server.basic.port=[${server.basic.port}], host=[${host}], port=[${port}]`)
        if (server.basic.host === host && server.basic.port === port) {
          console.log('update server executed.')
          server.basic.admin_account_name = admin_account_name
          server.basic.admin_account_pwd = admin_account_pwd
          server.basic.name = name
          server.basic.description = description
          return {
            code: 20000,
            message: 'success',
            data: {}
          }
        }
      }
      return {
        code: 20001,
        message: '服务器不存在！',
        data: {}
      }
    }
  },
  {
    url: '/api/v1/servers',
    type: 'get',
    response: config => {
      // console.log('get servers, config', config)
      const { from = 0, size = 20, keyword } = config.query

      const numFrom = +from
      const numSize = +size

      const mockList = List.filter(item => {
        // console.log('filtering item', item, 'keyword', keyword)
        return !(keyword && item.basic.host.indexOf(keyword) < 0 && item.basic.admin_account_name.indexOf(keyword) < 0)
      })

      console.log('get servers, mock, from, size, from + size', numFrom, numSize, numFrom + numSize)

      const pageList = mockList.filter((item, index) => index >= numFrom && index < (numFrom + numSize))
      const copies = []
      for (const server of pageList) {
        let copied = Object.assign({}, server)
        // console.log('before filter server', copied)
        copied = filterWithOptions(copied, config.query)
        // console.log('after filter server', copied)
        copies.push(copied)
      }
      return {
        code: 20000,
        message: 'success',
        data: {
          total_count: mockList.length,
          infos: copies
        }
      }
    }
  },
  {
    url: '/api/v1/servers',
    type: 'post',
    response: config => {
      const { host, port, os_type, admin_account_name, admin_account_pwd } = config.body
      console.log('create server, request body, host, port, os_type, admin_account_name, admin_account_pwd', host, port, os_type, admin_account_name, admin_account_pwd)
      const mockServer = newMockServer()
      mockServer.basic.host = host
      mockServer.basic.port = port
      mockServer.basic.os_type = os_type
      mockServer.basic.admin_account_name = admin_account_name
      mockServer.basic.admin_account_pwd = admin_account_pwd
      List.push(mockServer)
      return {
        code: 20000,
        message: 'success',
        data: {}
      }
    }
  }
]
