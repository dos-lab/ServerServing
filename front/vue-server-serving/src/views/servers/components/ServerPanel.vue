<template>
  <el-container v-loading="panelLoading" style="min-height: 500px">
    <el-main>
      <el-row type="flex" align="middle" justify="space-between">
        <el-col :span="12">
          <div style="font-size: 24px">
            IP: {{ server.basic ? server.basic.host : '未知' }} 端口：{{ server.basic ? server.basic.port : '未知' }}
          </div>
        </el-col>
        <el-col :span="6">
          <div style="display:flex; align-items: center; justify-content: end">
            <el-button style="margin-left: 10px" type="primary" @click="handleOriginalServerJsonButton">
              查看服务器原始json数据
            </el-button>
            <!--            <el-button style="margin-left: 10px" type="primary" @click="handleServerUpdateButton">-->
            <!--              编辑服务器-->
            <!--            </el-button>-->
            <el-popover
              v-model="deleteServerPopupVisible"
              placement="top"
              style="display: flex; align-items: center; "
            >
              <p>确定删除吗？</p>
              <div style="text-align: right; margin: 0">
                <el-button size="mini" type="text" @click="deleteServerPopupVisible = false">取消</el-button>
                <el-button type="primary" :loading="serverDeleting" size="mini" @click="handleDeleteServerButton">确定</el-button>
              </div>
              <el-button slot="reference" type="danger" style="margin-left: 10px">
                删除服务器
              </el-button>
            </el-popover>
          </div>
        </el-col>
      </el-row>
      <el-row style="margin-top: 20px" type="flex" justify="start">
        <el-col :span="24">
          <div style="display:flex; align-items: center; justify-content: end">
            <el-switch
              v-model="autoRefreshEnabled"
              active-text="自动刷新"
              @change="handleAutoRefreshEnableChanged"
            />
            <el-input-number v-model="autoRefreshInterval" style="margin-left: 20px;" :disabled="!autoRefreshEnabled" size="mini" controls-position="right" :min="5" :max="300" @change="handleAutoRefreshIntervalChange" />
            <div style="margin-left: 5px">秒</div>
            <div style="margin-left: 20px">
              上次自动刷新：{{ lastAutoRefreshing }}
            </div>
            <div style="margin-left: 20px;">
              <el-checkbox-group v-model="autoRefreshOptGroup" :disabled="!autoRefreshEnabled">
                <el-checkbox-button v-for="opt in autoRefreshOptGroupLabels" :key="opt" :label="opt">{{ opt }}</el-checkbox-button>
              </el-checkbox-group>
            </div>
          </div>
        </el-col>
      </el-row>
      <el-row :gutter="8" style="margin-top: 20px">
        <el-col :span="12">
          <el-container>
            <el-header style="height: 50px;">
              <div class="el-header" style="font-size: 24px;">
                硬件信息
              </div>
            </el-header>
            <el-row>
              <el-col :span="24">
                <el-table
                  :key="hardwareTableKey"
                  :data="hardwareTableData"
                  fit
                  highlight-current-row
                  max-height="500"
                  style="width: 100%;"
                >
                  <el-table-column label="硬件名" align="center" width="180">
                    <template slot-scope="{row}">
                      <span>
                        <el-popover
                          placement="top-start"
                          title="原始输出"
                          trigger="hover"
                          :content="row.output"
                        >
                          <span slot="reference"> {{ row.entry }} </span>
                        </el-popover>
                      </span>
                    </template>
                  </el-table-column>
                  <el-table-column label="描述">
                    <template slot-scope="{row}">
                      <span><div v-html="row.description" /></span>
                    </template>
                  </el-table-column>

                </el-table>
              </el-col>
            </el-row>
          </el-container>
        </el-col>

        <el-col :span="12">
          <el-container>
            <el-header style="height: 50px;">
              <div class="el-header" style="font-size: 24px;">
                硬件使用
              </div>
            </el-header>
            <el-row>
              <el-col :span="24">
                <el-table
                  :key="hardwareUsageTableKey"
                  :data="hardwareUsageTableData"
                  fit
                  highlight-current-row
                  max-height="500"
                  style="width: 100%;"
                >
                  <el-table-column label="硬件名" align="center" width="180">
                    <template slot-scope="{row}">
                      <span> {{ row.entry }} </span>
                    </template>
                  </el-table-column>
                  <el-table-column label="使用情况">
                    <template slot-scope="{row}">
                      <span><div v-html="row.description" /></span>
                    </template>
                  </el-table-column>

                </el-table>
              </el-col>
            </el-row>
          </el-container>
        </el-col>
      </el-row>

      <el-row :gutter="8" style="margin-top: 20px">
        <el-col :span="12">
          <el-container>
            <el-header style="height: 50px;">
              <div class="el-header" style="font-size: 24px; display: flex; align-items: center; justify-content: start">
                <el-popover
                  placement="top-start"
                  title="原始输出"
                  trigger="hover"
                  :content="server.account_infos ? server.account_infos.output : '未知'"
                >
                  <span slot="reference"> 账户信息 </span>
                </el-popover>
                <el-button style="margin-left: 20px" type="primary" @click="handleCreateAccountButton">
                  创建账户
                </el-button>
              </div>
            </el-header>
            <el-main style="padding-top: 0">
              <el-table
                :key="accountTableKey"
                v-loading="accountTableLoading"
                :data="server.account_infos ? server.account_infos.accounts : null"
                fit
                highlight-current-row
                max-height="500"
                style="width: 100%;"
              >
                <el-table-column label="序号" type="index" align="center">
                </el-table-column>
                <el-table-column label="账户" min-width="100px" align="center">
                  <template slot-scope="{row}">
                    <span v-if="row.not_exists_in_server">
                      <el-tooltip class="item" effect="dark" content="该账户不存在于服务器，可进行恢复" placement="top">
                        <div style="color: red">
                          {{ row.name }}
                        </div>
                      </el-tooltip>
                    </span>
                    <span v-else>{{ row.name }}</span>
                  </template>
                </el-table-column>
                <el-table-column label="密码" min-width="150px">
                  <template slot="header">
                    <el-tooltip class="item" effect="dark" content="密码无法直接从服务器获取，该列仅能展示在ServerServing创建的账户密码" placement="top">
                      <span>
                        密码
                      </span>
                    </el-tooltip>
                  </template>
                  <template slot-scope="{row}">
                    <span>{{ row.pwd && row.pwd.length > 0 ? row.pwd : '未知' }}</span>
                  </template>
                </el-table-column>
                <el-table-column label="操作" min-width="150px">
                  <template slot-scope="{row}">
                    <el-button v-if="row.not_exists_in_server === false" size="small" type="danger" @click="handleDeleteAccountDialogButton(row.name)">
                      删除
                    </el-button>
                    <el-button v-if="row.not_exists_in_server === true" size="small" type="primary" @click="handleRecoverAccountDialogButton(row.name, row.pwd)">
                      恢复
                    </el-button>
                  </template>
                </el-table-column>

              </el-table>
            </el-main>

          </el-container>
        </el-col>
        <el-col :span="12">
          <el-container>
            <el-header style="height: 50px;">
              <div class="el-header" style="font-size: 24px;">
                <el-popover
                  placement="top-start"
                  title="原始输出"
                  trigger="hover"
                  :content="server.cpu_mem_processes_usage_info ? server.cpu_mem_processes_usage_info.output : '未知'"
                >
                  <span slot="reference"> 进程信息 </span>
                </el-popover>
              </div>
            </el-header>
            <el-main style="padding-top: 0">
              <el-table
                :key="processTableKey"
                :data="server.cpu_mem_processes_usage_info ? server.cpu_mem_processes_usage_info.process_infos : null"
                fit
                highlight-current-row
                max-height="500"
                style="width: 100%;"
              >
                <el-table-column label="pid" align="center">
                  <template slot-scope="{row}">
                    <span>{{ row.pid }}</span>
                  </template>
                </el-table-column>
                <el-table-column label="账户">
                  <template slot-scope="{row}">
                    <span>{{ row.owner_account_name }}</span>
                  </template>
                </el-table-column>
                <el-table-column label="CPU(%)">
                  <template slot-scope="{row}">
                    <span>{{ row.cpu_usage.toFixed(2) }}</span>
                  </template>
                </el-table-column>
                <el-table-column label="MEM(%)">
                  <template slot-scope="{row}">
                    <span>{{ row.mem_usage.toFixed(2) }}</span>
                  </template>
                </el-table-column>
                <el-table-column label="命令">
                  <template slot-scope="{row}">
                    <span>{{ row.command }}</span>
                  </template>
                </el-table-column>

              </el-table>
            </el-main>

          </el-container>
        </el-col>
      </el-row>

      <el-row :gutter="8" style="margin-top: 20px">
        <el-col :span="12">
          <el-container>
            <el-header style="height: 50px;">
              <div class="el-header" style="font-size: 24px;">
                <el-popover
                  placement="top-start"
                  title="原始输出"
                  trigger="hover"
                  :content="server.remote_accessing_usage_info ? server.remote_accessing_usage_info.output : '未知'"
                >
                  <span slot="reference"> 正在远程登录这台服务器的账户 </span>
                </el-popover>
              </div>
            </el-header>
            <el-main style="padding-top: 0">
              <el-table
                :key="remoteAccessingTableKey"
                :data="server.remote_accessing_usage_info ? server.remote_accessing_usage_info.infos : null"
                fit
                highlight-current-row
                max-height="500"
                style="width: 100%;"
              >
                <el-table-column label="账户" width="300" align="center">
                  <template slot-scope="{row}">
                    <span>{{ row.account_name }}</span>
                  </template>
                </el-table-column>
                <el-table-column label="当前执行指令" min-width="150px">
                  <template slot-scope="{row}">
                    <span>{{ row.what || '未知' }}</span>
                  </template>
                </el-table-column>
              </el-table>
            </el-main>
          </el-container>
        </el-col>
        <el-col :span="12">
          <el-container>
            <el-header style="height: 50px;">
              <div class="el-header" style="font-size: 24px;">
                GPU使用信息
              </div>
            </el-header>
            <el-main style="padding-top: 0">
              <p>
                {{ server.server_gpu_usage_info ? server.server_gpu_usage_info.output : '未知' }}
              </p>
            </el-main>
          </el-container>
        </el-col>
      </el-row>

    </el-main>

    <el-dialog title="删除账户" :visible.sync="deleteAccountDialogVisible">
      <el-form>
        <el-form-item>
          <el-switch
            v-model="deleteAccountModel.doBackup"
            :disabled="!isAble2deleteAccountWithBackup.able"
            active-text="备份用户文件夹"/>
          <div v-if="isAble2deleteAccountWithBackup.able">
            备份至文件夹：{{ isAble2deleteAccountWithBackup.targetDir }}
          </div>
          <div v-else>
            当前不能备份到：{{ isAble2deleteAccountWithBackup.targetDir }}，原因为：{{ isAble2deleteAccountWithBackup.desc }}
          </div>
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button type="danger" :loading="deleteAccountConfirmLoading" @click="handleDeleteAccount">
          删除
        </el-button>
      </div>
    </el-dialog>

    <el-dialog title="恢复账户" :visible.sync="recoverAccountDialogVisible">
      <el-form ref="recoverAccountForm" :rules="recoverAccountRules" :model="recoverAccountModel">
        <el-form-item label="密码" prop="account_pwd">
          <el-input v-model="recoverAccountModel.account_pwd" placeholder="输入密码" show-password />
        </el-form-item>
        <el-form-item>
          <el-switch
            v-model="recoverAccountModel.do_backup"
            :disabled="!isAble2recoverAccountWithBackup.able"
            active-text="恢复备份用户文件夹"/>
          <div v-if="isAble2recoverAccountWithBackup.able">
            恢复备份文件夹：{{ isAble2recoverAccountWithBackup.targetDir }}
          </div>
          <div v-else>
            当前不能恢复备份：{{ isAble2recoverAccountWithBackup.targetDir }}，原因为：{{ isAble2recoverAccountWithBackup.desc }}
          </div>
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button type="primary" :loading="recoverAccountConfirmLoading" @click="handleRecoverAccount">
          恢复
        </el-button>
      </div>
    </el-dialog>

    <el-dialog title="创建账户" :visible.sync="createAccountFormVisible">
      <el-form ref="createAccountForm" label-position="top" :rules="createAccountRules" :model="createAccountModel">
        <el-form-item label="账户名称" prop="account_name">
          <el-input v-model="createAccountModel.account_name" />
        </el-form-item>
        <el-form-item label="密码" prop="account_pwd">
          <el-input v-model="createAccountModel.account_pwd" />
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button type="primary" :loading="createAccountConfirmLoading" @click="handleCreateAccountConfirm">
          创建
        </el-button>
      </div>
    </el-dialog>
    <el-drawer
      title="服务器原始json数据"
      :with-header="false"
      size="50%"
      :visible.sync="drawerVisible"
    >
      <json-editor ref="jsonEditor" v-model="jsonEditorValue" class="overflowAuto" />
    </el-drawer>
  </el-container>
</template>

<script>

// import Mock from 'mockjs'
import { getInfo, deleteServer, createAccount, deleteAccount, recoverAccount } from '@/api/server'
import JsonEditor from '@/components/JsonEditor'

const autoRefreshLabels2Attr = {
  '账户信息': 'with_accounts',
  '硬件使用与进程信息': 'with_cmp_usages',
  'GPU信息': 'with_gpu_usages',
  '远程访问信息': 'with_remote_access_usages'
}

const osTypesValue2Label = { 'os_type_linux': 'linux', 'os_type_windows_server': 'windows server' }

export default {
  components: { JsonEditor },
  props: {
    host: {
      type: String,
      default: () => '123.123.123.123'
    },
    port: {
      type: Number,
      default: () => 22
    }
  },
  data() {
    const nameReg = /^[a-zA-Z][0-9A-Za-z_]{2,14}$/
    const validateName = (rule, value, callback) => {
      if (!nameReg.test(value)) {
        callback(new Error('账户名仅支持字母开头的数字，字母，以及下划线的组合，并且大于等于3位，不可超过15位！'))
      } else {
        callback()
      }
    }
    const pwdReg = /^[a-zA-Z][0-9a-zA-Z~!@#$%^&*?]{5,14}$/
    const validateNewPwd = (rule, value, callback) => {
      if (!pwdReg.test(value)) {
        callback(new Error('密码应是6-15位字母开头的，数字、字母、特殊符号的混合！'))
      } else {
        callback()
      }
    }
    return {
      panelLoading: false,
      server: {
        basic: null,
        access_failed_info: null,
        account_infos: null,
        hardware_info: null,
        cpu_mem_processes_usage_info: null,
        remote_accessing_usage_info: null,
        server_gpu_usage_info: null
      },
      accountTableKey: 1,
      processTableKey: 2,
      hardwareTableKey: 3,
      hardwareUsageTableKey: 4,
      remoteAccessingTableKey: 5,
      defaultRefreshOpts: {
        with_accounts: true,
        with_cmp_usages: true,
        with_gpu_usages: true,
        with_hardware_info: true,
        with_remote_access_usages: true
      },
      autoRefreshOptGroup: [],
      autoRefreshOptGroupLabels: Object.keys(autoRefreshLabels2Attr),
      autoRefreshEnabled: false,
      autoRefreshInterval: 20,
      autoRefreshIntervalFunc: null,
      autoRefreshing: false,
      lastAutoRefreshing: '无',
      drawerVisible: false,
      jsonEditorValue: {},
      deleteServerPopupVisible: false,
      serverDeleting: false,
      serverUpdateModel: {
        host: '0.0.0.0',
        port: 22,
        admin_account_name: '',
        admin_account_pwd: '',
        os_type_label: 'linux'
      },
      serverUpdateRules: {
        admin_account_name: [
          { required: true, message: '管理员账户名不能为空', trigger: 'change' },
          { validator: validateName, trigger: 'change' }
        ],
        admin_account_pwd: [
          { required: true, message: '管理员密码不能为空', trigger: 'change' },
          { validator: validateNewPwd, trigger: 'change' }
        ]
      },
      osTypesValue2Label: osTypesValue2Label,
      serverUpdateFormVisible: false,
      serverUpdateConnectionTestLoading: false,
      serverUpdateConfirmLoading: false,
      accountTableLoading: false,
      createAccountModel: {
        account_name: '',
        account_pwd: ''
      },
      createAccountRules: {
        account_name: [
          { required: true, message: '账户名不能为空', trigger: 'change' },
          { validator: validateName, trigger: 'change' }
        ],
        account_pwd: [
          { required: true, message: '密码不能为空', trigger: 'change' },
          { validator: validateNewPwd, trigger: 'change' }
        ]
      },
      createAccountFormVisible: false,
      createAccountConfirmLoading: false,
      deleteAccountDialogVisible: false,
      deleteAccountModel: {
        account_name: '',
        doBackup: false
      },
      deleteAccountConfirmLoading: false,
      recoverAccountDialogVisible: false,
      recoverAccountRules: {
        account_pwd: [
          { required: true, message: '密码不能为空', trigger: 'change' },
          { validator: validateNewPwd, trigger: 'change' }
        ]
      },
      recoverAccountConfirmLoading: false,
      recoverAccountModel: {
        account_name: '',
        account_pwd: '',
        do_backup: false
      }
    }
  },
  computed: {
    hardwareTableData: function() {
      const data = []
      if (!this.server.hardware_info) {
        return data
      }
      const cpu_info = this.server.hardware_info.cpu_hardware_info.info
      data.push({
        output: this.server.hardware_info.cpu_hardware_info.output,
        entry: 'CPU',
        description: `架构：${cpu_info.architecture} <br> 类型：${cpu_info.model_name} <br> 核心数：${cpu_info.cores} <br> 每核心线程：${cpu_info.threads_per_core}`
      })
      const gpu_info = this.server.hardware_info.gpu_hardware_infos
      const gpus = []
      if (gpu_info.infos) {
        for (const gpu of gpu_info.infos) {
          gpus.push(`${gpu.product}`)
        }
      }
      data.push({
        output: this.server.hardware_info.gpu_hardware_infos.output,
        entry: 'GPU',
        description: gpus.length > 0 ? `型号：[${gpus.join(', ')}]` : '未知'
      })
      return data
    },
    hardwareUsageTableData: function() {
      const data = []
      if (this.server.cpu_mem_processes_usage_info === null) {
        return data
      }
      const cpu_mem_process_usage_info = this.server.cpu_mem_processes_usage_info
      if (cpu_mem_process_usage_info.failed_info !== null) {
        return data
      }
      const cpu_usage_desc = (function() {
        const cpu_usage = cpu_mem_process_usage_info.cpu_mem_usage.user_cpu_usage
        if (cpu_usage === null) {
          return '未知'
        }
        return `${cpu_usage.toFixed(2)}`
      }())
      data.push({
        entry: 'CPU使用率(%)',
        description: cpu_usage_desc
      })
      const mem_desc = (function() {
        const mem_usage = cpu_mem_process_usage_info.cpu_mem_usage.mem_usage
        const mem_total = cpu_mem_process_usage_info.cpu_mem_usage.mem_total
        if (mem_usage === null || mem_total == null) {
          return '未知'
        }
        return `${mem_usage.toFixed(2)} 共(${mem_total})`
      }())
      data.push({
        entry: '内存使用率(%)',
        description: mem_desc
      })
      return data
    },
    isAble2deleteAccountWithBackup: function() {
      const account_name = this.deleteAccountModel.account_name
      console.log('isAble2deleteAccountWithBackup account_name', account_name)
      if (!this.server.account_infos) {
        return {
          able: false,
          targetDir: null,
          desc: '账户信息获取失败！'
        }
      }
      for (const acc of this.server.account_infos.accounts) {
        if (account_name === acc.name) {
          if (!acc.backup_dir_info) {
            // 不清楚该账户的备份文件夹情况时，不允许进行备份。
            return {
              able: false,
              targetDir: null,
              desc: '服务器上的账户备份文件夹信息未知，不允许进行备份！'
            }
          }
          if (acc.backup_dir_info.path_exists) {
            return {
              able: false,
              targetDir: acc.backup_dir_info.backup_dir,
              desc: '该账户对应的备份文件夹路径已存在，不允许进行备份！'
            }
          }
          return {
            able: true,
            targetDir: acc.backup_dir_info.backup_dir,
            desc: ''
          }
        }
      }
      return {
        able: false,
        targetDir: null,
        desc: '未找到该账户信息！'
      }
    },
    isAble2recoverAccountWithBackup: function() {
      const account_name = this.recoverAccountModel.account_name
      console.log('isAble2recoverAccountWithBackup account_name', account_name)
      if (!this.server.account_infos) {
        return {
          able: false,
          targetDir: null,
          desc: '账户信息获取失败！'
        }
      }
      for (const acc of this.server.account_infos.accounts) {
        if (account_name === acc.name) {
          if (!acc.backup_dir_info) {
            // 不清楚该账户的备份文件夹情况时，不允许恢复备份。
            return {
              able: false,
              targetDir: null,
              desc: '服务器上的账户备份文件夹信息未知，不允许恢复备份！'
            }
          }
          if (!acc.backup_dir_info.path_exists) {
            return {
              able: false,
              targetDir: acc.backup_dir_info.backup_dir,
              desc: '该账户对应的备份文件夹路径不存在，无法恢复备份！'
            }
          }
          if (!acc.backup_dir_info.dir_exists) {
            return {
              able: false,
              targetDir: acc.backup_dir_info.backup_dir,
              desc: '该账户对应的备份文件夹路径存在，但并不是文件夹，无法恢复备份！'
            }
          }
          return {
            able: true,
            targetDir: acc.backup_dir_info.backup_dir,
            desc: ''
          }
        }
      }
      return {
        able: false,
        targetDir: null,
        desc: '未找到该账户信息！'
      }
    }
  },
  created() {
    this.panelLoading = true
    this.refresh_server(this.defaultRefreshOpts).finally(() => {
      setTimeout(() => {
        this.panelLoading = false
      }, 1000)
    })
  },
  methods: {
    extractProcessInfos() {
      if (this.server.cpu_mem_processes_usage_info &&
        !this.server.cpu_mem_processes_usage_info.failed_info &&
        this.server.cpu_mem_processes_usage_info.process_infos) {
        return this.server.cpu_mem_processes_usage_info.process_infos
      }
      return null
    },
    refresh_server(opts) {
      console.log('refresh_server, opts', opts)
      return getInfo(this.host, this.port, opts).then(res => {
        console.log('refresh_server getInfo', res)
        this.$emit('server-change', res.data)
        this.server.basic = res.data.basic || this.server.basic
        this.server.access_failed_info = res.data.access_failed_info || this.server.access_failed_info
        this.server.account_infos = res.data.account_infos || this.server.account_infos
        this.server.hardware_info = res.data.hardware_info || this.server.hardware_info
        this.server.cpu_mem_processes_usage_info = res.data.cpu_mem_processes_usage_info || this.server.cpu_mem_processes_usage_info
        this.server.remote_accessing_usage_info = res.data.remote_accessing_usage_info || this.server.remote_accessing_usage_info
        this.server.server_gpu_usage_info = res.data.server_gpu_usage_info || this.server.server_gpu_usage_info
      })
    },
    handleAutoRefreshIntervalChange(val) {
      console.log('handleAutoRefreshIntervalChange val', val)
      if (this.autoRefreshIntervalFunc !== null) {
        clearInterval(this.autoRefreshIntervalFunc)
      }
      if (!this.autoRefreshEnabled) {
        return
      }
      this.autoRefreshIntervalFunc = setInterval(() => {
        if (this.autoRefreshing) {
          return
        }
        const opts = {}
        console.log('handleAutoRefreshIntervalChange, autoRefreshOptGroup', this.autoRefreshOptGroup)
        for (const opt of this.autoRefreshOptGroupLabels) {
          const attr = autoRefreshLabels2Attr[opt]
          opts[attr] = false
        }
        for (const opt of this.autoRefreshOptGroup) {
          const attr = autoRefreshLabels2Attr[opt]
          opts[attr] = true
        }
        console.log('handleAutoRefreshIntervalChange, opts', opts)
        this.autoRefreshing = true
        this.refresh_server(opts).finally(() => {
          this.autoRefreshing = false
          const dt = new Date()
          const prefixZero = function(item) {
            item = `${item}`
            return item.length < 2 ? '0' + item : item
          }
          const hours = prefixZero(dt.getHours())
          const minutes = prefixZero(dt.getMinutes())
          const secs = prefixZero(dt.getSeconds())
          this.lastAutoRefreshing = `${hours}:${minutes}:${secs}`
        })
      }, val * 1000)
    },
    handleAutoRefreshEnableChanged(val) {
      this.handleAutoRefreshIntervalChange(this.autoRefreshInterval)
    },
    handleOriginalServerJsonButton() {
      const basic = Object.assign({}, this.server.basic)
      const access_failed_info = Object.assign({}, this.server.access_failed_info)
      const account_infos = Object.assign({}, this.server.account_infos)
      const hardware_info = Object.assign({}, this.server.hardware_info)
      const cpu_mem_processes_usage_info = Object.assign({}, this.server.cpu_mem_processes_usage_info)
      const remote_accessing_usage_info = Object.assign({}, this.server.remote_accessing_usage_info)
      const server_gpu_usage_info = Object.assign({}, this.server.server_gpu_usage_info)
      this.jsonEditorValue = {
        basic: basic,
        account_infos: account_infos,
        access_failed_info: access_failed_info,
        hardware_info: hardware_info,
        cpu_mem_processes_usage_info: cpu_mem_processes_usage_info,
        remote_accessing_usage_info: remote_accessing_usage_info,
        server_gpu_usage_info: server_gpu_usage_info
      }
      this.drawerVisible = true
    },
    handleDeleteServerButton() {
      this.serverDeleting = true
      return deleteServer(this.host, this.port).then(() => {
        this.$message.success('删除服务器成功！')
        this.$emit('server-delete')
        this.deleteServerPopupVisible = false
      }).catch(err => {
        console.log('delete server failed, err', err)
      }).finally(() => {
        this.serverDeleting = false
      })
    },
    withRecoverAutoRefresh(func) {
      const wasAutoRefreshing = this.pauseAutoRefresh()
      return func().finally(() => {
        this.recoverAutoRefresh(wasAutoRefreshing)
      })
    },
    handleCreateAccountConfirm() {
      this.$refs['createAccountForm'].validate((valid) => {
        if (!valid) {
          console.log('handleCreateAccountConfirm not valid')
          return null
        }
        this.createAccountConfirmLoading = true
        const _this = this
        return _this.withRecoverAutoRefresh(function() {
          return createAccount(_this.host, _this.port, _this.createAccountModel.account_name, _this.createAccountModel.account_pwd).then((res) => {
            console.log('create account, res', res)
            _this.$message.success('创建账户成功！')
            return _this.refreshAccounts()
          }).catch((err) => {
            console.log('create account err', err)
          }).finally(() => {
            _this.createAccountConfirmLoading = false
            _this.createAccountFormVisible = false
          })
        })
      })
    },
    handleCreateAccountButton() {
      this.createAccountFormVisible = true
      this.createAccountModel = {
        account_name: '',
        account_pwd: ''
      }
    },
    refreshAccounts() {
      this.accountTableLoading = true
      return this.refresh_server({
        with_accounts: true
      }).finally(() => {
        setTimeout(() => {
          this.accountTableLoading = false
        }, 1000)
      })
    },
    pauseAutoRefresh() {
      const wasAutoRefreshing = this.autoRefreshing
      this.autoRefreshing = false
      this.handleAutoRefreshEnableChanged(false)
      return wasAutoRefreshing
    },
    recoverAutoRefresh(wasAutoRefreshing) {
      if (wasAutoRefreshing) {
        this.autoRefreshing = true
        this.handleAutoRefreshEnableChanged(true)
      }
    },
    handleDeleteAccountDialogButton(account_name) {
      this.deleteAccountModel.account_name = account_name
      this.deleteAccountDialogVisible = true
    },
    handleRecoverAccountDialogButton(account_name, account_pwd) {
      this.recoverAccountModel.account_name = account_name
      this.recoverAccountModel.account_pwd = account_pwd
      this.recoverAccountDialogVisible = true
    },
    handleDeleteAccount() {
      console.log('handleDeleteAccountButton deleteAccountModel', this.deleteAccountModel)
      this.deleteAccountConfirmLoading = true
      const _this = this
      return _this.withRecoverAutoRefresh(function() {
        return deleteAccount(_this.host, _this.port, _this.deleteAccountModel.account_name, _this.deleteAccountModel.doBackup).then(() => {
          _this.$message.success('删除账户成功！')
          return _this.refreshAccounts()
        }).catch(err => {
          console.log('删除账户失败', err)
        }).finally(() => {
          console.log('handleDeleteAccount finally')
          _this.deleteAccountConfirmLoading = false
          _this.deleteAccountDialogVisible = false
        })
      })
    },
    handleRecoverAccount() {
      console.log('handleRecoverAccountButton recoverAccountModel', this.recoverAccountModel)
      this.$refs['recoverAccountForm'].validate((valid) => {
        if (!valid) {
          console.log('handleRecoverAccount model not valid')
          return null
        }
        this.recoverAccountConfirmLoading = true
        const _this = this
        return _this.withRecoverAutoRefresh(function() {
          return recoverAccount(_this.host, _this.port, _this.recoverAccountModel.account_name, _this.recoverAccountModel.account_pwd, _this.recoverAccountModel.do_backup).then(() => {
            _this.$message.success('恢复账户成功！')
            return _this.refreshAccounts()
          }).catch(err => {
            console.log('恢复账户失败', err)
          }).finally(() => {
            _this.recoverAccountConfirmLoading = false
            _this.recoverAccountDialogVisible = false
          })
        })
      })
    }
    // handleServerUpdateButton() {
    //   this.serverUpdateModel = {
    //     host: this.server.basic.host,
    //     port: this.server.basic.port,
    //     admin_account_name: this.server.basic.admin_account_name,
    //     admin_account_pwd: this.server.basic.admin_account_pwd,
    //     os_type: this.server.basic.os_type
    //   }
    //   this.serverUpdateFormVisible = true
    // }
    // handleUpdateServerConnectionTest() {
    //   this.$refs['dataForm'].validate((valid) => {
    //     console.log('handleUpdateServerConnectionTest, valid, this.temp', valid, this.serverUpdateModel)
    //     if (!valid) {
    //       return
    //     }
    //     this.serverUpdateConnectionTestLoading = true
    //     return connectionTest(this.server.basic.host, this.server.basic.port, {
    //       os_type: this.server.basic.os_type,
    //       admin_account_name: this.serverUpdateModel.admin_account_name,
    //       admin_account_pwd: this.serverUpdateModel.admin_account_pwd
    //     }).then((res) => {
    //       console.log('handleUpdateServerConnectionTest response', res)
    //       if (res.data.connected === true) {
    //         this.$message.success('连接成功！')
    //       } else {
    //         this.$message.error('连接失败！')
    //       }
    //     }).finally(() => {
    //       this.serverUpdateConnectionTestLoading = false
    //     })
    //   })
    // },
    // handleUpdateServerConfirm() {
    //   this.$refs['dataForm'].validate((valid) => {
    //     console.log('handleUpdateServerConfirm, valid, this.temp', valid, this.serverUpdateModel)
    //     if (!valid) {
    //       return
    //     }
    //
    //   }
    // }
  }
}
</script>

<style lang="scss" scoped>
.box-center {
  margin: 0 auto;
  display: table;
}

.text-muted {
  color: #777;
}

.user-profile {
  .user-name {
    font-weight: bold;
  }

  .box-center {
    padding-top: 10px;
  }

  .user-role {
    padding-top: 10px;
    font-weight: 400;
    font-size: 14px;
  }

  .box-social {
    padding-top: 30px;

    .el-table {
      border-top: 1px solid #dfe6ec;
    }
  }

  .user-follow {
    padding-top: 20px;
  }
}

.user-bio {
  margin-top: 20px;
  color: #606266;

  span {
    padding-left: 4px;
  }

  .user-bio-section {
    font-size: 14px;
    padding: 15px 0;

    .user-bio-section-header {
      border-bottom: 1px solid #dfe6ec;
      padding-bottom: 10px;
      margin-bottom: 10px;
      font-weight: bold;
    }
  }
}

.overflowAuto {
  overflow-y: auto;
  position: absolute;
  width: 100%;
  height: 100%;
}
.overflowAuto::-webkit-scrollbar {
  height: 6px;
  width: 6px;
}
.overflowAuto::-webkit-scrollbar-thumb {
  background: rgb(224, 214, 235);
}
</style>
