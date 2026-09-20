<template>
  <AdminShell title="权限说明" subtitle="角色-入口-数据范围权限矩阵">
    <section class="panel">
      <div class="panel-title">角色权限矩阵</div>
      <el-table :data="rows" border stripe>
        <el-table-column prop="role" label="角色" width="150" />
        <el-table-column prop="entry" label="登录入口" min-width="180" />
        <el-table-column prop="scope" label="数据范围" min-width="220" />
        <el-table-column prop="can" label="可执行操作" min-width="260" />
      </el-table>
    </section>

    <section class="panel">
      <div class="panel-title">边界保护</div>
      <el-alert type="success" :closable="false" title="已实施" description="管理员账号不可停用/改角色；楼栋删除前检查工单引用；移动端按 /m/dorm 与 /m/worker 路由守卫角色；401/403 自动回登录。" />
      <el-alert style="margin-top:10px" type="warning" :closable="false" title="演示环境注意" description="所有演示账号密码为 admin123，正式答辩/部署前请修改 .env 与种子账号。" />
    </section>
  </AdminShell>
</template>

<script setup lang="ts">
import AdminShell from '../components/AdminShell.vue'

const rows = [
  { role: '超级管理员', entry: 'PC /admin/*', scope: '全部数据（工单/人员/楼栋/规则/日志）', can: '管理 CRUD、工单总览、统计、日志' },
  { role: '维修工人', entry: 'H5 /m/worker', scope: '自己的工单、管辖楼栋活动工单', can: '开工/完工/批量完工、2D/3D 地图' },
  { role: '宿管', entry: 'H5 /m/dorm', scope: '绑定楼栋工单', can: '报修、本栋列表、未开工取消' },
]
</script>

<style scoped>
.panel {
  padding: 18px 20px;
  margin-bottom: 16px;
  background: #fff;
  border: 1px solid #eef2f7;
  border-radius: 14px;
}
.panel-title {
  margin-bottom: 12px;
  color: #334155;
  font-size: 14px;
  font-weight: 700;
}
</style>