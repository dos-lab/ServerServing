<template>
  <div class="app-container">
    <el-container>
      <el-main>
        <el-row type="flex" align="middle" :gutter="8">
          <el-col :span="12">
            <div style="display: flex; align-items: center; justify-content: left">
              <el-input v-model="listQuery.searchKeyword" placeholder="关键词" style="width: 200px;" class="filter-item" @keyup.enter.native="handleFilter" />
              <el-button v-waves style="margin-left: 20px" class="filter-item" type="primary" icon="el-icon-search" @click="handleFilter">
                搜索
              </el-button>
              <el-button v-waves v-permission="['admin']" style="margin-left: 20px" type="primary" @click="handleRegisterServerButtonClick">
                注册服务器
              </el-button>
            </div>
          </el-col>
        </el-row>
        <el-row style="margin-top: 20px">
          <el-col :span="24">
            <el-collapse
              :key="collapseKey"
              v-model="activeNames"
              v-loading="listLoading"
              @change="handleCollapseChange"
            >
              <div v-for="server in list" :key="genCollapseName(server)">
                <keep-alive>
                  <el-collapse-item :name="genCollapseName(server)" class="servers-collapse-item" style="border-bottom: 1px solid #EEEEEE;box-sizing: border-box;">
                    <span slot="title" class="servers-collapse-title">
                      <div style="min-width: 20px; margin-left: 20px;">
                        IP:
                      </div>
                      <div class="link-type" style="min-width: 100px;">
                        {{ server.basic.host }}
                      </div>
                      <div style="min-width: 40px; margin-left: 20px;">
                        Port:
                      </div>
                      <div style="min-width: 100px;">
                        {{ server.basic.port }}
                      </div>
                      <div style="min-width: 300px;">
                        管理员账户: {{ server.basic.admin_account_name }}
                      </div>
                      <div style="min-width: 250px;">
                        CPU(%): {{ extractCPUMemUsage(server, 'user_cpu_usage') }}
                      </div>
                      <div style="min-width: 200px;">
                        Memory(%): {{ `${extractCPUMemUsage(server, 'mem_usage')} (${extractCPUMemUsage(server, 'mem_total')})` || '未知' }}
                      </div>
                    </span>
                    <div style="margin-left: 20px">
                      <div v-if="collapsed[genCollapseName(server)]">
                        <server-panel :host="server.basic.host" :port="server.basic.port" @server-change="handleServerChange" @server-delete="handleServerDelete" />
                      </div>
                    </div>
                  </el-collapse-item>
                </keep-alive>
              </div>
            </el-collapse>
          </el-col>
        </el-row>
      </el-main>
    </el-container>

    <el-dialog title="注册服务器" :visible.sync="registerServerFormVisible">
      <el-form ref="dataForm" label-position="top" :rules="rules" :model="registerServerModel" style="width: 80%; margin-left:50px;">
        <el-form-item label="服务器ip地址" prop="host">
          <el-input v-model="registerServerModel.host" />
        </el-form-item>
        <el-form-item label="端口号" prop="port">
          <el-input-number v-model="registerServerModel.port" size="small" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="操作系统类型" prop="os_type">
          <el-radio v-model="registerServerModel.os_type_label" label="linux">linux</el-radio>
          <el-radio v-model="registerServerModel.os_type_label" label="windows server">windows server</el-radio>
        </el-form-item>
        <el-form-item label="管理员账户（服务器使用该账户通信）" prop="admin_account_name">
          <el-input v-model="registerServerModel.admin_account_name" />
        </el-form-item>
        <el-form-item label="管理员账户密码" prop="admin_account_pwd">
          <el-input v-model="registerServerModel.admin_account_pwd" />
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button :loading="registerServerConnectionTestLoading" @click="handleRegisterServerConnectionTest">
          测试连通性
        </el-button>
        <el-button type="primary" :loading="registerServerConfirmLoading" @click="handleRegisterServerConfirm">
          注册
        </el-button>
      </div>
    </el-dialog>

    <pagination v-show="total>0" :total="total" :page.sync="listQuery.page" :limit.sync="listQuery.limit" @pagination="getServers(defaultGetServerParams)" />

  </div>
</template>

<script>
import { getList, connectionTest, createServer } from '@/api/server'
import waves from '@/directive/waves' // waves directive
import Pagination from '@/components/Pagination' // secondary package based on el-pagination
import ServerPanel from '@/views/servers/components/ServerPanel'
import permission from '@/directive/permission'
// import Collapse from './components/Collapse'
// import CollapseItem from './components/CollapseItem'

const osTypesLabel2Value = { 'linux': 'os_type_linux', 'windows server': 'os_type_windows_server' }

export default {
  name: 'Servers',
  components: { Pagination, ServerPanel },
  directives: { waves, permission },
  filters: {
    isAdminFilter(isAdmin) {
      return isAdmin ? 'success' : 'info'
    }
  },
  data() {
    const hostReg = /^(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])$/
    const validateHost = (rule, value, callback) => {
      if (!hostReg.test(value)) {
        callback(new Error('不合法的ip地址！'))
      } else {
        callback()
      }
    }
    const nameReg = /^[a-zA-Z][0-9A-Za-z_]{2,14}/
    const validateName = (rule, value, callback) => {
      if (!nameReg.test(value)) {
        callback(new Error('账户名仅支持字母开头的数字，字母，以及下划线的组合，并且大于等于3位，不可超过15位！'))
      } else {
        callback()
      }
    }
    return {
      activeNames: [],
      tableKey: 0,
      collapseKey: 1,
      list: null,
      total: 0,
      listLoading: true,
      listQuery: {
        page: 1,
        limit: 20,
        searchKeyword: '',
        sort: '+id'
      },
      importanceOptions: [1, 2, 3],
      // calendarTypeOptions,
      sortOptions: [{ label: 'ID Ascending', key: '+id' }, { label: 'ID Descending', key: '-id' }],
      statusOptions: ['published', 'draft', 'deleted'],
      showReviewer: false,
      registerServerModel: {
        host: '0.0.0.0',
        port: 22,
        admin_account_name: '',
        admin_account_pwd: '',
        os_type_label: 'linux'
      },
      registerServerFormVisible: false,
      registerServerConnectionTestLoading: false,
      registerServerConfirmLoading: false,
      rules: {
        host: [
          { required: true, message: 'ip地址不能为空', trigger: 'change' },
          { validator: validateHost, trigger: 'change' }
        ],
        admin_account_name: [
          { required: true, message: '管理员账户名不能为空', trigger: 'change' },
          { validator: validateName, trigger: 'change' }
        ],
        admin_account_pwd: [
          { required: true, message: '管理员密码不能为空', trigger: 'change' }
        ]
      },
      downloadLoading: false,
      defaultGetServerParams: {
        with_accounts: false,
        with_cmp_usages: true,
        with_gpu_usages: false,
        with_hardware_info: false,
        with_remote_access_usages: false
      },
      collapsed: {}
    }
  },
  created() {
    this.getServers(this.defaultGetServerParams)
  },
  methods: {
    genCollapseName(server) {
      return `host_${server.basic.host}#port_${server.basic.port}`
    },
    getServers(params) {
      this.listLoading = true
      const copied = Object.assign({}, params)
      const from = +(this.listQuery.page - 1) * (+this.listQuery.limit)
      const size = +this.listQuery.limit
      copied.from = from
      copied.size = size
      copied.keyword = this.listQuery.searchKeyword
      console.log('getServers, param', copied)
      getList(copied).then(res => {
        console.log('initServers res', res)
        this.list = res.data.infos
        this.total = res.data.total_count
      }).finally(() => {
        // Just to simulate the time of the request
        this.listLoading = false
      })
    },
    handleFilter() {
      this.listQuery.page = 1
      this.getServers(this.defaultGetServerParams)
    },
    toServerDetailPage(row) {
      console.log('toServerDetailPage row', row)
    },
    extractCPUMemUsage(server, field) {
      if (server.cpu_mem_processes_usage_info && server.cpu_mem_processes_usage_info.failed_info === null) {
        const cpu_mem_usage = server.cpu_mem_processes_usage_info.cpu_mem_usage
        const value = cpu_mem_usage[field]
        // console.log('extractCPUMemUsage', value, typeof value)
        if (typeof value === 'number' && !isNaN(value)) {
          return value.toFixed(2)
        }
        return value
      }
      return '未知'
    },
    handleCollapseChange(val) {
      console.log('servers collapse change, val', val)
      for (const collapseName of val) {
        this.collapsed[collapseName] = true
      }
    },
    handleServerChange(server) {
      console.log('server change, server', server)
      for (const s of this.list) {
        if (s.basic.host === server.basic.host && s.basic.port === server.basic.port) {
          s.basic.admin_account_name = server.basic.admin_account_name
          s.cpu_mem_processes_usage_info = server.cpu_mem_processes_usage_info || s.cpu_mem_processes_usage_info
        }
      }
    },
    handleRegisterServerButtonClick() {
      this.registerServerModel = {
        host: '0.0.0.0',
        port: 22,
        admin_account_name: '',
        admin_account_pwd: '',
        os_type_label: 'linux'
      }
      console.log('register server clicked')
      this.registerServerFormVisible = true
      this.$nextTick(() => {
        this.$refs['dataForm'].clearValidate()
      })
    },
    handleRegisterServerConnectionTest() {
      this.$refs['dataForm'].validate((valid) => {
        console.log('handleRegisterServerConnectionTest, valid, this.temp', valid, this.registerServerModel)
        if (!valid) {
          return
        }
        const os_type = osTypesLabel2Value[this.registerServerModel.os_type_label] || 'os_type_linux'
        this.registerServerConnectionTestLoading = true
        return connectionTest(this.registerServerModel.host,
          this.registerServerModel.port,
          os_type,
          this.registerServerModel.admin_account_name,
          this.registerServerModel.admin_account_pwd).then((res) => {
          console.log('handleRegisterServerConnectionTest response', res)
          if (res.data.connected === true) {
            this.$message.success('连接成功！')
          } else {
            this.$message.error(`连接失败！原因：${res.data.cause}`)
          }
        }).finally(() => {
          this.registerServerConnectionTestLoading = false
        })
      })
    },
    handleRegisterServerConfirm() {
      this.$refs['dataForm'].validate((valid) => {
        console.log('handleRegisterServerConfirm, valid', valid, this.registerServerModel)
        if (!valid) {
          return
        }
        const os_type = osTypesLabel2Value[this.registerServerModel.os_type_label] || 'os_type_linux'
        this.registerServerConfirmLoading = true
        const _this = this
        return createServer(this.registerServerModel.host,
          this.registerServerModel.port,
          os_type,
          this.registerServerModel.admin_account_name,
          this.registerServerModel.admin_account_pwd).then(() => {
          _this.$message.success('注册成功！')
          this.registerServerFormVisible = false
        }).catch(err => {
          console.log('handleRegisterServerConfirm createServer err', err)
        }).finally(() => {
          this.registerServerConfirmLoading = false
          return this.getServers(this.defaultGetServerParams)
        })
      })
    },
    handleServerDelete() {
      this.getServers(this.defaultGetServerParams)
    }
  }
}
</script>

<style scoped>
.servers-collapse-title {
  display: inline-flex;
  vertical-align: middle;
  align-content: flex-start;
  margin: 0 20px 0 0;
  width: 100%;
}

.servers-collapse-item {
  border-bottom: 1px solid #EEEEEE;
  box-sizing: border-box;
}

.servers-collapse-title:hover {
  background-color: #EEEEEE;
}
</style>
