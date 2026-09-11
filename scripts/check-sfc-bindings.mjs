#!/usr/bin/env node
// 前端 SFC 静态自检：找出「模板里用到、但 <script setup> 里没有声明」的标识符。
// vite build 不做类型/绑定检查，这类问题只会在浏览器运行时报 ReferenceError，
// 所以改动 .vue 后用本脚本兜底检查。
// 用法：node scripts/check-sfc-bindings.mjs frontend/pc-admin [frontend/mobile-web ...]

import { readFileSync, readdirSync, statSync } from 'node:fs'
import { join, relative } from 'node:path'
import { createRequire } from 'node:module'

const GLOBALS = new Set([
  'Math', 'JSON', 'Date', 'Number', 'String', 'Boolean', 'Array', 'Object', 'RegExp', 'Map', 'Set',
  'Promise', 'Error', 'Intl', 'Symbol', 'BigInt', 'NaN', 'Infinity', 'undefined', 'null', 'true', 'false',
  'parseInt', 'parseFloat', 'isNaN', 'isFinite', 'encodeURIComponent', 'decodeURIComponent',
  'window', 'document', 'localStorage', 'sessionStorage', 'console', 'navigator', 'location',
  'setTimeout', 'clearTimeout', 'setInterval', 'clearInterval', 'requestAnimationFrame',
  'WebSocket', 'Event', 'URL', 'Blob', 'FormData', 'HTMLElement', 'IntersectionObserver', 'ResizeObserver',
  'this', 'new', 'typeof', 'instanceof', 'in', 'of', 'return', 'await', 'async', 'function', 'if', 'else',
  'delete', 'void', 'throw', 'case', 'switch', 'do', 'while', 'for', 'const', 'let', 'var', 'class',
  '$event', '$slots', '$attrs', '$props', '$emit', '$refs', '$nextTick', '$forceUpdate',
])

function walk(dir, out = []) {
  for (const name of readdirSync(dir)) {
    const full = join(dir, name)
    const st = statSync(full)
    if (st.isDirectory()) walk(full, out)
    else if (name.endsWith('.vue')) out.push(full)
  }
  return out
}

// 去掉字符串字面量，避免把文案里的词当成标识符
function stripLiterals(expr) {
  return expr.replace(/'[^']*'/g, "''").replace(/"[^"]*"/g, '""').replace(/`[^`]*`/g, '``')
}

function collectLocals(template) {
  const locals = new Set()
  // v-for="(a, b) in list" / v-for="a in list" / v-for="a of list"
  const forRe = /v-for\s*=\s*"([^"]*)"/g
  let m
  while ((m = forRe.exec(template))) {
    const head = m[1].split(/\s+in\s+|\s+of\s+/)[0] || ''
    for (const token of head.replace(/[(){}[\]]/g, ' ').split(/[\s,]+/)) {
      const name = token.split(':').pop().trim()
      if (/^[A-Za-z_$][\w$]*$/.test(name)) locals.add(name)
    }
  }
  // v-slot / #default="{ row }"
  const slotRe = /(?:v-slot(?::[\w-]+)?|#[\w-]*)\s*=\s*"([^"]*)"/g
  while ((m = slotRe.exec(template))) {
    for (const token of m[1].replace(/[{}]/g, ' ').split(/[\s,]+/)) {
      const name = token.trim()
      if (/^[A-Za-z_$][\w$]*$/.test(name)) locals.add(name)
    }
  }
  return locals
}

function expressionsOf(template) {
  const exprs = []
  let m
  const mustache = /\{\{([\s\S]*?)\}\}/g
  while ((m = mustache.exec(template))) exprs.push(m[1])
  const tagRe = /<[A-Za-z][^>]*>/g
  while ((m = tagRe.exec(template))) {
    const tag = m[0]
    const attrRe = /(:|v-bind:|@|v-on:|v-|#)([^\s=/>]*)\s*=\s*"([^"]*)"/g
    let a
    while ((a = attrRe.exec(tag))) exprs.push(a[3])
  }
  return exprs
}

function identifiers(expr) {
  const cleaned = stripLiterals(expr)
  const ids = []
  const re = /[A-Za-z_$][\w$]*/g
  let m
  while ((m = re.exec(cleaned))) {
    const name = m[0]
    const before = cleaned.slice(0, m.index).trimEnd()
    if (before.endsWith('.')) continue            // 属性访问 obj.prop
    const after = cleaned.slice(m.index + name.length)
    if (/^\s*:/.test(after) && /[{,[]\s*$/.test(before)) continue // 对象字面量的 key
    if (GLOBALS.has(name)) continue
    ids.push(name)
  }
  return ids
}

// 检查 <script setup> 里使用了 x.value、但 x 从未声明的情况（vite build 不报错，运行时会 ReferenceError）
function scriptRefIssues(source) {
  const cleaned = source.replace(/\/\*[\s\S]*?\*\//g, '').replace(/\/\/[^\n]*/g, '')
  const declared = new Set()
  const add = (name) => {
    if (name && /^[A-Za-z_$][\w$]*$/.test(name)) declared.add(name)
  }
  let m
  const declRe = /(?:const|let|var)\s+([A-Za-z_$][\w$]*)/g
  while ((m = declRe.exec(cleaned))) add(m[1])
  const destrRe = /(?:const|let|var)\s*[{[][^}\]]*[}\]]/g
  while ((m = destrRe.exec(cleaned))) {
    const inner = m[0].replace(/^[^{[]*[{[]/, '').replace(/[}\]]$/, '')
    for (const part of inner.split(',')) add(part.split(':').pop().trim())
  }
  const importRe = /import\s+([^'"]+?)\s+from/g
  while ((m = importRe.exec(cleaned))) {
    for (const part of m[1].replace(/[{}*]/g, ' ').split(',')) add(part.split(/\s+as\s+/).pop().trim())
  }
  const paramRe = /(?:function\s+[A-Za-z_$][\w$]*\s*\(|\(|=>\s*\()([^)]*)\)/g
  while ((m = paramRe.exec(cleaned))) {
    for (const part of m[1].split(',')) add(part.split(':')[0].trim())
  }
  const missing = new Set()
  const useRe = /([A-Za-z_$][\w$]*)\.value\b/g
  while ((m = useRe.exec(cleaned))) if (!declared.has(m[1])) missing.add(m[1])
  return [...missing]
}
let problems = 0
for (const dir of process.argv.slice(2)) {
  const require = createRequire(join(process.cwd(), dir, 'package.json'))
  const { parse, compileScript } = require('@vue/compiler-sfc')
  for (const file of walk(join(dir, 'src'))) {
    const source = readFileSync(file, 'utf8')
    const { descriptor, errors } = parse(source, { filename: file })
    if (errors.length || !descriptor.scriptSetup || !descriptor.template) continue
    let bindings
    try {
      bindings = Object.keys(compileScript(descriptor, { id: 'check' }).bindings || {})
    } catch (err) {
      console.log(`SKIP ${relative(process.cwd(), file)} (${err.message})`)
      continue
    }
    const known = new Set(bindings)
    for (const local of collectLocals(descriptor.template.content)) known.add(local)
    const seen = new Map()
    for (const expr of expressionsOf(descriptor.template.content)) {
      for (const id of identifiers(expr)) {
        if (known.has(id)) continue
        if (!seen.has(id)) seen.set(id, expr.replace(/\s+/g, ' ').trim().slice(0, 70))
      }
    }
    if (seen.size) {
      problems += seen.size
      console.log(`\n${relative(process.cwd(), file)}`)
      for (const [id, expr] of seen) console.log(`  ✗ ${id}    ← ${expr}`)
    }
    const refIssues = scriptRefIssues(descriptor.scriptSetup.content)
    if (refIssues.length) {
      problems += refIssues.length
      console.log(`\n${relative(process.cwd(), file)} (script)`)
      for (const id of refIssues) console.log(`  x ${id}.value 使用了但未声明`)
    }
  }
}
console.log(problems ? `\n发现 ${problems} 处可疑引用（请人工确认）` : '\n未发现未声明的模板引用 ✅')
