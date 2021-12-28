<template>
  <div class="app-container">
    <div class="filter-container">
      <el-input v-model="listQuery.searchKeyword" placeholder="关键词" style="width: 200px;" class="filter-item" @keyup.enter.native="handleFilter" />
      <el-button style="margin-left: 20px" v-waves class="filter-item" type="primary" icon="el-icon-search" @click="handleFilter">
        搜索
      </el-button>
    </div>

    <el-table
      :key="tableKey"
      v-loading="listLoading"
      :data="list"
      border
      fit
      highlight-current-row
      style="width: 100%;"
    >
      <el-table-column label="ID" prop="id" align="center" width="80">
        <template slot-scope="{row}">
          <span>{{ row.id }}</span>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="150px" align="center">
        <template slot-scope="{row}">
          <span>{{ row.created_at | parseTime('{y}-{m}-{d} {h}:{i}') }}</span>
        </template>
      </el-table-column>
      <el-table-column label="用户名" min-width="150px">
        <template slot-scope="{row}">
          <span class="link-type" @click="handleUpdate(row)">{{ row.name }}</span>
        </template>
      </el-table-column>
      <el-table-column label="管理员" class-name="status-col" width="100">
        <template slot-scope="{row}">
          <el-tag :type="row.admin | isAdminFilter">
            {{ row.admin ? '是' : '否' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center" width="230" class-name="small-padding fixed-width">
        <!--        <template slot-scope="{row,$index}">-->
        <template slot-scope="{row}">
          <el-button type="primary" size="mini" :disabled="!checkPermission(['admin']) && row.name !== $store.getters.name" @click="handleUpdate(row)">
            编辑信息
          </el-button>
          <el-button v-if="row.admin" :loading="updateAdminButtonLoading" :disabled="!checkPermission(['admin'])" size="mini" @click="handleUpdateAdmin(row,false)">
            取消管理员
          </el-button>
          <el-button v-if="!row.admin" :loading="updateAdminButtonLoading" :disabled="!checkPermission(['admin'])" size="mini" type="success" @click="handleUpdateAdmin(row, true)">
            设置管理员
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <pagination v-show="total>0" :total="total" :page.sync="listQuery.page" :limit.sync="listQuery.limit" @pagination="getUsers" />

    <el-dialog :title="textMap[dialogStatus]" :visible.sync="dialogFormVisible">
      <el-form ref="dataForm" :rules="rules" :model="temp" label-position="left" label-width="90px" style="width: 400px; margin-left:50px;">
        <el-form-item label="用户名" prop="name">
          <el-input v-model="temp.name" disabled />
        </el-form-item>
        <el-form-item label="管理员" prop="admin">
          <el-radio v-model="temp.admin" :disabled="!checkPermission(['admin'])" :label="true">是</el-radio>
          <el-radio v-model="temp.admin" :disabled="!checkPermission(['admin'])" :label="false">否</el-radio>
        </el-form-item>
        <el-form-item label="密码" prop="pwd">
          <el-input v-model="temp.pwd" placeholder="输入密码" show-password />
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="dialogFormVisible = false">
          Cancel
        </el-button>
        <el-button type="primary" @click="updateData()">
          Confirm
        </el-button>
      </div>
    </el-dialog>

    <el-dialog :visible.sync="dialogPvVisible" title="Reading statistics">
      <el-table :data="pvData" border fit highlight-current-row style="width: 100%">
        <el-table-column prop="key" label="Channel" />
        <el-table-column prop="pv" label="Pv" />
      </el-table>
      <span slot="footer" class="dialog-footer">
        <el-button type="primary" @click="dialogPvVisible = false">Confirm</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { getList, getInfo, update } from '@/api/user_mgr'
import waves from '@/directive/waves' // waves directive
import Pagination from '@/components/Pagination' // secondary package based on el-pagination
import checkPermission from '@/utils/permission'

export default {
  name: 'Users',
  components: { Pagination },
  directives: { waves },
  filters: {
    isAdminFilter(isAdmin) {
      return isAdmin ? 'success' : 'info'
    }
  },
  data() {
    const pwdReg = /(?!^(\d*|[a-zA-Z]*|[~!@#$%^&*?]*)$)^[\w~!@#$%^&*?]{6,12}$/
    const validateNewPwd = (rule, value, callback) => {
      if (!pwdReg.test(value)) {
        callback(new Error('密码应是6-12位数字、字母的混合！'))
      } else {
        callback()
      }
    }

    const nameReg = /[0-9A-Za-z_]{1,20}/
    const validateName = (rule, value, callback) => {
      if (!nameReg.test(value)) {
        callback(new Error('用户名仅支持数字，字母，以及下划线的组合，并且不可超过20位！'))
      } else {
        callback()
      }
    }

    return {
      tableKey: 0,
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
      sortOptions: [{ label: 'ID Ascending', key: '+id' }, { label: 'ID Descending', key: '-id' }],
      statusOptions: ['published', 'draft', 'deleted'],
      showReviewer: false,
      temp: {
        id: undefined,
        name: '',
        pwd: '',
        admin: undefined
      },
      dialogFormVisible: false,
      dialogStatus: '',
      textMap: {
        update: 'Edit',
        create: 'Create'
      },
      dialogPvVisible: false,
      pvData: [],
      rules: {
        pwd: [
          { required: true, message: '密码不能为空', trigger: 'change' },
          { validator: validateNewPwd, trigger: 'change' }
        ],
        name: [
          { required: true, message: '用户名不能为空', trigger: 'change' },
          { validator: validateName, trigger: 'change' }
        ],
        admin: [
          { required: true, message: '管理员选项必须选择', trigger: 'change' }
        ]
      },
      downloadLoading: false,
      updateAdminButtonLoading: false
    }
  },
  created() {
    this.getUsers()
  },
  methods: {
    checkPermission,
    updateUser(userID, data) {
      this.listLoading = true
      return update(userID, data).then(res => {
        for (const user of this.list) {
          if (user.id === +userID) {
            user.name = data.name || user.name
            user.pwd = data.pwd || user.pwd
            if (typeof data.admin !== 'undefined') {
              user.admin = data.admin
            }
            break
          }
        }
      }).finally(() => {
        this.listLoading = false
      })
    },
    getUsers() {
      this.listLoading = true
      const from = +(this.listQuery.page - 1) * (+this.listQuery.limit)
      const size = +this.listQuery.limit
      console.log('getUsers, from, size, searchKeyword', from, size, this.listQuery.searchKeyword)
      return getList({
        from: from,
        size: size,
        search_keyword: this.listQuery.searchKeyword
      }).then(response => {
        this.list = response.data.infos
        this.total = response.data.total_count
      }).finally(() => {
        this.listLoading = false
      })
    },
    getUserInfo(userID) {
      const res = getInfo(userID)
      console.log('getUserInfo res', res)
    },
    handleFilter() {
      this.listQuery.page = 1
      this.getUsers()
    },
    handleUpdateAdmin(row, switch2IsAdmin) {
      this.updateAdminButtonLoading = true
      return this.updateUser(row.id, {
        admin: switch2IsAdmin
      }).then(res => {
        console.log('handleUpdateAdmin update success, res', res)
        this.$message({
          message: '成功！',
          type: 'success'
        })
      }).catch(err => {
        console.log('更新管理员失败, err', err)
      }).finally(() => {
        if (row.name === this.$store.getters.name && switch2IsAdmin === false) {
          this.$store.dispatch('user/changeRoles', ['editor'])
        }
        this.updateAdminButtonLoading = false
      })
    },
    sortChange(data) {
      const { prop, order } = data
      if (prop === 'id') {
        this.sortByID(order)
      }
    },
    sortByID(order) {
      if (order === 'ascending') {
        this.listQuery.sort = '+id'
      } else {
        this.listQuery.sort = '-id'
      }
      this.handleFilter()
    },
    handleUpdate(row) {
      if (!checkPermission(['admin']) && row.name !== this.$store.getters.name) {
        return
      }
      this.temp = Object.assign({}, row) // copy obj
      this.dialogStatus = 'update'
      this.dialogFormVisible = true
      this.$nextTick(() => {
        this.$refs['dataForm'].clearValidate()
      })
    },
    updateData() {
      this.$refs['dataForm'].validate((valid) => {
        console.log('users updateData, valid, this.temp', valid, this.temp)
        if (!valid) {
          return
        }
        const tempData = Object.assign({}, this.temp)
        this.updateUser(tempData.id, tempData).then(_ => {
          const index = this.list.findIndex(v => v.id === this.temp.id)
          this.list.splice(index, 1, this.temp)
          this.dialogFormVisible = false
          this.$notify({
            title: '更新成功！',
            message: '用户信息更新成功',
            type: 'success',
            duration: 2000
          })
        }).catch(err => {
          console.log('users update data failed, err', err)
          this.dialogFormVisible = false
          this.$notify({
            title: 'Failed',
            message: 'Update Failed',
            type: 'error',
            duration: 2000
          })
        })
      })
    },
    getSortClass: function(key) {
      const sort = this.listQuery.sort
      return sort === `+${key}` ? 'ascending' : 'descending'
    }
  }
}
</script>
