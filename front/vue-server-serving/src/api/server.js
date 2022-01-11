import request from '@/utils/request'

// params: from int, size int, searchKeyword string
export function getList(params) {
  console.log('try get server Info List', params)
  return request({
    url: '/api/v1/servers',
    method: 'get',
    params: params
  })
}

export function getInfo(host, port, params) {
  if (params.with_accounts === undefined) {
    params.with_accounts = false
  }
  if (params.with_cmp_usages === undefined) {
    params.with_cmp_usages = false
  }
  if (params.with_hardware_info === undefined) {
    params.with_hardware_info = false
  }
  if (params.with_gpu_usages === undefined) {
    params.with_gpu_usages = false
  }
  if (params.with_remote_access_usages === undefined) {
    params.with_remote_access_usages = false
  }
  console.log('try get server Info, host, port, params', host, port, params)
  return request({
    url: `/api/v1/servers/${host}/${port}`,
    method: 'get',
    params: params
  })
}

export function connectionTest(host, port, os_type, account_name, account_pwd) {
  console.log('try connectionTest server Info, host, port, params', host, port, os_type, account_name, account_pwd)
  return request({
    url: `/api/v1/servers/connections/${host}/${port}`,
    method: 'get',
    params: {
      os_type: os_type,
      account_name: account_name,
      account_pwd: account_pwd
    }
  })
}

export function createServer(name, description, host, port, os_type, admin_account_name, admin_account_pwd) {
  return request({
    url: `/api/v1/servers`,
    method: 'post',
    data: {
      name: name,
      description: description,
      host: host,
      port: port,
      os_type: os_type,
      admin_account_name: admin_account_name,
      admin_account_pwd: admin_account_pwd
    }
  })
}

export function deleteServer(host, port) {
  return request({
    url: `/api/v1/servers`,
    method: 'delete',
    data: {
      host: host,
      port: port
    }
  })
}

export function createAccount(host, port, account_name, account_pwd) {
  return request({
    url: `/api/v1/servers/accounts`,
    method: 'post',
    data: {
      host: host,
      port: port,
      account_name: account_name,
      account_pwd: account_pwd
    }
  })
}

export function deleteAccount(host, port, account_name, doBackup) {
  return request({
    url: `/api/v1/servers/accounts`,
    method: 'delete',
    data: {
      host: host,
      port: port,
      account_name: account_name,
      backup: doBackup
    }
  })
}

export function recoverAccount(host, port, account_name, account_pwd, recoverBackup) {
  return request({
    url: `/api/v1/servers/accounts`,
    method: 'put',
    data: {
      host: host,
      port: port,
      account_name: account_name,
      account_pwd: account_pwd,
      recover: true,
      recover_backup: recoverBackup
    }
  })
}

export function backupDirInfo(host, port, account_name) {
  return request({
    url: `/api/v1/servers/accounts/backupDir`,
    method: 'get',
    params: {
      host: host,
      port: port,
      account_name: account_name
    }
  })
}

export function updateServer(host, port, name, description, admin_account_name, admin_account_pwd) {
  return request({
    url: `/api/v1/servers/${host}/${port}`,
    method: 'put',
    data: {
      admin_account_name: admin_account_name,
      admin_account_pwd: admin_account_pwd,
      description: description,
      name: name
    }
  })
}
