const Mock = require('mockjs')

const List = []
const count = 100

const newMockAccount = function(host, port) {
  return Mock.mock({
    created_at: +Mock.Random.date('T'),
    updated_at: +Mock.Random.date('T'),
    deleted_at: null,
    name: '@first',
    pwd: '@first',
    host: host,
    port: port,
    uid: '@integer(1000, 1100)',
    gid: '@integer(1000, 1100)',
    not_exists_in_server: '@bool',
    backup_dir_info: {
      output: 'some output for backup dir info',
      failed_info: null,
      backup_dir: '/backup/123456.backup',
      path_exists: '@bool',
      dir_exists: '@bool'
    }
  })
}

const newMockServer = function() {
  const mockServer = Mock.mock({
    basic: {
      created_at: +Mock.Random.date('T'),
      updated_at: +Mock.Random.date('T'),
      deleted_at: +Mock.Random.date('T'),
      host: '@ip',
      port: '@integer(22, 50)',
      admin_account_name: '@first',
      admin_account_pwd: "@string('lower')",
      os_type: 'os_type_linux'
    },
    access_failed_info: null,
    account_infos: {
      output: 'some server original output for accunt infos',
      failed_info: null,
      accounts: []
    },
    hardware_info: {
      cpu_hardware_info: {
        output: 'some server original output for cpu mem processes usage info,some server \\noriginal output for cpu mem processes usage infosome server original output for cpu mem processes usage infosome server original output for cpu mem processes usage infosome server original\\n output for cpu mem processes usage infosome server original output for cpu mem processes usage infosome server original output for cpu mem processes usage infosome server original output for cp\\nu mem processes usage infosome server original output for cpu mem processes usage infosome server original output for cpu mem processes usage \ninfosome server original output for cpu mem processes usage infosome server original output for cpu mem processes usage infosome server original output for cpu mem processes usage infosome server original output for cpu mem processes usage infosome server original o\\nutput for cpu mem processes usage info\',',
        failed_info: null,
        info: {
          architecture: 'x86_64',
          model_name: 'Intel(R) Xeon(R) CPU E5-2682 v4 @ 2.50GHz',
          cores: 20,
          threads_per_core: 1
        }
      },
      gpu_hardware_info: {
        output: '00:02.0 VGA compatible controller: Cirrus Logic GD 5446\n',
        failed_info: null,
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
      failed_info: null,
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
      failed_info: null,
      infos: [
        {
          account_name: 'someuser',
          what: 'w -s -h -u'
        }
      ]
    },
    server_gpu_usage_info: {
      output: 'some server original output for gpu usages',
      failed_info: null
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

function filterWithOptions(copied, query) {
  console.log('filter Option', query)
  copied.account_infos = query.with_accounts === 'true' ? copied.account_infos : null
  copied.cpu_mem_processes_usage_info = query.with_cmp_usages === 'true' ? copied.cpu_mem_processes_usage_info : null
  copied.server_gpu_usage_info = query.with_gpu_usages === 'true' ? copied.server_gpu_usage_info : null
  copied.hardware_info = query.with_hardware_info === 'true' ? copied.hardware_info : null
  copied.remote_accessing_usage_info = query.with_remote_access_usages === 'true' ? copied.remote_accessing_usage_info : null
  return copied
}

module.exports = [
  {
    url: '/app/v1/servers/connection/([0-9.]+)/([0-9]+)',
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
          connected: true
        }
      }
    }
  },
  {
    url: '/app/v1/servers/',
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
    url: '/app/v1/servers/([0-9.]+)/([0-9]+)',
    type: 'get',
    response: config => {
      // console.log('get server, config', config)
      const host = config.params[0]
      const port = +config.params[1]
      // const {
      //   with_accounts,
      //   with_cmp_usages,
      //   with_gpu_usages,
      //   with_hardware_info,
      //   with_remote_access_usages
      // } = config.query
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
    }
  },
  {
    url: '/app/v1/servers/',
    type: 'get',
    response: config => {
      // console.log('get servers, config', config)
      const { from = 0, size = 20, searchKeyword } = config.query

      const numFrom = +from
      const numSize = +size

      const mockList = List.filter(item => {
        // console.log('filtering item', item, 'keyword', keyword)
        return !(searchKeyword && item.basic.host.indexOf(searchKeyword) < 0 && item.basic.admin_account_name.indexOf(searchKeyword) < 0)
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
    url: '/app/v1/servers/accounts/',
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
    url: '/app/v1/servers/',
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
