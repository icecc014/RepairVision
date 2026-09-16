// 从「1 号宿舍楼的自定义布局」生成双端内置标准层常量（frontend/*/src/utils/standardFloor.ts）。
//
// 用法：
//   node scripts/gen-standard-floor.mjs [layout.json]
// 默认读取 .tools/backups/building1_layout_20260916.json
// 重新导出布局：
//   docker exec repairvision-mysql-1 mysql -uroot -p<密码> --default-character-set=utf8mb4 -N \
//     -e "select layout_json from map_db.buildings where id=1;" > layout.json
import { readFileSync, writeFileSync } from 'node:fs'

const src = process.argv[2] || '.tools/backups/building1_layout_20260916.json'
const raw = JSON.parse(readFileSync(src, 'utf8'))

const cols = Number(raw.cols) || 7
const rows = Number(raw.rows) || 10
const kindOf = (v) => String(v || 'custom')
const blocks = (raw.blocks || []).map((b) => {
  const block = {
    id: String(b.id || `${kindOf(b.kind)}-${b.row}-${b.col}`),
    kind: kindOf(b.kind),
    row: Number(b.row) || 0,
    col: Number(b.col) || 0,
    rowSpan: Number(b.rowSpan) || 1,
    colSpan: Number(b.colSpan) || 1,
  }
  if (b.label) block.label = String(b.label)
  return block
})

const payload = JSON.stringify({ version: 2, cols, rows, blocks })
const stamp = new Date().toISOString().slice(0, 19).replace('T', ' ')
const content = [
  '// 本文件由 scripts/gen-standard-floor.mjs 自动生成，请勿手工修改。',
  `// 数据来源：1 号宿舍楼（map_db.buildings id=${1}）的自定义布局，生成时间 ${stamp}`,
  `// 统计：${cols} × ${rows} 格，共 ${blocks.length} 个图元（房间 ${blocks.filter((b) => b.kind === 'room').length} 间）`,
  '',
  `export const STANDARD_FLOOR_COLS = ${cols}`,
  `export const STANDARD_FLOOR_ROWS = ${rows}`,
  `export const STANDARD_FLOOR_JSON = ${JSON.stringify(payload)}`,
  '',
].join('\n')

for (const app of ['frontend/pc-admin', 'frontend/mobile-web']) {
  const out = `${app}/src/utils/standardFloor.ts`
  writeFileSync(out, content, 'utf8')
  console.log('written', out, content.length, 'bytes')
}
console.log(`blocks=${blocks.length} rooms=${blocks.filter((b) => b.kind === 'room').length} cols=${cols} rows=${rows}`)
